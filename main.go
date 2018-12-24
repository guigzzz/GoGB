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

	// cpu.GetRAM()[0xFF44] = 144

	screenRenderer := NewScreenRenderer(ppu, 160, 144)
	screenRenderer.startRendering()

	// debug := backend.NewDebugHarness()

	for {
		cpu.DecodeAndExecuteNext()

		// if cpu.PC >= 0x006a {
		// 	debug.PrintDebug(cpu)
		// }
	}
}
