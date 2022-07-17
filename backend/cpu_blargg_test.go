package backend

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	blargg = "../rom/cpu_instrs.gb"
	timing = "../rom/instr_timing.gb"
	wario  = "../rom/wario_walking_demo.gb"
)

func Init(path string) (*Emulator, *RecordingLogger) {
	logger := NewRecordingLogger()

	emulator := NewEmulator(WithRom(path), WithLogger(logger), WithDisableApu())

	emulator.ppu.ram[LCDC] |= 1 << lcdDisplayEnable

	return emulator, logger
}

const EXPECTED_SUCCESS_LOG = "cpu_instrs\n\n01:ok  02:ok  03:ok  04:ok  05:ok  06:ok  07:ok  08:ok  09:ok  10:ok  11:ok  \n\nPassed all tests\n"

func TestRunBlarggTests(t *testing.T) {

	emulator, logger := Init(blargg)

	cpu := emulator.cpu

	for cpu.PC != 0x06F1 && cpu.cycleCounter < 500000000 {
		emulator.RunForAFrame()
	}

	// emulator state should be always exactly the same after the test passes
	assert.Equal(t, uint16(0x06F1), cpu.PC)
	assert.Equal(t, uint64(0xe023860), cpu.cycleCounter)

	assert.Equal(t, EXPECTED_SUCCESS_LOG, logger.contents)

	ref := getImage("ref/blargg.png")
	if !assert.Equal(t, ref, emulator.GetImage()) {
		emulator.dumpScreenToPng("out/blargg.png")
	}
}

const INSTR_TIMING_SUCCESS_LOG = "instr_timing\n\n\nPassed\n"

func TestRunInstrTimingTest(t *testing.T) {
	emulator, logger := Init(timing)

	cpu := emulator.cpu

	for cpu.PC != 0xc8b0 && cpu.cycleCounter < 500000000 {
		emulator.RunForAFrame()
	}

	assert.Equal(t, INSTR_TIMING_SUCCESS_LOG, logger.contents)

	// emulator state should be always exactly the same after the test passes
	assert.Equal(t, uint16(0xc8b0), cpu.PC)
	assert.Equal(t, uint64(0x2b8178), cpu.cycleCounter)
}

func BenchmarkRunEmulatorForAFrame(b *testing.B) {
	emulator, _ := Init(wario)

	for n := 0; n < b.N; n++ {
		emulator.RunForAFrame()
	}
}
