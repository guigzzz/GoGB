package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"runtime"

	"github.com/guigzzz/GoGB/backend"
)

func init() {
	runtime.LockOSThread()
}

func main() {

	debug := flag.Bool("debug", false, "run the emulator in debug mode")
	flag.Parse()

	if len(flag.Args()) != 1 {
		fmt.Println(fmt.Sprintf("Usage: ./%s <path to rom>", path.Base(os.Args[0])))
		os.Exit(0)
	}

	rom, err := ioutil.ReadFile(flag.Arg(0))
	if err != nil {
		panic(err)
	}

	cpu := backend.NewCPU(rom, *debug)
	ppu := backend.NewPPU(cpu)
	screenRenderer := NewScreenRenderer(ppu, cpu, 175, 155)

	go ppu.Renderer()
	screenRenderer.startRendering()
}
