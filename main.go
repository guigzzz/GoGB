package main

import (
	"fmt"
	"runtime"
	"time"

	"github.com/guigzzz/GoGB/backend"
)

func init() {
	runtime.LockOSThread()
}

func main() {

	cpu := backend.NewCPU()
	ppu := backend.NewPPU(cpu)
	go ppu.Renderer()

	screenRenderer := NewScreenRenderer(ppu, 200, 200)

	go func() {
		fmt.Println("[CPU] Booting...")
		for cpu.PC != 0x0100 {
			time.Sleep(2)
			cpu.DecodeAndExecuteNext()
		}
		fmt.Println("cpu booted")
	}()
	screenRenderer.startRendering()

	// f, err := os.Create("img.jpg")
	// if err != nil {
	// 	panic(err)
	// }
	// defer f.Close()
	// jpeg.Encode(f, ppu.Image, nil)
}
