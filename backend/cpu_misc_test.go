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

func TestRet(t *testing.T) {
	c := NewTestCPU()

	c.PC = 0x100
	c.pushPC()

	c.PC = 0x110
	c.RetZ()
	assert.Equal(t, c.PC, uint16(0x111))
	c.RetNZ()
	assert.Equal(t, c.PC, uint16(0x100))

	c.SetFlag(CFlag)
	c.PC = 0x200
	c.pushPC()

	c.PC = 0x300
	c.RetNC()
	assert.Equal(t, c.PC, uint16(0x301))
	c.RetC()
	assert.Equal(t, c.PC, uint16(0x200))
}

func TestCall(t *testing.T) {
	c := NewTestCPU()
	c.SP = 0xFFFE
	c.PC = 0x100

	c.Call(0x200)
	assert.Equal(t, c.PC, uint16(0x200))
	assert.Equal(t, c.ram[c.SP], byte(0x03))
	assert.Equal(t, c.ram[c.SP+1], byte(0x1))

	c.SetFlag(ZFlag)
	c.CallNZ(0x300)
	assert.Equal(t, c.PC, uint16(0x203))
	c.CallC(0x400)
	assert.Equal(t, c.PC, uint16(0x206))
	c.CallZ(0x300)
	assert.Equal(t, c.PC, uint16(0x300))
	assert.Equal(t, c.ram[c.SP], byte(0x09))
	assert.Equal(t, c.ram[c.SP+1], byte(0x2))
}
