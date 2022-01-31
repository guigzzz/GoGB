package backend

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTimerFiresWhenItShould(t *testing.T) {

	c := NewTestCPU()

	c.cycleCounter = 1024

	tma := byte(0x5)

	c.writeMemory(0xFF07, 0b100) // TAC
	c.writeMemory(0xFF06, tma)   // TMA
	c.writeMemory(0xFF05, 0xFF)  // TIMA

	c.checkForTimerIncrementAndInterrupt()

	assert.Equal(t, tma, c.readMemory(0xFF05))
	assert.Equal(t, byte(0x4), c.readMemory(0xFF0F))
}

func TestTimerDoesNotFireWhenTimerOff(t *testing.T) {

	c := NewTestCPU()

	c.cycleCounter = 1024

	tma := byte(0x5)

	c.writeMemory(0xFF07, 0b0)  // TAC
	c.writeMemory(0xFF06, tma)  // TMA
	c.writeMemory(0xFF05, 0xFF) // TIMA

	c.checkForTimerIncrementAndInterrupt()

	assert.Equal(t, byte(0xFF), c.readMemory(0xFF05))
	assert.Equal(t, byte(0), c.readMemory(0xFF0F))
}

func TestTimerDoesNotFireWhenItShouldNot(t *testing.T) {

	c := NewTestCPU()

	c.cycleCounter = 1020

	tma := byte(0x5)

	c.writeMemory(0xFF07, 0b100) // TAC
	c.writeMemory(0xFF06, tma)   // TMA
	c.writeMemory(0xFF05, 0xFF)  // TIMA

	c.checkForTimerIncrementAndInterrupt()

	assert.Equal(t, byte(0xFF), c.readMemory(0xFF05))
	assert.Equal(t, byte(0), c.readMemory(0xFF0F))
}
