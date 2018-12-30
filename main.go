package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"runtime"

	"github.com/guigzzz/GoGB/backend"
)

func init() {
	runtime.LockOSThread()
}

func main() {

	if len(os.Args) != 2 {
		fmt.Println(fmt.Sprintf("Usage: ./%s <path to rom>", os.Args[0]))
		os.Exit(0)
	}

	rom, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}

	cpu := backend.NewCPU(rom)
	ppu := backend.NewPPU(cpu)
	screenRenderer := NewScreenRenderer(ppu, cpu, 175, 155)

	go ppu.Renderer()
	go cpu.Runner()
	screenRenderer.startRendering()
}
