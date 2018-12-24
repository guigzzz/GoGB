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
	go cpu.Run()
	ppu := backend.NewPPU(cpu)
	go ppu.Renderer()

	screenRenderer := NewScreenRenderer(ppu, 160, 144)
	screenRenderer.startRendering()
}
