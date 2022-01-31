package backend

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSaveEmulatorState(t *testing.T) {

	rom, err := ioutil.ReadFile(blargg)
	if err != nil {
		panic(err)
	}

	ppu, cpu, mmu, _ := composeForTests(rom)

	for i := 0; i < 5000; i++ {
		ppu.RunEmulatorForAFrame()
	}

	ref := getImage("ref/blargg.png")
	if !assert.Equal(t, ref, ppu.Image) {
		ppu.dumpScreenToPng("out/blargg.png")
	}

	DumpEmulatorState(blargg, ppu, cpu, mmu)

	ppu, _, apu, _ := LoadSave(blargg)

	apu.Disable()

	for i := 0; i < 100; i++ {
		ppu.RunEmulatorForAFrame()
	}

	if !assert.Equal(t, ref, ppu.Image) {
		ppu.dumpScreenToPng("out/blargg.png")
	}
}
