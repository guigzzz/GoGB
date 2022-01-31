package backend

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func InitApu() *APU {
	ram := make([]byte, 1<<16)

	apu := NewAPU(ram)
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

func TestGetWaveOutput(t *testing.T) {
	apu := InitApu()

	apu.positionCounterWave = 0

	apu.ram[NR32] = 0b0010_0000
	apu.ram[0xFF30] = 0xC0

	apu.ram[NR52] = 0xF
	apu.ram[NR51] = 0xF // only enable right channel

	left, right := apu.getWaveOutput()
	assert.Equal(t, byte(0), left)
	assert.Equal(t, byte(0xC), right)
}

func TestUpdateWaveFrequencyTimerTick(t *testing.T) {

	apu := InitApu()

	apu.ram[NR33] = 0x11
	apu.ram[NR34] = 0x1
	apu.positionCounterWave = 31

	apu.updateState()

	assert.Equal(t, 3550, apu.frequencyTimerWave)
	assert.Equal(t, 0, apu.positionCounterWave)
}

func TestUpdateWaveFrequencyTimerNoTick(t *testing.T) {

	apu := InitApu()

	apu.ram[NR33] = 0x11
	apu.ram[NR34] = 0x1

	apu.frequencyTimerWave = 1
	apu.positionCounterWave = 31

	apu.updateState()

	assert.Equal(t, 0, apu.frequencyTimerWave)
	assert.Equal(t, 31, apu.positionCounterWave)
}
