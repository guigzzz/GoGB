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

	cpu := backend.NewCPU()

	data, err := ioutil.ReadFile("rom/cpu_instrs/individual/09-op r,r.gb")
	if err != nil {
		panic(err)
	}

	fmt.Println(len(data))

	cpu.LoadToRAM(data)
	cpu.PC = 0x0100

	ppu := backend.NewPPU(cpu)
	go ppu.Renderer()

	screenRenderer := NewScreenRenderer(ppu, 200, 200)

	// debug := backend.NewDebugHarness()

	go func() {
		for {
			cpu.DecodeAndExecuteNext()
		}
	}()
	screenRenderer.startRendering()
}
