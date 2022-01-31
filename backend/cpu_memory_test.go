package backend

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoad(t *testing.T) {
	c := NewTestCPU()

	c.reg[D] = 1

	c.LoadReg(B, D)
	assert.Equal(t, c.reg[B], byte(1))

	c.Load(B, 10)
	assert.Equal(t, c.reg[B], byte(10))
}

func TestLoadHL(t *testing.T) {
	c := NewTestCPU()

	c.Writedouble(H, L, 0xC0FF)

	c.writeMemory(c.ReadHL(), 5)
	c.LoadHL(A)
	assert.Equal(t, c.reg[A], byte(5))

	c.reg[A] = 25
	c.Writedouble(H, L, 0xC000)
	c.StoreReg(A)
	assert.Equal(t, c.readMemory(c.ReadHL()), byte(25))
}

func TestLoadHigh(t *testing.T) {
	c := NewTestCPU()
	c.writeMemory(0xFF40, 10)

	c.LoadHigh(0x40)
	assert.Equal(t, c.reg[A], byte(10))

	c.reg[A] = 100
	c.StoreHigh(0x50)
	assert.Equal(t, c.readMemory(0xFF50), byte(100))
}

func TestPushPC(t *testing.T) {
	c := NewTestCPU()
	c.SP = 0xFFFE

	c.PC = 0xF00F
	c.pushPC()

	assert.Equal(t, c.readMemory(0xFFFD), byte(0xF0))
	assert.Equal(t, c.readMemory(0xFFFC), byte(0x0F))
	assert.Equal(t, c.SP, uint16(0xFFFC))
}

func TestLDHL(t *testing.T) {
	c := NewTestCPU()

	c.SP = 0xFFF8
	c.LoadHLSPN(2)

	assert.Equal(t, c.ReadHL(), uint16(0xFFFA))
	assertFlagsSet(t, c.reg[F])
}

func TestStoreSPNN(t *testing.T) {
	c := NewTestCPU()
	c.SP = 0xFFF8
	c.StoreSPNN(0xC100)

	assert.Equal(t, c.readMemory(0xC100), byte(0xF8))
	assert.Equal(t, c.readMemory(0xC101), byte(0xFF))
}

const JOYPAD = 0xFF00

func TestJoypadNothingPressed(t *testing.T) {
	c := NewTestCPU()

	c.writeMemory(JOYPAD, 0b01_1111)

	v := c.readMemory(JOYPAD)
	assert.Equal(t, byte(0b1101_1111), v)
}

func TestJoypadPressed(t *testing.T) {
	c := NewTestCPU()

	c.mmu.KeyPressedMap["start"] = true

	c.writeMemory(JOYPAD, 0b01_1111)

	v := c.readMemory(JOYPAD)
	assert.Equal(t, byte(0b1101_0111), v)
}
