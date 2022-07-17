package backend

import "fmt"

type MBC5 struct {
	RamEnabled      bool
	SelectedROMBank uint16
	SelectedRAMBank byte

	Rom []byte
	Ram []byte

	NumRomBanks byte
	NumRamBanks byte
}

func NewMBC5(rom []byte, useRam, useBattery bool) *MBC5 {
	m := new(MBC5)

	m.SelectedROMBank = 1

	headerSize := getROMSize(rom[0x0148])
	if len(rom) != headerSize {
		panic(fmt.Sprintf(
			"Actual rom has size %d but header says it has size %d, this is inconsistent",
			len(rom), headerSize))
	}
	m.Rom = rom

	m.NumRomBanks = byte(headerSize / 0x4000)

	if useRam {
		ramSize := getRAMSize(rom[0x0149])
		m.NumRamBanks = byte(ramSize / 0x2000)
		m.Ram = make([]byte, getRAMSize(rom[0x0149]))
	}

	if useBattery {
		fmt.Println("MBC5 - WARNING: battery backed ram not implemented")
	}

	return m
}

func (m *MBC5) ReadMemory(address uint16) byte {

	if address < 0x4000 {
		return m.Rom[address]
	}
	if 0x4000 <= address && address < 0x8000 {

		offset := uint32(address) - 0x4000
		bankAddress := (uint32(m.SelectedROMBank) * 0x4000) + offset
		return m.Rom[bankAddress]

	} else if 0xA000 <= address && address < 0xC000 {

		if m.RamEnabled {
			offset := uint32(address) - 0xA000
			bankAddress := (uint32(m.SelectedRAMBank) * 0x2000) + offset
			return m.Ram[bankAddress]
		}

		return 0xFF
	}

	panic(fmt.Sprintf("Got unexpected read address not handled by MBC %d", address))
}

func (m *MBC5) WriteMemory(address uint16, value byte) {

	if address < 0x2000 {

		m.RamEnabled = value&0xA == 0xA

	} else if 0x2000 <= address && address < 0x3000 {
		m.SelectedROMBank = m.SelectedROMBank&0xFF00 | uint16(value)
		m.SelectedROMBank %= uint16(m.NumRomBanks)
	} else if 0x3000 <= address && address < 0x4000 {
		m.SelectedROMBank = m.SelectedROMBank&0xFF | uint16(value)<<8
		m.SelectedROMBank %= uint16(m.NumRomBanks)
	} else if 0x4000 <= address && address < 0x6000 {

		// either ram bank to use
		// or RTC register to access
		if m.RamEnabled {
			m.SelectedRAMBank = value % m.NumRamBanks
		}

	} else if 0x6000 <= address && address < 0x8000 {
		// latch clock data
	} else if 0xA000 <= address && address < 0xC000 {

		if m.RamEnabled {
			offset := uint32(address) - 0xA000
			bankAddress := (uint32(m.SelectedRAMBank) * 0x2000) + offset
			m.Ram[bankAddress] = value
		}

	} else {

		panic(fmt.Sprintf("Got unexpected write address not handled by MBC %d", address))
	}
}
