package backend

import "fmt"

type MBC3 struct {
	RamEnabled      bool
	SelectedROMBank byte
	SelectedRAMBank byte

	Rom []byte
	Ram []byte
}

func NewMBC3(rom []byte, useRam, useTimer, useBattery bool) *MBC3 {
	m := new(MBC3)

	m.SelectedROMBank = 1

	headerSize := getROMSize(rom[0x0148])
	if len(rom) != headerSize {
		panic(fmt.Sprintf(
			"Actual rom has size %d but header says it has size %d, this is inconsistent",
			len(rom), headerSize))
	}
	m.Rom = rom

	if useRam {
		m.Ram = make([]byte, getRAMSize(rom[0x0149]))
	} else {
		m.Ram = make([]byte, 0)
	}

	return m
}

func (m *MBC3) ReadMemory(address uint16) byte {

	if 0x0000 <= address && address < 0x4000 {
		return m.Rom[address]
	}
	if 0x4000 <= address && address < 0x8000 {
		offset := uint32(address) - 0x4000
		bankAddress := (uint32(m.SelectedROMBank) * 0x4000) + offset
		return m.Rom[bankAddress]
	}
	if 0xA000 <= address && address < 0xC000 {
		if m.RamEnabled {
			if m.SelectedRAMBank < 8 {
				offset := uint32(address) - 0xA000
				bankAddress := (uint32(m.SelectedRAMBank) * 0x2000) + offset
				return m.Ram[bankAddress]
			}
			panic("MBC3: RTC unimplemented")
		}
		return 0xFF
	}

	panic(fmt.Sprintf("Got unexpected read address not handled by MBC %d", address))
}

func (m *MBC3) WriteMemory(address uint16, value byte) {

	if 0x0000 <= address && address < 0x2000 {

		m.RamEnabled = value&0xA == 0xA

	} else if 0x2000 <= address && address < 0x4000 {
		value &= 0x7F // mask off lower 7 bits
		if value == 0 {
			value++
		}
		m.SelectedROMBank = value
	} else if 0x4000 <= address && address < 0x6000 {

		// either rom bank to use
		// or RTC register to access
		m.SelectedRAMBank = value

	} else if 0x6000 <= address && address < 0x8000 {
		// latch clock data
	} else if 0xA000 <= address && address < 0xC000 {

		if m.RamEnabled && len(m.Ram) > 0 {
			offset := uint32(address) - 0xA000
			bankAddress := (uint32(m.SelectedRAMBank) * 0x2000) + offset
			m.Ram[bankAddress] = value
		}

	} else {

		panic(fmt.Sprintf("Got unexpected write address not handled by MBC %d", address))
	}
}
