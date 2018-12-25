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
