package backend

import (
	"crypto/md5"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	blargg = "../rom/cpu_instrs.gb"
	timing = "../rom/instr_timing.gb"
	wario  = "../rom/wario_walking_demo.gb"
)

func apuFactory(c *CPU) APU {
	return &NullAPU{}
}

func Init(path string) (*PPU, *CPU) {
	rom, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}

	cpu := NewCPU(rom, false, NewNullLogger(), apuFactory)
	ppu := NewPPU(cpu)
	return ppu, cpu
}

func TestRunBlarggTests(t *testing.T) {

	ppu, cpu := Init(blargg)

	for cpu.PC != 0x06F1 && cpu.cycleCounter < 500000000 {
		ppu.RunEmulatorForAFrame()
	}

	// emulator state should be always exactly the same after the test passes
	assert.Equal(t, uint16(0x06F1), cpu.PC)
	assert.Equal(t, uint64(0xe023860), cpu.cycleCounter)

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

func TestRunInstrTimingTest(t *testing.T) {
	ppu, cpu := Init(timing)

	for cpu.PC != 0xc8b0 && cpu.cycleCounter < 500000000 {
		ppu.RunEmulatorForAFrame()
	}

	// emulator state should be always exactly the same after the test passes
	assert.Equal(t, uint16(0xc8b0), cpu.PC)
	assert.Equal(t, uint64(0x2b8178), cpu.cycleCounter)
}

func BenchmarkRunEmulatorForAFrame(b *testing.B) {
	ppu, _ := Init(wario)

	for n := 0; n < b.N; n++ {
		ppu.RunEmulatorForAFrame()
	}
}
