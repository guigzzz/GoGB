package backend

import (
	"fmt"
	"io"
)

type APU interface {
	ToReadCloser() io.ReadCloser
	StepAPU()
	AudioRegisterWriteCallback(addr uint16, oldValue, value byte)
}

type NullAPU struct{}

func (a *NullAPU) StepAPU()                                       {}
func (a *NullAPU) Read(p []byte) (int, error)                     { return 0, nil }
func (a *NullAPU) ToReadCloser() io.ReadCloser                    { return io.NopCloser(a) }
func (a *NullAPU) AudioRegisterWriteCallback(_ uint16, _, _ byte) {}

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
	periodTimerSquare1      int
	currentVolumeSquare1    int
	sweepEnabled            bool
	shadowFrequency         int
	sweepTimer              int

	waveDutyPositionSquare2 int
	frequencyTimerSquare2   int
	lengthTimerSquare2      int
	periodTimerSquare2      int
	currentVolumeSquare2    int

	frequencyTimerWave  int
	positionCounterWave int
	lengthTimerWave     int

	frequencyTimerNoise int
	lengthTimerNoise    int
	lsfr                uint16
	periodTimerNoise    int
	currentVolumeNoise  int

	frameSequencerCounter byte

	sampleBuf []byte
	samples   chan []byte

	// for testing
	emitSamples bool
}

const (
	SAMPLE_BUFFER_SIZE = 1024
)

func NewAPU(c *CPU) *APUImpl {
	apu := new(APUImpl)

	apu.ram = c.ram

	apu.sampleBuf = make([]byte, SAMPLE_BUFFER_SIZE)
	apu.samples = make(chan []byte)

	apu.emitSamples = true

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

	amplitude := WAVE_DUTY_TABLE[duty][a.waveDutyPositionSquare1]
	output := amplitude * byte(a.currentVolumeSquare1)

	if !a.isByteBitSet(NR52, 0) {
		output = 0
	}

	leftEnable := a.isByteBitSet(NR51, 4)
	rightEnable := a.isByteBitSet(NR51, 0)

	leftOutput := boolToNum(leftEnable) * output
	rightOutput := boolToNum(rightEnable) * output
	return leftOutput, rightOutput
}

func (a *APUImpl) getSquare2Output() (byte, byte) {
	duty := a.ram[NR21] & 0b1100_000 >> 6

	amplitude := WAVE_DUTY_TABLE[duty][a.waveDutyPositionSquare2]
	output := amplitude * byte(a.currentVolumeSquare2)

	if !a.isByteBitSet(NR52, 1) {
		output = 0
	}

	leftEnable := a.isByteBitSet(NR51, 5)
	rightEnable := a.isByteBitSet(NR51, 1)

	leftOutput := boolToNum(leftEnable) * output
	rightOutput := boolToNum(rightEnable) * output
	return leftOutput, rightOutput
}

var volumeCodeToShift = [4]byte{4, 0, 1, 2}

func (a *APUImpl) getWaveOutput() (byte, byte) {

	index := a.positionCounterWave / 2

	sample := a.ram[0xFF30+index]
	shift := volumeCodeToShift[a.ram[NR32]&0b110_0000>>5]

	if a.positionCounterWave%2 == 0 {
		sample >>= 4
	}

	output := sample & 0xF >> shift

	if !a.isByteBitSet(NR52, 2) {
		output = 0
	}

	leftEnable := a.isByteBitSet(NR51, 6)
	rightEnable := a.isByteBitSet(NR51, 2)

	leftOutput := boolToNum(leftEnable) * output
	rightOutput := boolToNum(rightEnable) * output
	return leftOutput, rightOutput
}

func (a *APUImpl) getNoiseOutput() (byte, byte) {
	amplitude := byte(^(a.lsfr & 1))
	output := amplitude * byte(a.currentVolumeNoise)

	if !a.isByteBitSet(NR52, 3) {
		output = 0
	}

	leftEnable := a.isByteBitSet(NR51, 7)
	rightEnable := a.isByteBitSet(NR51, 3)

	leftOutput := boolToNum(leftEnable) * output
	rightOutput := boolToNum(rightEnable) * output
	return leftOutput, rightOutput
}

func (a *APUImpl) AudioRegisterWriteCallback(addr uint16, oldValue, value byte) {

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
		isTrigger := value&0x80 > 0
		isLength := value&0x40 > 0
		lengthToggled := isLength && oldValue&0x40 == 0
		notLengthClock := a.frameSequencerCounter%2 == 1

		// Extra length clocking occurs when writing to NRx4 when the frame sequencer's
		// next step is one that doesn't clock the length counter. In this case, if the
		// length counter was PREVIOUSLY disabled and now enabled and the length counter
		// is not zero, it is decremented. If this decrement makes it zero and trigger is clear,
		// the channel is disabled.
		if lengthToggled && notLengthClock && a.lengthTimerSquare1 > 0 {
			a.lengthTimerSquare1--
			if a.lengthTimerSquare1 == 0 && !isTrigger {
				a.clearBit(NR52, 0)
			}
		}

		// If a channel is triggered when the frame sequencer's next step is one
		// that doesn't clock the length counter and the length counter is now enabled
		// and length is being set to 64 (256 for wave channel) because it was previously
		// zero, it is set to 63 instead (255 for wave channel).
		if isTrigger && a.lengthTimerSquare1 == 0 {
			a.lengthTimerSquare1 = 64
			if isLength && notLengthClock {
				a.lengthTimerSquare1--
			}
		}

		dacEnabled := a.ram[NR12]&0xF8 > 0
		if isTrigger && dacEnabled {
			a.setBit(NR52, 0)
		}

		if isTrigger {
			a.periodTimerSquare1 = int(a.ram[NR12] & 0b111)
			a.currentVolumeSquare1 = int(a.ram[NR12] >> 4)
		}

		lsb := a.ram[NR13]
		msb := a.ram[NR14] & 0b111
		frequency := uint16(msb)<<8 | uint16(lsb)
		a.shadowFrequency = int(frequency)
		reg := a.ram[NR10]
		period := reg & 0b111_0000 >> 4
		if period == 0 {
			a.sweepTimer = 8
		} else {
			a.sweepTimer = int(period)
		}

		shift := reg & 0b111
		a.sweepEnabled = period > 0 || shift > 0

		if shift > 0 {
			negate := reg&0b1000 > 0
			a.sweepComputeNewFrequency(int(shift), negate)
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
		isTrigger := value&0x80 > 0
		isLength := value&0x40 > 0
		lengthToggled := isLength && oldValue&0x40 == 0
		notLengthClock := a.frameSequencerCounter%2 == 1

		// Extra length clocking occurs when writing to NRx4 when the frame sequencer's
		// next step is one that doesn't clock the length counter. In this case, if the
		// length counter was PREVIOUSLY disabled and now enabled and the length counter
		// is not zero, it is decremented. If this decrement makes it zero and trigger is clear,
		// the channel is disabled.
		if lengthToggled && notLengthClock && a.lengthTimerSquare2 > 0 {
			a.lengthTimerSquare2--
			if a.lengthTimerSquare2 == 0 && !isTrigger {
				a.clearBit(NR52, 1)
			}
		}

		// If a channel is triggered when the frame sequencer's next step is one
		// that doesn't clock the length counter and the length counter is now enabled
		// and length is being set to 64 (256 for wave channel) because it was previously
		// zero, it is set to 63 instead (255 for wave channel).
		if isTrigger && a.lengthTimerSquare2 == 0 {
			a.lengthTimerSquare2 = 64
			if isLength && notLengthClock {
				a.lengthTimerSquare2--
			}
		}

		dacEnabled := a.ram[NR22]&0xF8 > 0
		if isTrigger && dacEnabled {
			a.setBit(NR52, 1)
		}

		if isTrigger {
			a.periodTimerSquare2 = int(a.ram[NR22] & 0b111)
			a.currentVolumeSquare2 = int(a.ram[NR22] >> 4)
		}
	case NR30:
		dacDisabled := value&0x80 == 0
		if dacDisabled {
			a.clearBit(NR52, 2)
		}
	case NR31:
		a.lengthTimerWave = 256 - int(a.ram[NR31])
	case NR34:
		isTrigger := value&0x80 > 0
		isLength := value&0x40 > 0
		lengthToggled := isLength && oldValue&0x40 == 0
		notLengthClock := a.frameSequencerCounter%2 == 1

		// Extra length clocking occurs when writing to NRx4 when the frame sequencer's
		// next step is one that doesn't clock the length counter. In this case, if the
		// length counter was PREVIOUSLY disabled and now enabled and the length counter
		// is not zero, it is decremented. If this decrement makes it zero and trigger is clear,
		// the channel is disabled.
		if lengthToggled && notLengthClock && a.lengthTimerWave > 0 {
			a.lengthTimerWave--
			if a.lengthTimerWave == 0 && !isTrigger {
				a.clearBit(NR52, 2)
			}
		}

		// If a channel is triggered when the frame sequencer's next step is one
		// that doesn't clock the length counter and the length counter is now enabled
		// and length is being set to 64 (256 for wave channel) because it was previously
		// zero, it is set to 63 instead (255 for wave channel).
		if isTrigger && a.lengthTimerWave == 0 {
			a.lengthTimerWave = 256
			if isLength && notLengthClock {
				a.lengthTimerWave--
			}
		}

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
		isTrigger := value&0x80 > 0
		isLength := value&0x40 > 0
		lengthToggled := isLength && oldValue&0x40 == 0
		notLengthClock := a.frameSequencerCounter%2 == 1

		// Extra length clocking occurs when writing to NRx4 when the frame sequencer's
		// next step is one that doesn't clock the length counter. In this case, if the
		// length counter was PREVIOUSLY disabled and now enabled and the length counter
		// is not zero, it is decremented. If this decrement makes it zero and trigger is clear,
		// the channel is disabled.
		if lengthToggled && notLengthClock && a.lengthTimerNoise > 0 {
			a.lengthTimerNoise--
			if a.lengthTimerNoise == 0 && !isTrigger {
				a.clearBit(NR52, 3)
			}
		}

		// If a channel is triggered when the frame sequencer's next step is one
		// that doesn't clock the length counter and the length counter is now enabled
		// and length is being set to 64 (256 for wave channel) because it was previously
		// zero, it is set to 63 instead (255 for wave channel).
		if isTrigger && a.lengthTimerNoise == 0 {
			a.lengthTimerNoise = 64
			if isLength && notLengthClock {
				a.lengthTimerNoise--
			}
		}

		dacEnabled := a.ram[NR42]&0xF8 > 0
		if isTrigger && dacEnabled {
			a.setBit(NR52, 3)
		}

		if isTrigger {
			a.periodTimerNoise = int(a.ram[NR42] & 0b111)
			a.currentVolumeNoise = int(a.ram[NR42] >> 4)
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

func (a *APUImpl) sweepComputeNewFrequency(shift int, negate bool) int {
	newFreq := a.shadowFrequency >> int(shift)
	if negate {
		newFreq = a.shadowFrequency - newFreq
	} else {
		newFreq = a.shadowFrequency + newFreq
	}

	if newFreq > 2047 {
		a.clearBit(NR52, 0)
	}

	return newFreq
}

func (a *APUImpl) updateSweep() {
	if a.sweepTimer > 0 {
		a.sweepTimer--
	}
	if a.sweepTimer == 0 {
		reg := a.ram[NR10]
		period := reg & 0b111_0000 >> 4
		if period == 0 {
			a.sweepTimer = 8
		} else {
			a.sweepTimer = int(period)
		}

		shift := reg & 0b111
		negate := reg&0b1000 > 0
		if a.sweepEnabled && period > 0 {
			newFreq := a.sweepComputeNewFrequency(int(shift), negate)

			if newFreq <= 2047 && shift > 0 {
				a.ram[NR13] = byte(newFreq)
				a.ram[NR14] = byte(newFreq & 0b111_0000_0000 >> 8)

				a.shadowFrequency = newFreq
			}

			// for overflow check
			a.sweepComputeNewFrequency(int(shift), negate)
		}

	}
}

func (a *APUImpl) updateVolumeEnvelope() {

	if a.ram[NR12]&0b111 > 0 {
		if a.periodTimerSquare1 > 0 {
			a.periodTimerSquare1--
		}

		if a.periodTimerSquare1 == 0 {
			a.periodTimerSquare1 = int(a.ram[NR12] & 0b111)

			isAdd := a.ram[NR12]&0x8 > 0
			if isAdd && a.currentVolumeSquare1 < 0xF {
				a.currentVolumeSquare1++
			} else if !isAdd && a.currentVolumeSquare1 > 0 {
				a.currentVolumeSquare1--
			}
		}
	}

	if a.ram[NR22]&0b111 > 0 {
		if a.periodTimerSquare2 > 0 {
			a.periodTimerSquare2--
		}

		if a.periodTimerSquare2 == 0 {
			a.periodTimerSquare2 = int(a.ram[NR22] & 0b111)

			isAdd := a.ram[NR12]&0x8 > 0
			if isAdd && a.currentVolumeSquare2 < 0xF {
				a.currentVolumeSquare2++
			} else if !isAdd && a.currentVolumeSquare2 > 0 {
				a.currentVolumeSquare2--
			}
		}
	}

	if a.ram[NR42]&0b111 > 0 {
		if a.periodTimerNoise > 0 {
			a.periodTimerNoise--
		}

		if a.periodTimerNoise == 0 {
			a.periodTimerNoise = int(a.ram[NR42] & 0b111)

			isAdd := a.ram[NR12]&0x8 > 0
			if isAdd && a.currentVolumeNoise < 0xF {
				a.currentVolumeNoise++
			} else if !isAdd && a.currentVolumeNoise > 0 {
				a.currentVolumeNoise--
			}
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

	if a.frameSequencerCounter%8 == 7 {
		a.updateVolumeEnvelope()
	}

	modFour := a.frameSequencerCounter % 4
	if modFour == 2 || modFour == 6 {
		a.updateSweep()
	}

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

	a.frequencyTimerWave--
	if a.frequencyTimerWave <= 0 {
		lsb := a.ram[NR33]
		msb := a.ram[NR34] & 0b111
		frequency := uint16(msb)<<8 | uint16(lsb)
		a.frequencyTimerWave = int((2048 - frequency) * 2)
		a.positionCounterWave = (a.positionCounterWave + 1) % 64
	}

	a.frequencyTimerNoise--
	if a.frequencyTimerNoise <= 0 {
		shift := a.ram[NR43] & 0b1111_0000 >> 4
		divisorCode := a.ram[NR43] & 0b111
		divisor := CODE_TO_DIVISOR[divisorCode]
		a.frequencyTimerNoise = int(divisor) << int(shift)

		xor := (a.lsfr & 1) ^ (a.lsfr & 2 >> 1)
		newLsfr := (a.lsfr >> 1) | (xor << 14)

		width := a.isByteBitSet(NR43, 3)
		if width {
			newLsfr = newLsfr&0b1111_1111_1011_1111 | (xor << 6)
		}
		a.lsfr = newLsfr & 0x7FFF
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
	leftWaveOutput, rightWaveOutput := a.getWaveOutput()
	leftNoiseOutput, rightNoiseOutput := a.getNoiseOutput()

	leftVolume := uint16(a.ram[NR50] & 0b111_0000 >> 4)
	rightVolume := uint16(a.ram[NR50] & 0b111)

	left := leftVolume * (uint16(leftSquare1Output) + uint16(leftSquare2Output) + uint16(leftWaveOutput) + uint16(leftNoiseOutput))
	right := rightVolume * (uint16(rightSquare1Output) + uint16(rightSquare2Output) + uint16(rightWaveOutput) + uint16(rightNoiseOutput))

	if a.emitSamples {
		a.emitSample(left)
		a.emitSample(right)
	}
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
