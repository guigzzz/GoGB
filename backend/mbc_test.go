package backend

import (
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
		// "mbc1/bits_mode.gb",
		// "mbc1/bits_ramg.gb",
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

			cpu := NewCPU(rom, false, NewNullLogger(), NullApuFactory)
			ppu := NewPPU(cpu)

			for i := 0; i < 200; i++ {
				ppu.RunEmulatorForAFrame()
			}

			if !assert.Equal(t, ref, ppu.Image) {
				name := strings.TrimSuffix(r, filepath.Ext(r))
				ppu.dumpScreenToPng("out/" + name + ".png")
			}
		})
	}

}
