package backend

import "fmt"

type MBC3 struct {
	ramEnabled      bool
	selectedROMBank byte
	selectedRAMBank byte

	rom []byte
	ram []byte
}

func NewMBC3(rom []byte, useRam, useTimer, useBattery bool) *MBC3 {
	m := new(MBC3)

	m.selectedROMBank = 1

	headerSize := getROMSize(rom[0x0148])
	if len(rom) != headerSize {
		panic(fmt.Sprintf(
			"Actual rom has size %d but header says it has size %d, this is inconsistent",
			len(rom), headerSize))
	}
	m.rom = rom

	if useRam {
		m.ram = make([]byte, getRAMSize(rom[0x0149]))
	} else {
		m.ram = make([]byte, 0)
	}

	return m
}

func (m *MBC3) DelegateReadToMBC(address uint16) bool {
	return 0x0000 <= address && address < 0x8000 ||
		0xA000 <= address && address < 0xC000
}

func (m *MBC3) DelegateWriteToMBC(address uint16) bool {
	return 0x0000 <= address && address < 0x8000 ||
		0xA000 <= address && address < 0xC000
}

func (m *MBC3) ReadMemory(address uint16) byte {

	if 0x0000 <= address && address < 0x4000 {
		return m.rom[address]
	}
	if 0x4000 <= address && address < 0x8000 {
		offset := uint32(address) - 0x4000
		bankAddress := (uint32(m.selectedROMBank) * 0x4000) + offset
		return m.rom[bankAddress]
	}
	if 0xA000 <= address && address < 0xC000 {
		if m.ramEnabled {
			if m.selectedRAMBank < 8 {
				offset := uint32(address) - 0xA000
				bankAddress := (uint32(m.selectedRAMBank) * 0x2000) + offset
				return m.ram[bankAddress]
			}
			panic("MBC3: RTC unimplemented")
		}
		return 0xFF
	}

	panic(fmt.Sprintf("Got unexpected read address not handled by MBC %d", address))
}

func (m *MBC3) WriteMemory(address uint16, value byte) {

	if 0x0000 <= address && address < 0x2000 {

		m.ramEnabled = value&0xA == 0xA

	} else if 0x2000 <= address && address < 0x4000 {
		value &= 0x7F // mask off lower 7 bits
		if value == 0 {
			value++
		}
		m.selectedROMBank = value
	} else if 0x4000 <= address && address < 0x6000 {

		// either rom bank to use
		// or RTC register to access
		m.selectedRAMBank = value

	} else if 0x6000 <= address && address < 0x8000 {
		// latch clock data
	} else if 0xA000 <= address && address < 0xC000 {

		if m.ramEnabled && len(m.ram) > 0 {
			offset := uint32(address) - 0xA000
			bankAddress := (uint32(m.selectedRAMBank) * 0x2000) + offset
			m.ram[bankAddress] = value
		}

	} else {

		panic(fmt.Sprintf("Got unexpected write address not handled by MBC %d", address))
	}
}
