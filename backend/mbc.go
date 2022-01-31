package backend

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// MBC represents a memory bank controller
type MBC interface {
	ReadMemory(uint16) byte
	WriteMemory(uint16, byte)
}

type MbcWrapper struct {
	mbc MBC
}

func getType(myvar interface{}) string {
	if t := reflect.TypeOf(myvar); t.Kind() == reflect.Ptr {
		return t.Elem().Name()
	} else {
		return t.Name()
	}
}

const TYPE_FIELD = "Type"
const MBC_FIELD = "MBC"

func (w MbcWrapper) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		TYPE_FIELD: getType(w.mbc),
		MBC_FIELD:  w.mbc,
	})
}

func (w *MbcWrapper) UnmarshalJSON(d []byte) error {

	var objMap map[string]*json.RawMessage
	if e := json.Unmarshal(d, &objMap); e != nil {
		return e
	}

	var t string
	if e := json.Unmarshal(*objMap[TYPE_FIELD], &t); e != nil {
		return e
	}

	v := *objMap[MBC_FIELD]
	switch t {
	case "MBC0":
		var mbc MBC0
		if e := json.Unmarshal(v, &mbc); e != nil {
			return e
		}
		w.mbc = &mbc
	case "MBC1":
		var mbc MBC1
		if e := json.Unmarshal(v, &mbc); e != nil {
			return e
		}
		w.mbc = &mbc
	case "MBC3":
		var mbc MBC3
		if e := json.Unmarshal(v, &mbc); e != nil {
			return e
		}
		w.mbc = &mbc
	case "MBC5":
		var mbc MBC5
		if e := json.Unmarshal(v, &mbc); e != nil {
			return e
		}
		w.mbc = &mbc
	default:
		panic("Got unexpected type: " + t)
	}

	return nil
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

func NewMBC(rom []byte) MBC {

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
		return NewMBC5(rom, false, false)
	case 0x1A:
		return NewMBC5(rom, true, false)
	case 0x1B:
		return NewMBC5(rom, true, true)
	case 0x1C:
		fmt.Println("WARNING: MBC5 with rumble requested")
		return NewMBC5(rom, false, false)
	case 0x1D:
		fmt.Println("WARNING: MBC5 with rumble requested")
		// what is Battery-backed rumble ??
		return NewMBC5(rom, false, false)
	case 0x1E:
		fmt.Println("WARNING: MBC5 with rumble requested")
		return NewMBC5(rom, true, true)

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
