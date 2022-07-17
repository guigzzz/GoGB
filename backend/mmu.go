package backend

type MMU struct {
	ram []byte

	KeyPressedMap map[string]bool
	mbc           MBC

	logger Logger

	audioRegisterWriteCallback AudioRegisterWriteCallback
}

type AudioRegisterWriteCallback = func(uint16, byte, byte)

func NewMMU(ram []byte, mbc MBC, logger Logger, audioRegisterWriteCallback AudioRegisterWriteCallback) *MMU {
	mmu := new(MMU)

	mmu.ram = ram
	mmu.mbc = mbc

	mmu.KeyPressedMap = map[string]bool{
		"up": false, "down": false, "left": false, "right": false,
		"A": false, "B": false, "start": false, "select": false,
	}

	if logger == nil {
		mmu.logger = NewPrintLogger()
	} else {
		mmu.logger = logger
	}

	mmu.audioRegisterWriteCallback = audioRegisterWriteCallback

	return mmu
}

var audioRegOrLookup = [32]byte{
	0x80, 0x3F, 0x00, 0xFF, 0xBF, // NR10-NR14
	0xFF, 0x3F, 0x00, 0xFF, 0xBF, // NR20-NR24
	0x7F, 0xFF, 0x9F, 0xFF, 0xBF, // NR30-NR34
	0xFF, 0xFF, 0x00, 0x00, 0xBF, // NR40-NR44
	0x00, 0x00, 0x70, // NR50-NR52
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, // unused regs
}

func delegateToMBC(address uint16) bool {
	return address < 0x8000 || 0xA000 <= address && address < 0xC000
}

func (m *MMU) readMemory(address uint16) byte {

	if delegateToMBC(address) {

		return m.mbc.ReadMemory(address)

	} else if 0xFEA0 <= address && address < 0xFF00 {
		return 00
	} else if 0xFF10 <= address && address <= 0xFF2F {
		// audio regs
		or := audioRegOrLookup[address-0xFF10]
		return m.ram[address] | or
	}
	return m.ram[address]
}

func (m *MMU) writeMemory(address uint16, value byte) {

	if delegateToMBC(address) {

		m.mbc.WriteMemory(address, value)

	} else if 0xFEA0 <= address && address < 0xFF00 {
		// ignore
	} else if 0xFF10 <= address && address <= 0xFF2F {
		oldValue := m.ram[address]
		// audio
		if address == 0xFF26 {
			// NR52 only top bits are writeable
			m.ram[address] = value & 0xF0

			// check if APU disabled. If yes, then clear all regs
			if value&0x80 == 0 {
				for i := 0xFF10; i <= 0xFF2F; i++ {
					m.ram[i] = 0
				}
			}

		} else {
			// ignore all writes if APU disabled
			isOn := m.ram[0xFF26]&0x80 > 0
			if isOn {
				m.ram[address] = value
			}
		}
		m.audioRegisterWriteCallback(address, oldValue, value)
	} else {
		if address == 0xFF02 && value == 0x81 {
			m.logger.Log(string(m.ram[0xFF01]))
		} else if address == 0xFF46 {
			m.DMA(value)
			m.ram[0xFF46] = value
		} else if address == 0xFF00 {
			m.ram[0xFF00] = 0b1100_0000 | (value & 0b11_0000) | m.readKeyPressed(value)
		} else if address == 0xFF04 {
			// when DIV is written
			// it is reset to 0
			m.ram[0xFF04] = 0
		} else {
			m.ram[address] = value
		}
	}
}

func (m *MMU) DMA(sourceAddress byte) {
	blockAddress := uint16(sourceAddress) << 8
	for i := uint16(0); i < 0xA0; i++ {
		m.ram[0xFE00+i] = m.ram[blockAddress+i]
	}
}

func (m *MMU) readKeyPressed(code byte) byte {
	regValue := byte(0xF)
	if code&0x20 == 0 { // 0b1101_1111
		if m.KeyPressedMap["start"] {
			regValue &^= 0x8
		}
		if m.KeyPressedMap["select"] {
			regValue &^= 0x4
		}
		if m.KeyPressedMap["B"] {
			regValue &^= 0x2
		}
		if m.KeyPressedMap["A"] {
			regValue &^= 0x1
		}
	}
	if code&0x10 == 0 { // 0b1110_1111
		if m.KeyPressedMap["down"] {
			regValue &^= 0x8
		}
		if m.KeyPressedMap["up"] {
			regValue &^= 0x4
		}
		if m.KeyPressedMap["left"] {
			regValue &^= 0x2
		}
		if m.KeyPressedMap["right"] {
			regValue &^= 0x1
		}
	}
	return regValue
}
