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

func Init(path string) (*PPU, *CPU, *RecordingLogger) {
	rom, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}

	logger := NewRecordingLogger()
	cpu := NewCPU(rom, false, logger, apuFactory)
	ppu := NewPPU(cpu)
	return ppu, cpu, logger
}

const EXPECTED_SUCCESS_LOG = "cpu_instrs\n\n01:ok  02:ok  03:ok  04:ok  05:ok  06:ok  07:ok  08:ok  09:ok  10:ok  11:ok  \n\nPassed all tests\n"

func TestRunBlarggTests(t *testing.T) {

	ppu, cpu, logger := Init(blargg)

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

	assert.Equal(t, EXPECTED_SUCCESS_LOG, logger.contents)

	hash := hasher.Sum(nil)
	trueHash := []byte{208, 216, 82, 235, 32, 231, 249, 27, 62, 163, 210, 223, 40, 85, 174, 11}

	for i := range hash {
		if hash[i] != trueHash[i] {
			t.Errorf("Blargg test failed.")
			return
		}
	}
}

const INSTR_TIMING_SUCCESS_LOG = "instr_timing\n\n\nPassed\n"

func TestRunInstrTimingTest(t *testing.T) {
	ppu, cpu, logger := Init(timing)

	for cpu.PC != 0xc8b0 && cpu.cycleCounter < 500000000 {
		ppu.RunEmulatorForAFrame()
	}

	assert.Equal(t, INSTR_TIMING_SUCCESS_LOG, logger.contents)

	// emulator state should be always exactly the same after the test passes
	assert.Equal(t, uint16(0xc8b0), cpu.PC)
	assert.Equal(t, uint64(0x2b8178), cpu.cycleCounter)
}

func BenchmarkRunEmulatorForAFrame(b *testing.B) {
	ppu, _, _ := Init(wario)

	for n := 0; n < b.N; n++ {
		ppu.RunEmulatorForAFrame()
	}
}
