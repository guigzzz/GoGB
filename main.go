package main

import (
	"fmt"
	"io/ioutil"
	"runtime"

	"github.com/guigzzz/GoGB/backend"
)

func init() {
	runtime.LockOSThread()
}

func main() {

	cpu := backend.NewHLECPU()

	data, err := ioutil.ReadFile("rom/cpu_instrs/individual/09-op r,r.gb")
	if err != nil {
		panic(err)
	}
	cpu.LoadToRAM(data)

	ppu := backend.NewPPU(cpu)
	go ppu.Renderer()

	screenRenderer := NewScreenRenderer(ppu, 200, 200)

	debug := backend.NewDebugHarness()

	go func() {
		for {
			// debug.PrintDebug(cpu)
			cpu.DecodeAndExecuteNext()

			if cpu.PC < 0x0100 {
				fmt.Println("PC < 0x0100. Test failed?")
				break
			}
		}
		debug.PrintDebug(cpu)
	}()
	screenRenderer.startRendering()
}
