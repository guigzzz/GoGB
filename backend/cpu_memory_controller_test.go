package backend

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadMemory(t *testing.T) {
	c := NewTestCPU()

	c.ram[0x3000] = 10
	assert.Equal(t, c.readMemory(0x3000), byte(10))

	c.cartridgeROM[0x4500] = 15
	assert.Equal(t, c.readMemory(0x4500), byte(15))

	c.ram[0xA500] = 20
	assert.Equal(t, c.readMemory(0xA500), byte(20))

	c.cartridgeRAMEnabled = true
	c.selectedRAMBank = 1
	c.writeMemory(0xA500, 25)
	assert.Equal(t, c.readMemory(0xA500), byte(25))

}

func TestWriteMemory(t *testing.T) {
	c := NewTestCPU()

	c.writeMemory(0x1000, 0)
	assert.Equal(t, c.cartridgeRAMEnabled, false)

	c.writeMemory(0x1000, 0xA)
	assert.Equal(t, c.cartridgeRAMEnabled, true)

	c.writeMemory(0x7000, 0)
	assert.Equal(t, c.ROMMode, true)

	c.writeMemory(0x3000, 0xFF)
	assert.Equal(t, c.selectedROMBank, byte(0x1F))

	c.writeMemory(0x5000, 0xFF)
	assert.Equal(t, c.selectedROMBank, byte(0x7F))

	c.writeMemory(0x6500, 1)
	assert.Equal(t, c.ROMMode, false)

	c.writeMemory(0x5000, 0xF)
	assert.Equal(t, c.selectedRAMBank, byte(0))

	c.writeMemory(0x5000, 0x20)
	assert.Equal(t, c.selectedRAMBank, byte(1))
}
