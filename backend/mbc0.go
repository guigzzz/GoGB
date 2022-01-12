package backend

import "fmt"

type MBC0 struct {
	Rom []byte
}

func NewMBC0(rom []byte) *MBC0 {
	m := new(MBC0)

	if len(rom) > (1 << 15) {
		panic("Cartridge has no MBC but the rom is larger than 32KB, this is inconsistent")
	}

	m.Rom = rom

	return m
}

func (m *MBC0) ReadMemory(address uint16) byte {
	if 0x0000 <= address && address < 0x8000 {
		return m.Rom[address]
	} else if 0xA000 <= address && address < 0xC000 {
		return 0xFF
	}
	panic(fmt.Sprintf("Got unexpected read address not handled by MBC %d", address))
}

func (m *MBC0) WriteMemory(address uint16, value byte) {
	if 0x0000 <= address && address < 0x8000 ||
		0xA000 <= address && address < 0xC000 {

		// ignore all writes to these address blocks

	} else {
		panic(fmt.Sprintf("Got unexpected write address not handled by MBC %d", address))
	}
}
