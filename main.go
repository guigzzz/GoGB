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

var files = []string{
	"rom/cpu_instrs/individual/01-special.gb",            // 0 FAIL #6
	"rom/cpu_instrs/individual/02-interrupts.gb",         // 1 CRASH
	"rom/cpu_instrs/individual/03-op sp,hl.gb",           // 2 HANG/INF LOOP
	"rom/cpu_instrs/individual/04-op r,imm.gb",           // 3 FAIL
	"rom/cpu_instrs/individual/05-op rp.gb",              // 4 PASS
	"rom/cpu_instrs/individual/06-ld r,r.gb",             // 5 PASS
	"rom/cpu_instrs/individual/07-jr,jp,call,ret,rst.gb", // 6 CRASH
	"rom/cpu_instrs/individual/08-misc instrs.gb",        // 7 Goes to Noop
	"rom/cpu_instrs/individual/09-op r,r.gb",             // 8 FAIL
	"rom/cpu_instrs/individual/10-bit ops.gb",            // 9 PASS
	"rom/cpu_instrs/individual/11-op a,(hl).gb",          // 10 CRASH
}

func main() {

	cpu := backend.NewHLECPU()

	data, err := ioutil.ReadFile(files[8])
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
