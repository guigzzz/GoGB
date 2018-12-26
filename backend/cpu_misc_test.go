package backend

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJumps(t *testing.T) {
	c := NewTestCPU()
	c.SetFlag(ZFlag)
	c.ResetFlag(CFlag)

	target := uint16(0x8000)
	c.PC = 0
	c.JumpNZ(target)
	assert.Equal(t, c.PC, uint16(3))

	c.PC = 0
	c.JumpZ(target)
	assert.Equal(t, c.PC, uint16(0x8000))

	c.PC = 0
	c.JumpC(target)
	assert.Equal(t, c.PC, uint16(3))

	c.PC = 0
	c.JumpNC(target)
	assert.Equal(t, c.PC, uint16(0x8000))
}

func TestStack(t *testing.T) {
	c := NewTestCPU()
	c.SP = 0xFFFE
	c.PC = 0x0110

	c.pushPC()
	assert.Equal(t, c.ram[0xFFFD], byte(0x01))
	assert.Equal(t, c.ram[0xFFFC], byte(0x10))

	c.reg[B] = 0xFF
	c.reg[C] = 0x2
	c.pushDouble(B, C)
	c.popDouble(H, L)

	assert.Equal(t, c.reg[H], byte(0xFF))
	assert.Equal(t, c.reg[L], byte(0x2))

	assert.Equal(t, c.popPC(), uint16(0x0110))
}
