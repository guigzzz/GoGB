package backend

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTimerFiresWhenItShould(t *testing.T) {

	c := NewTestCPU()

	c.cycleCounter = 1024

	tma := byte(0x5)

	c.ram[0xFF07] = 0b100 // TAC
	c.ram[0xFF06] = tma   // TMA
	c.ram[0xFF05] = 0xFF  // TIMA

	c.checkForTimerIncrementAndInterrupt()

	assert.Equal(t, tma, c.ram[0xFF05])
	assert.Equal(t, byte(0x4), c.ram[0xFF0F])
}

func TestTimerDoesNotFireWhenTimerOff(t *testing.T) {

	c := NewTestCPU()

	c.cycleCounter = 1024

	tma := byte(0x5)

	c.ram[0xFF07] = 0b0  // TAC
	c.ram[0xFF06] = tma  // TMA
	c.ram[0xFF05] = 0xFF // TIMA

	c.checkForTimerIncrementAndInterrupt()

	assert.Equal(t, byte(0xFF), c.ram[0xFF05])
	assert.Equal(t, byte(0), c.ram[0xFF0F])
}

func TestTimerDoesNotFireWhenItShouldNot(t *testing.T) {

	c := NewTestCPU()

	c.cycleCounter = 1020

	tma := byte(0x5)

	c.ram[0xFF07] = 0b100 // TAC
	c.ram[0xFF06] = tma   // TMA
	c.ram[0xFF05] = 0xFF  // TIMA

	c.checkForTimerIncrementAndInterrupt()

	assert.Equal(t, byte(0xFF), c.ram[0xFF05])
	assert.Equal(t, byte(0), c.ram[0xFF0F])
}
