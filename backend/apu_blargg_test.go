package backend

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

type AudioTest struct {
	name                  string
	successProgramCounter uint16
}

func TestRunBlarggAudioTestRoms(t *testing.T) {

	roms := []AudioTest{
		{name: "01-registers.gb", successProgramCounter: 0xCAA2},
		{name: "02-len ctr.gb", successProgramCounter: 0xcef7},
	}

	for _, r := range roms {
		fullPath := "../rom/sound_rom_singles/" + r.name
		t.Run(r.name, func(t *testing.T) {
			rom, err := ioutil.ReadFile(fullPath)
			if err != nil {
				panic(err)
			}

			cpu := NewCPU(rom, false, NewNullLogger(), nil)

			// setup reading so that the APU doesn't block
			buf := make([]byte, 100000)
			go (func() {
				for {
					cpu.apu.ToReadCloser().Read(buf)
				}
			})()

			ppu := NewPPU(cpu)

			for cpu.PC != r.successProgramCounter && cpu.cycleCounter < 50_000_000 {
				ppu.RunEmulatorForAFrame()
			}

			assert.Equal(t, r.successProgramCounter, cpu.PC)
		})
	}

}
