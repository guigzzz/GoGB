package backend

import (
	"crypto/md5"
	"io/ioutil"
	"testing"
	"time"
)

const (
	romPath = "../rom/cpu_instrs.gb"
)

func TestRunBlarggTests(t *testing.T) {

	rom, err := ioutil.ReadFile(romPath)
	if err != nil {
		panic(err)
	}

	cpu := NewCPU(rom, false)
	ppu := NewPPU(cpu)

	go ppu.Renderer()

	for cpu.PC != 0x06F1 && cpu.cycleCounter < 500000000 {
		time.Sleep(100 * time.Millisecond)
	}

	hasher := md5.New()
	for i := 0; i < 144; i++ {
		for j := 0; j < 160; j++ {
			pixel := ppu.Image.RGBAAt(j, i)
			hasher.Write([]byte{pixel.R, pixel.G, pixel.B, pixel.A})
		}
	}

	hash := hasher.Sum(nil)
	trueHash := []byte{208, 216, 82, 235, 32, 231, 249, 27, 62, 163, 210, 223, 40, 85, 174, 11}

	for i := range hash {
		if hash[i] != trueHash[i] {
			t.Errorf("Blargg test failed.")
			return
		}
	}
}
