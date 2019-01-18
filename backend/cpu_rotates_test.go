package backend

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRotateLeftCarryReg(t *testing.T) {
	c := NewTestCPU()
	c.reg[B] = 0x85
	c.ram[0] = 0

	c.RotateLeftCReg(B)
	assert.Equal(t, c.reg[B], byte(0x0B))
	assertFlagsSet(t, c.reg[F], CFlag)
}

func TestRotateLeftCarryHL(t *testing.T) {
	c := NewTestCPU()
	c.reg[B] = 0x85

	c.Writedouble(H, L, 0xC000)
	c.writeMemory(c.ReadHL(), 0)

	c.RotateLeftCHL()
	assert.Equal(t, c.ram[c.ReadHL()], byte(0))
	assertFlagsSet(t, c.reg[F], ZFlag)
}

func TestRotateLeftReg(t *testing.T) {
	c := NewTestCPU()
	c.reg[L] = 0x80
	c.ram[0x0080] = 0x11

	c.RotateLeftReg(L)
	assert.Equal(t, c.reg[L], byte(0))
	assertFlagsSet(t, c.reg[F], CFlag, ZFlag)
}

func TestRotateLeftHL(t *testing.T) {
	c := NewTestCPU()
	c.reg[L] = 0x80
	c.Writedouble(H, L, 0xC000)
	c.writeMemory(c.ReadHL(), 0x11)

	c.RotateLeftHL()
	assert.Equal(t, c.ram[c.ReadHL()], byte(0x22))
	assertFlagsSet(t, c.reg[F])
}

func TestRotateRightCarryReg(t *testing.T) {
	c := NewTestCPU()
	c.ram[0] = 0
	c.reg[C] = 0x1

	c.RotateRightCReg(C)
	assert.Equal(t, c.reg[C], byte(0x80))
	assertFlagsSet(t, c.reg[F], CFlag)
}

func TestRotateRightCarryHL(t *testing.T) {
	c := NewTestCPU()
	c.Writedouble(H, L, 0xC000)
	c.writeMemory(c.ReadHL(), 0)
	c.reg[C] = 0x1

	c.RotateRightCHL()
	assert.Equal(t, c.ram[c.ReadHL()], byte(0))
	assertFlagsSet(t, c.reg[F], ZFlag)
}

func TestRotateRightReg(t *testing.T) {
	c := NewTestCPU()
	c.ram[0] = 0x8A
	c.reg[A] = 0x01

	c.RotateRightReg(A)
	assert.Equal(t, c.reg[A], byte(0))
	assertFlagsSet(t, c.reg[F], CFlag, ZFlag)
}

func TestRotateRightHL(t *testing.T) {
	c := NewTestCPU()
	c.ram[0] = 0x8A
	c.Writedouble(H, L, 0xC000)
	c.writeMemory(c.ReadHL(), 0x8A)
	c.reg[A] = 0x01

	c.RotateRightHL()
	assert.Equal(t, c.ram[c.ReadHL()], byte(0x45))
	assertFlagsSet(t, c.reg[F])
}
