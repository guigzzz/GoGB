package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"runtime/pprof"

	"github.com/guigzzz/GoGB/backend"
)

func main() {

	debug := flag.Bool("debug", false, "run the emulator in debug mode")
	profile := flag.Bool("profile", false, "profile the emulator")
	flag.Parse()

	if *profile {
		f, err := os.Create("emulator.prof")
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	if len(flag.Args()) != 1 {
		fmt.Printf("Usage: ./%s <path to rom>\n", path.Base(os.Args[0]))
		os.Exit(0)
	}

	romPath := flag.Arg(0)

	var cpu *backend.CPU
	var ppu *backend.PPU
	if backend.SaveExistsForRom(romPath) {
		ppu, cpu = backend.LoadSave(romPath)
	} else {
		rom, err := ioutil.ReadFile(romPath)
		if err != nil {
			panic(err)
		}

		cpu = backend.NewCPU(rom, *debug, nil, nil)
		ppu = backend.NewPPU(cpu)
	}

	RunGame(ppu, cpu)

	defer backend.DumpEmulatorState(romPath, ppu, cpu)
}
