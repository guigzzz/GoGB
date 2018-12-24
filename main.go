package main

import (
	"runtime"

	"github.com/guigzzz/GoGB/backend"
)

func init() {
	runtime.LockOSThread()
}

func main() {

	cpu := backend.NewCPU()
	ppu := backend.NewPPU(cpu)
	go ppu.Renderer()

	// screenRenderer := NewScreenRenderer(ppu, 160, 144)
	// screenRenderer.startRendering()

	debug := backend.NewDebugHarness()

	for cpu.PC < 0x00fe {
		cpu.DecodeAndExecuteNext()

		if cpu.PC >= 0x00e6 {
			debug.PrintDebug(cpu)
		}
	}
}
