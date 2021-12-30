package backend

import (
	"fmt"
	"io"
)

type APU interface {
	ToReadCloser() io.ReadCloser
	StepAPU()
	AudioRegisterWriteCallback(addr uint16, value byte)
}

type NullAPU struct{}

func (a *NullAPU) StepAPU()                                    {}
func (a *NullAPU) Read(p []byte) (int, error)                  { return 0, nil }
func (a *NullAPU) ToReadCloser() io.ReadCloser                 { return io.NopCloser(a) }
func (a *NullAPU) AudioRegisterWriteCallback(_ uint16, _ byte) {}

const (
	// Square 1
	NR10 = 0xFF10 // -PPP NSSS Sweep period, negate, shift
	NR11 = 0xFF11 // DDLL LLLL Duty, Length load (64-L)
	NR12 = 0xFF12 // VVVV APPP Starting volume, Envelope add mode, period
	NR13 = 0xFF13 // FFFF FFFF Frequency LSB
	NR14 = 0xFF14 // TL-- -FFF Trigger, Length enable, Frequency MSB

	// Square 2
	// FF15 -- not used
	NR21 = 0xFF16 // DDLL LLLL Duty, Length load (64-L)
	NR22 = 0xFF17 // VVVV APPP Starting volume, Envelope add mode, period
	NR23 = 0xFF18 // FFFF FFFF Frequency LSB
	NR24 = 0xFF19 // TL-- -FFF Trigger, Length enable, Frequency MSB

	// Wave
	NR30 = 0xFF1A // E--- ---- DAC power
	NR31 = 0xFF1B // LLLL LLLL Length load (256-L)
	NR32 = 0xFF1C // -VV- ---- Volume code (00=0%, 01=100%, 10=50%, 11=25%)
	NR33 = 0xFF1D // FFFF FFFF Frequency LSB
	NR34 = 0xFF1E // TL-- -FFF Trigger, Length enable, Frequency MSB

	// Noise
	// FF1F -- not used
	NR41 = 0xFF20 // --LL LLLL Length load (64-L)
	NR42 = 0xFF21 // VVVV APPP Starting volume, Envelope add mode, period
	NR43 = 0xFF22 // SSSS WDDD Clock shift, Width mode of LFSR, Divisor code
	NR44 = 0xFF23 // TL-- ---- Trigger, Length enable

	// Control/Status
	NR50 = 0xFF24 // ALLL BRRR Vin L enable, Left vol, Vin R enable, Right vol
	NR51 = 0xFF25 // NW21 NW21 Left enables, Right enables
	NR52 = 0xFF26 // P--- NW21 Power control/status, Channel length statuses
)

var WAVE_DUTY_TABLE = [4][8]byte{
	{0, 0, 0, 0, 0, 0, 0, 1},
	{0, 0, 0, 0, 0, 0, 1, 1},
	{0, 0, 0, 0, 1, 1, 1, 1},
	{1, 1, 1, 1, 1, 1, 0, 0},
}

var CODE_TO_DIVISOR = [8]byte{8, 16, 32, 48, 64, 80, 96, 112}

type APUImpl struct {
	ram []byte

	cycleCounter int

	waveDutyPositionSquare1 int
	frequencyTimerSquare1   int
	lengthTimerSquare1      int

	waveDutyPositionSquare2 int
	frequencyTimerSquare2   int
	lengthTimerSquare2      int

	lengthTimerWave int

	frequencyTimerNoise int
	lengthTimerNoise    int
	lsfr                int

	frameSequencerCounter byte

	sampleBuf []byte
	samples   chan []byte
}

const (
	SAMPLE_BUFFER_SIZE = 1024
)

func NewAPU(c *CPU) *APUImpl {
	apu := new(APUImpl)

	apu.ram = c.ram

	apu.cycleCounter = 0
	apu.waveDutyPositionSquare1 = 0
	apu.waveDutyPositionSquare2 = 0
	apu.frequencyTimerSquare1 = 0
	apu.frequencyTimerSquare2 = 0
	apu.frequencyTimerNoise = 0
	apu.lsfr = 0
	apu.lengthTimerSquare1 = 0
	apu.lengthTimerSquare2 = 0
	apu.lengthTimerWave = 0
	apu.lengthTimerNoise = 0

	apu.sampleBuf = make([]byte, SAMPLE_BUFFER_SIZE)
	apu.samples = make(chan []byte)

	return apu
}

func (a *APUImpl) isByteBitSet(addr uint16, bit uint) bool {
	if bit > 7 {
		panic(fmt.Sprintf("Unexpected byte bit: %d", bit))
	}
	return (a.ram[addr]>>bit)&1 > 0
}

func boolToNum(b bool) byte {
	if b {
		return 1
	}
	return 0
}

func (a *APUImpl) getSquare1Output() (byte, byte) {
	duty := a.ram[NR11] & 0b1100_000 >> 6
	volume := byte(0xFF) // a.ram[NR12] & 0b1111_0000 >> 4

	amplitude := WAVE_DUTY_TABLE[duty][a.waveDutyPositionSquare1]
	output := amplitude * volume

	leftEnable := a.isByteBitSet(NR51, 4)
	rightEnable := a.isByteBitSet(NR51, 0)

	leftOutput := boolToNum(leftEnable) * output
	rightOutput := boolToNum(rightEnable) * output
	return leftOutput, rightOutput
}

func (a *APUImpl) getSquare2Output() (byte, byte) {
	duty := a.ram[NR21] & 0b1100_000 >> 6
	volume := byte(0xFF) // a.ram[NR22] & 0b1111_0000 >> 4

	amplitude := WAVE_DUTY_TABLE[duty][a.waveDutyPositionSquare2]
	output := amplitude * volume

	leftEnable := a.isByteBitSet(NR51, 5)
	rightEnable := a.isByteBitSet(NR51, 1)

	leftOutput := boolToNum(leftEnable) * output
	rightOutput := boolToNum(rightEnable) * output
	return leftOutput, rightOutput
}

func (a *APUImpl) getNoiseOutput() (byte, byte) {
	volume := byte(0xFF) // a.ram[NR42] & 0b1111_0000 >> 4
	amplitude := byte((^a.lsfr) & 1)
	output := amplitude * volume

	leftEnable := a.isByteBitSet(NR51, 7)
	rightEnable := a.isByteBitSet(NR51, 3)

	leftOutput := boolToNum(leftEnable) * output
	rightOutput := boolToNum(rightEnable) * output
	return leftOutput, rightOutput
}

func (a *APUImpl) AudioRegisterWriteCallback(addr uint16, value byte) {

	switch addr {
	case NR11:
		lengthLoad := a.ram[NR11] & 0b11_1111
		a.lengthTimerSquare1 = 64 - int(lengthLoad)
	case NR12:
		dacDisabled := value&0xF8 == 0
		if dacDisabled {
			a.clearBit(NR52, 0)
		}
	case NR14:
		if a.lengthTimerSquare1 == 0 {
			a.lengthTimerSquare1 = 64
		}
		isTrigger := value&0x80 > 0
		dacEnabled := a.ram[NR12]&0xF8 > 0
		if isTrigger && dacEnabled {
			a.setBit(NR52, 0)
		}
	case NR21:
		lengthLoad := a.ram[NR21] & 0b11_1111
		a.lengthTimerSquare2 = 64 - int(lengthLoad)
	case NR22:
		dacDisabled := value&0xF8 == 0
		if dacDisabled {
			a.clearBit(NR52, 1)
		}
	case NR24:
		if a.lengthTimerSquare2 == 0 {
			a.lengthTimerSquare2 = 64
		}
		isTrigger := value&0x80 > 0
		dacEnabled := a.ram[NR22]&0xF8 > 0
		if isTrigger && dacEnabled {
			a.setBit(NR52, 1)
		}
	case NR30:
		dacDisabled := value&0x80 == 0
		if dacDisabled {
			a.clearBit(NR52, 2)
		}
	case NR31:
		a.lengthTimerWave = 256 - int(a.ram[NR31])
	case NR34:
		if a.lengthTimerWave == 0 {
			a.lengthTimerWave = 256
		}
		isTrigger := value&0x80 > 0
		dacEnabled := a.ram[NR30]&0x80 > 0
		if isTrigger && dacEnabled {
			a.setBit(NR52, 2)
		}
	case NR41:
		lengthLoad := a.ram[NR41] & 0b11_1111
		a.lengthTimerNoise = 64 - int(lengthLoad)
	case NR42:
		dacDisabled := value&0xF8 == 0
		if dacDisabled {
			a.clearBit(NR52, 3)
		}
	case NR44:
		if a.lengthTimerNoise == 0 {
			a.lengthTimerNoise = 64
		}
		isTrigger := value&0x80 > 0
		dacEnabled := a.ram[NR42]&0xF8 > 0
		if isTrigger && dacEnabled {
			a.setBit(NR52, 3)
		}
	}
}

func (a *APUImpl) clearBit(addr uint16, bit uint) {
	if bit > 7 {
		panic("unexpected")
	}
	a.ram[addr] &^= 1 << bit
}

func (a *APUImpl) setBit(addr uint16, bit uint) {
	if bit > 7 {
		panic("unexpected")
	}
	a.ram[addr] |= 1 << bit
}

func (a *APUImpl) updateLengthTimers() {
	if a.isByteBitSet(NR14, 6) && a.lengthTimerSquare1 > 0 {
		a.lengthTimerSquare1--
		if a.lengthTimerSquare1 == 0 {
			a.clearBit(NR52, 0)
		}
	}

	if a.isByteBitSet(NR24, 6) && a.lengthTimerSquare2 > 0 {
		a.lengthTimerSquare2--
		if a.lengthTimerSquare2 == 0 {
			a.clearBit(NR52, 1)
		}
	}

	if a.isByteBitSet(NR34, 6) && a.lengthTimerWave > 0 {
		a.lengthTimerWave--
		if a.lengthTimerWave == 0 {
			a.clearBit(NR52, 2)
		}
	}

	if a.isByteBitSet(NR44, 6) && a.lengthTimerNoise > 0 {
		a.lengthTimerNoise--
		if a.lengthTimerNoise == 0 {
			a.clearBit(NR52, 3)
		}
	}
}

func (a *APUImpl) updateFrameSequencer() {

	if a.cycleCounter%8192 > 0 {
		return
	}

	if a.frameSequencerCounter%2 == 0 {
		a.updateLengthTimers()
	}

	// if a.frameSequencerCounter%8 == 7 {
	// 	// vol env
	// }

	// modFour := a.frameSequencerCounter % 4
	// if modFour == 2 || modFour == 6 {
	// 	// sweep
	// }

	a.frameSequencerCounter++
}

func (a *APUImpl) updateState() {

	a.updateFrameSequencer()

	a.frequencyTimerSquare1--
	if a.frequencyTimerSquare1 <= 0 {
		lsb := a.ram[NR13]
		msb := a.ram[NR14] & 0b111
		frequency := uint16(msb)<<8 | uint16(lsb)
		a.frequencyTimerSquare1 = int((2048-frequency)*4) + a.frequencyTimerSquare1
		a.waveDutyPositionSquare1 = (a.waveDutyPositionSquare1 + 1) % 8
	}

	a.frequencyTimerSquare2--
	if a.frequencyTimerSquare2 <= 0 {
		lsb := a.ram[NR23]
		msb := a.ram[NR24] & 0b111
		frequency := uint16(msb)<<8 | uint16(lsb)
		a.frequencyTimerSquare2 = int((2048-frequency)*4) + a.frequencyTimerSquare2
		a.waveDutyPositionSquare2 = (a.waveDutyPositionSquare2 + 1) % 8
	}

	a.frequencyTimerNoise--
	if a.frequencyTimerNoise <= 0 {
		shift := a.ram[NR43] & 0b1111_0000 >> 4
		divisorCode := a.ram[NR43] & 0b111
		divisor := CODE_TO_DIVISOR[divisorCode]
		a.frequencyTimerNoise = int(divisor) << int(shift)

		width := a.isByteBitSet(NR43, 3)

		xor := (a.lsfr & 1) ^ (a.lsfr & 2 >> 1)
		newLsfr := (a.lsfr>>1)&0b1011_1111_1111_1111 | (xor << 14)
		if width {
			newLsfr = newLsfr&0b1111_1111_1011_111 | (xor << 6)
		}
		a.lsfr = newLsfr
	}
}

func (a *APUImpl) emitSample(sample uint16) {
	low := sample & 0xFF
	high := sample & 0xFF00 >> 8
	a.sampleBuf = append(a.sampleBuf, byte(low), byte(high))
	if len(a.sampleBuf) >= SAMPLE_BUFFER_SIZE {
		a.samples <- a.sampleBuf
		a.sampleBuf = a.sampleBuf[:0]
	}
}

func (a *APUImpl) StepAPU() {

	a.updateState()

	a.cycleCounter++

	if a.cycleCounter%87 > 0 {
		return
	}

	leftSquare1Output, rightSquare1Output := a.getSquare1Output()
	leftSquare2Output, rightSquare2Output := a.getSquare2Output()
	leftNoiseOutput, rightNoiseOutput := a.getNoiseOutput()

	leftVolume := uint16(a.ram[NR50] & 0b111_0000 >> 4)
	rightVolume := uint16(a.ram[NR50] & 0b111)

	left := leftVolume * (uint16(leftSquare1Output) + uint16(leftSquare2Output) + uint16(leftNoiseOutput)) / 3
	right := rightVolume * (uint16(rightSquare1Output) + uint16(rightSquare2Output) + uint16(rightNoiseOutput)) / 3

	a.emitSample(left)
	a.emitSample(right)
}

func (a *APUImpl) Read(p []byte) (n int, err error) {
	select {
	case buf := <-a.samples:
		count := copy(p, buf)
		return count, nil
	default:
		return 0, nil
	}
}
func (a *APUImpl) ToReadCloser() io.ReadCloser { return io.NopCloser(a) }
