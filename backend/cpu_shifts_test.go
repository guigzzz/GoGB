package backend

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShiftLeftArithmeticReg(t *testing.T) {
	c := NewTestCPU()
	c.reg[D] = 0x80
	c.ram[0] = 0xFF

	c.ShiftLeftArithmeticReg(D)
	assert.Equal(t, c.reg[D], byte(0))
	assertFlagsSet(t, c.reg[F], CFlag, ZFlag)
}

func TestShiftLeftArithmeticHL(t *testing.T) {
	c := NewTestCPU()
	c.reg[D] = 0x80
	c.ram[0] = 0xFF

	c.ShiftLeftArithmeticHL()
	assert.Equal(t, c.ram[c.ReadHL()], byte(0xFE))
	assertFlagsSet(t, c.reg[F], CFlag)
}

func TestShiftRightArithmeticReg(t *testing.T) {
	c := NewTestCPU()
	c.reg[A] = 0x8A
	c.ram[0] = 0x01

	c.ShiftRightArithmeticReg(A)
	assert.Equal(t, c.reg[A], byte(0xC5))
	assertFlagsSet(t, c.reg[F])
}

func TestShiftRightArithmeticHL(t *testing.T) {
	c := NewTestCPU()
	c.reg[A] = 0x8A
	c.ram[0] = 0x01

	c.ShiftRightArithmeticHL()
	assert.Equal(t, c.ram[c.ReadHL()], byte(0))
	assertFlagsSet(t, c.reg[F], CFlag, ZFlag)
}

func TestShiftRightLogicalReg(t *testing.T) {
	c := NewTestCPU()
	c.reg[A] = 0x01
	c.ram[0] = 0xFF

	c.ShiftRightLogicalReg(A)
	assert.Equal(t, c.reg[A], byte(0))
	assertFlagsSet(t, c.reg[F], CFlag, ZFlag)
}

func TestShiftRightLogicalHL(t *testing.T) {
	c := NewTestCPU()
	c.reg[A] = 0x01
	c.ram[0] = 0xFF

	c.ShiftRightLogicalHL()
	assert.Equal(t, c.ram[c.ReadHL()], byte(0x7F))
	assertFlagsSet(t, c.reg[F], CFlag)
}
