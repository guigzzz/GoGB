package backend

import (
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
		{name: "03-trigger.gb", successProgramCounter: 0xc74d},
		{name: "04-sweep.gb", successProgramCounter: 0xcc63},
		{name: "06-overflow on trigger.gb", successProgramCounter: 0xc93e},
	}

	for _, r := range roms {
		fullPath := "../rom/sound_rom_singles/" + r.name
		t.Run(r.name, func(t *testing.T) {
			emulator := NewEmulator(fullPath, WithDisableApu())

			cpu := emulator.cpu

			for cpu.PC != r.successProgramCounter && cpu.cycleCounter < 50_000_000 {
				emulator.RunForAFrame()
			}

			assert.Equal(t, r.successProgramCounter, cpu.PC)
		})
	}

}
