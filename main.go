package main

import (
	"io/ioutil"
	"runtime"

	"github.com/guigzzz/GoGB/backend"
)

func init() {
	runtime.LockOSThread()
}

var files = []string{
	"rom/cpu_instrs/individual/01-special.gb",            // 0 PASS
	"rom/cpu_instrs/individual/02-interrupts.gb",         // 1 FAIL #2
	"rom/cpu_instrs/individual/03-op sp,hl.gb",           // 2 PASS
	"rom/cpu_instrs/individual/04-op r,imm.gb",           // 3 PASS
	"rom/cpu_instrs/individual/05-op rp.gb",              // 4 PASS
	"rom/cpu_instrs/individual/06-ld r,r.gb",             // 5 PASS
	"rom/cpu_instrs/individual/07-jr,jp,call,ret,rst.gb", // 6 CRASH
	"rom/cpu_instrs/individual/08-misc instrs.gb",        // 7 PASS
	"rom/cpu_instrs/individual/09-op r,r.gb",             // 8 PASS
	"rom/cpu_instrs/individual/10-bit ops.gb",            // 9 PASS
	"rom/cpu_instrs/individual/11-op a,(hl).gb",          // 10 PASS
}

func main() {

	cpu := backend.NewHLECPU()

	data, err := ioutil.ReadFile(files[6])
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
			debug.PrintDebugShort(cpu)
			// debug.PrintDebug(cpu)
			// debug.RecordNextExercisedOp(cpu)
			cpu.DecodeAndExecuteNext()

			if cpu.PC == 0xCC62 {
				break
			}
		}
		// debug.PrintDebug(cpu)
		debug.GetExercicedOpSummary()
	}()
	screenRenderer.startRendering()

	// debug.GetExercicedOpSummary()
}
