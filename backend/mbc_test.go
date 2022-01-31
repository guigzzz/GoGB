package backend

import (
	"encoding/json"
	"io/ioutil"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

const mbcTestRomPath = "../rom/emulator-only"

func TestRunMbcTests(t *testing.T) {

	roms := []string{
		"mbc1/bits_bank1.gb",
		"mbc1/bits_bank2.gb",
		"mbc1/bits_mode.gb",
		"mbc1/bits_ramg.gb",
		// "mbc1/multicart_rom_8Mb.gb",
		"mbc1/ram_64kb.gb",
		"mbc1/ram_256kb.gb",
		"mbc1/rom_1Mb.gb",
		"mbc1/rom_2Mb.gb",
		"mbc1/rom_4Mb.gb",
		// "mbc1/rom_8Mb.gb",
		// "mbc1/rom_16Mb.gb",
		"mbc1/rom_512kb.gb",

		"mbc5/rom_1Mb.gb",
		"mbc5/rom_2Mb.gb",
		"mbc5/rom_4Mb.gb",
		"mbc5/rom_8Mb.gb",
		"mbc5/rom_16Mb.gb",
		// "mbc5/rom_32Mb.gb",
		// "mbc5/rom_64Mb.gb",
		"mbc5/rom_512kb.gb",
	}

	ref := getImage("ref/mbc.png")

	for _, r := range roms {

		t.Run(r, func(t *testing.T) {
			rom, err := ioutil.ReadFile(path.Join(mbcTestRomPath, r))
			if err != nil {
				panic(err)
			}

			ppu, _, _, _ := composeForTests(rom)

			for i := 0; i < 500; i++ {
				ppu.RunEmulatorForAFrame()
			}

			if !assert.Equal(t, ref, ppu.Image) {
				name := strings.TrimSuffix(r, filepath.Ext(r))
				ppu.dumpScreenToPng("out/" + name + ".png")
			}
		})
	}

}

func makeRom() []byte {
	rom := make([]byte, 65536)
	rom[0x148] = 1
	rom[0x149] = 1
	return rom
}

func runTest(t *testing.T, m MBC) {
	wrapped := MbcWrapper{m}

	d, err := json.Marshal(wrapped)
	if err != nil {
		t.Error(err)
	}

	var out MbcWrapper
	if e := json.Unmarshal(d, &out); e != nil {
		t.Error(e)
	}

	assert.Equal(t, out.mbc, m)
}

func TestMarshalMbc0(t *testing.T) {
	rom := make([]byte, 32768)
	mbc := NewMBC0(rom)

	runTest(t, mbc)
}

func TestMarshalMbc1(t *testing.T) {
	mbc := NewMBC1(makeRom(), true, false)
	mbc.NumRomBanks = 1
	mbc.NumRamBanks = 2
	mbc.SelectedRAMBank = 3
	mbc.SelectedROMBank = 4

	runTest(t, mbc)
}

func TestMarshalMbc3(t *testing.T) {
	mbc := NewMBC3(makeRom(), true, false, false)
	mbc.SelectedRAMBank = 3
	mbc.SelectedROMBank = 4
	mbc.RamEnabled = true

	runTest(t, mbc)
}

func TestMarshalMbc5(t *testing.T) {
	mbc := NewMBC5(makeRom(), true, false)
	mbc.NumRomBanks = 1
	mbc.NumRamBanks = 2
	mbc.SelectedRAMBank = 3
	mbc.SelectedROMBank = 4
	mbc.RamEnabled = true

	runTest(t, mbc)
}
