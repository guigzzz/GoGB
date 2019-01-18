package backend

import "fmt"

// MBC represents a memory bank controller
type MBC interface {
	ReadMemory(uint16) byte
	WriteMemory(uint16, byte)
	DelegateReadToMBC(uint16) bool
	DelegateWriteToMBC(uint16) bool
}

func getROMSize(sizeIndex byte) int {
	if sizeIndex > 8 {
		panic(fmt.Sprintf("Got invalid rom size index %d", sizeIndex))
	}
	return (1 << 15) << sizeIndex
}

func getRAMSize(sizeIndex byte) int {
	switch sizeIndex {
	case 0:
		return 0
	case 1:
		return 1 << 11
	case 2:
		return 1 << 13
	case 3:
		return 1 << 15
	case 4:
		return 1 << 17
	case 5:
		return 1 << 16
	default:
		panic(fmt.Sprintf("Got invalid rom size index %d", sizeIndex))
	}
}

func getMemoryControllerFrom(rom []byte) MBC {

	mbcNumber := rom[0x147]

	switch mbcNumber {
	case 0x00:
		return NewMBC0(rom)
	case 0x01:
		return NewMBC1(rom, false, false)
	case 0x02:
		return NewMBC1(rom, true, false)
	case 0x03:
		fmt.Println("WARNING: this games uses MBC1 with battery, presumably for saves, however GoGB doesnt support that")
		return NewMBC1(rom, true, true)

	case 0x05:
		panic("MBC2 unimplemented")
	case 0x06:
		panic("MBC2 + RAM + Battery unimplemented")

	case 0x08:
		panic("ROM + RAM unimplemented")
	case 0x09:
		panic("ROM + RAM + Battery unimplemented")

	case 0x0B:
		panic("MMM01 unimplemented")
	case 0x0C:
		panic("MMM01 + RAM unimplemented")
	case 0x0D:
		panic("MMM01 + RAM + Battery unimplemented")

	case 0x0F:
		panic("MBC3 + Timer + Battery unimplemented")
	case 0x10:
		panic("MBC3 + RAM + Timer + Battery unimplemented")
	case 0x11:
		return NewMBC3(rom, false, false, false)
	case 0x12:
		return NewMBC3(rom, true, false, false)
	case 0x13:
		fmt.Println("WARNING: this games uses MBC3 with battery, presumably for saves, however GoGB doesnt support that")
		return NewMBC3(rom, true, false, true)

	case 0x19:
		panic("MBC5 unimplemented")
	case 0x1A:
		panic("MBC5 + RAM unimplemented")
	case 0x1B:
		panic("MBC5 + Rumble unimplemented")
	case 0x1C:
		panic("MBC5 + RAM + Rumble uimplemented")
	case 0x1D:
		panic("MBC5 + RAM + Battery + Rumble unimplemented")

	case 0x20:
		panic("MBC6 + RAM + Battery unimplemented")
	case 0x22:
		panic("MBC7 + RAM + Bat. + Accelerometer unimplemented")

	case 0xFC:
		panic("POCKET CAMERA unimplemented")
	case 0xFD:
		panic("BANDAI TAMA5 unimplemented")
	case 0xFE:
		panic("HuC3 unimplemented")
	case 0xFF:
		panic("HuC1 + RAM + Battery unimplemented")

	default:
		panic(fmt.Sprintf("Got unusued mbcNumber %d\n", mbcNumber))
	}
}
