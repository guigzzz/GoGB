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

	pcInc, cycleInc := c.JumpNZ(target)
	assert.Equal(t, pcInc, 3)
	assert.Equal(t, cycleInc, 12)

	c.PC = 0
	pcInc, cycleInc = c.JumpZ(target)
	assert.Equal(t, pcInc, 0)
	assert.Equal(t, cycleInc, 16)
	assert.Equal(t, c.PC, uint16(0x8000))

	c.PC = 0
	pcInc, cycleInc = c.JumpC(target)
	assert.Equal(t, pcInc, 3)
	assert.Equal(t, cycleInc, 12)

	c.PC = 0
	pcInc, cycleInc = c.JumpNC(target)
	assert.Equal(t, pcInc, 0)
	assert.Equal(t, cycleInc, 16)
	assert.Equal(t, c.PC, uint16(0x8000))
}

func TestStack(t *testing.T) {
	c := NewTestCPU()
	c.SP = 0xFFFE
	c.PC = 0x0110

	c.pushPC()
	assert.Equal(t, c.readMemory(0xFFFD), byte(0x01))
	assert.Equal(t, c.readMemory(0xFFFC), byte(0x10))

	c.reg[B] = 0xFF
	c.reg[C] = 0x2
	c.pushDouble(B, C)
	c.popDouble(H, L)

	assert.Equal(t, c.reg[H], byte(0xFF))
	assert.Equal(t, c.reg[L], byte(0x2))

	assert.Equal(t, c.popPC(), uint16(0x0110))
}

func TestRet(t *testing.T) {
	c := NewTestCPU()

	c.PC = 0x100
	c.pushPC()

	c.PC = 0x110
	pcInc, cycleInc := c.RetZ()
	assert.Equal(t, pcInc, 1)
	assert.Equal(t, cycleInc, 8)

	pcInc, cycleInc = c.RetNZ()
	assert.Equal(t, pcInc, 0)
	assert.Equal(t, cycleInc, 20)

	c.SetFlag(CFlag)
	c.PC = 0x200
	c.pushPC()

	c.PC = 0x300
	pcInc, cycleInc = c.RetNC()
	assert.Equal(t, pcInc, 1)
	assert.Equal(t, cycleInc, 8)

	pcInc, cycleInc = c.RetC()
	assert.Equal(t, pcInc, 0)
	assert.Equal(t, cycleInc, 20)
	assert.Equal(t, c.PC, uint16(0x200))
}

func TestCall(t *testing.T) {
	c := NewTestCPU()
	c.SP = 0xFFFE
	c.PC = 0x100

	c.Call(0x200)
	assert.Equal(t, c.PC, uint16(0x200))
	assert.Equal(t, c.readMemory(c.SP), byte(0x03))
	assert.Equal(t, c.readMemory(c.SP+1), byte(0x1))

	c.SetFlag(ZFlag)
	pcInc, cycleInc := c.CallNZ(0x300)
	assert.Equal(t, 3, pcInc)
	assert.Equal(t, 12, cycleInc)

	pcInc, cycleInc = c.CallC(0x400)
	assert.Equal(t, 3, pcInc)
	assert.Equal(t, 12, cycleInc)

	pcInc, cycleInc = c.CallZ(0x300)
	assert.Equal(t, 0, pcInc)
	assert.Equal(t, 24, cycleInc)

	assert.Equal(t, uint16(0x300), c.PC)
	assert.Equal(t, byte(0x03), c.readMemory(c.SP))
	assert.Equal(t, byte(0x2), c.readMemory(c.SP+1))
}
