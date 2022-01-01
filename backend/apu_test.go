package backend

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func InitApu() *APUImpl {
	cpu := NewTestCPU()
	apu := NewAPU(cpu)
	apu.emitSamples = false

	return apu
}

func TestGetNoiseOutput(t *testing.T) {
	apu := InitApu()

	apu.lsfr = 0
	apu.currentVolumeNoise = 0xF

	apu.ram[NR52] = 0xF
	apu.ram[NR51] = 0xF // only enable right channel

	left, right := apu.getNoiseOutput()
	assert.Equal(t, byte(0), left)
	assert.Equal(t, byte(0xF), right)
}

func TestUpdateNoiseFrequencyTimerNoUpdate(t *testing.T) {
	apu := InitApu()

	apu.ram[NR43] = 0b1111_0111
	apu.lsfr = 0x7FFF

	apu.frequencyTimerNoise = 1

	apu.updateState()

	assert.Equal(t, 0, apu.frequencyTimerNoise)
	assert.Equal(t, uint16(0x7FFF), apu.lsfr)
}

func TestUpdateNoiseFrequencyTimerNoWidth(t *testing.T) {
	apu := InitApu()

	apu.ram[NR43] = 0b1111_0111
	apu.lsfr = 0x7FFF

	apu.updateState()

	expectedFrequencyTimer := 112 << 15
	assert.Equal(t, expectedFrequencyTimer, apu.frequencyTimerNoise)

	assert.Equal(t, uint16(0x3FFF), apu.lsfr)
}

func TestUpdateNoiseFrequencyTimerWidth(t *testing.T) {
	apu := InitApu()

	apu.ram[NR43] = 0b1111_1111
	apu.lsfr = 0x7FFF

	apu.updateState()

	expectedFrequencyTimer := 112 << 15
	assert.Equal(t, expectedFrequencyTimer, apu.frequencyTimerNoise)

	assert.Equal(t, uint16(0x3FBF), apu.lsfr)
}
