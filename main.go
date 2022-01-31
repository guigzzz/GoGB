package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"runtime/pprof"

	"github.com/guigzzz/GoGB/backend"
)

func main() {

	debug := flag.Bool("debug", false, "run the emulator in debug mode")
	profile := flag.Bool("profile", false, "profile the emulator")
	loadSave := flag.Bool("load-save", false, "try to load a save")
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

	var emu *backend.Emulator

	if *loadSave && backend.SaveExistsForRom(romPath) {
		emu = backend.LoadSave(romPath)
	} else {
		emu = backend.NewEmulator(romPath, *debug, true)
	}

	if *loadSave {
		defer backend.DumpEmulatorState(romPath, emu)
	}

	RunGame(emu)
}
