package backend

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSaveEmulatorState(t *testing.T) {
	emulator := NewEmulator(blargg, WithDisableApu())

	for i := 0; i < 5000; i++ {
		emulator.RunForAFrame()
	}

	ref := getImage("ref/blargg.png")
	if !assert.Equal(t, ref, emulator.ppu.Image) {
		emulator.ppu.dumpScreenToPng("out/blargg.png")
	}

	DumpEmulatorState(blargg, emulator)

	emulator = LoadSave(blargg)

	emulator.apu.Disable()

	for i := 0; i < 100; i++ {
		emulator.RunForAFrame()
	}

	if !assert.Equal(t, ref, emulator.ppu.Image) {
		emulator.ppu.dumpScreenToPng("out/blargg.png")
	}
}
