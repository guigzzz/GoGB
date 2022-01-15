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

	cpu := NewCPU(rom, false, NewNullLogger(), NullApuFactory)
	ppu := NewPPU(cpu)

	for i := 0; i < 5000; i++ {
		ppu.RunEmulatorForAFrame()
	}

	ref := getImage("ref/blargg.png")
	if !assert.Equal(t, ref, ppu.Image) {
		ppu.dumpScreenToPng("out/blargg.png")
	}

	DumpEmulatorState(blargg, ppu, cpu)

	ppu, _ = LoadSave(blargg, NullApuFactory)

	ppu.RunEmulatorForAFrame()

	if !assert.Equal(t, ref, ppu.Image) {
		ppu.dumpScreenToPng("out/blargg.png")
	}
}
