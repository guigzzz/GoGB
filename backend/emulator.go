package backend

import (
	"image"
	"io"
	"os"
)

type Emulator struct {
	ppu *PPU
	cpu *CPU
	mmu *MMU
	apu *APU

	enableApu bool
	logger    Logger
	debug     bool
}

func (e *Emulator) SetKeyIsPressed(key string, isPressed bool) {
	e.mmu.KeyPressedMap[key] = isPressed
}

func (e *Emulator) RunForAFrame() {
	e.ppu.RunEmulatorForAFrame()
}

func (e *Emulator) GetAudioStream() io.ReadCloser {
	return e.apu.ToReadCloser()
}

func (e *Emulator) GetImage() *image.RGBA {
	return e.ppu.Image
}

func WithDisableApu() func(*Emulator) {
	return func(e *Emulator) {
		e.enableApu = false
	}
}

func WithLogger(logger Logger) func(*Emulator) {
	return func(e *Emulator) {
		e.logger = logger
	}
}

func WithDebug(debug bool) func(*Emulator) {
	return func(e *Emulator) {
		e.debug = debug
	}
}

func NewEmulator(path string, options ...func(*Emulator)) *Emulator {
	emu := new(Emulator)
	emu.enableApu = true
	emu.debug = false
	emu.logger = NewNullLogger()

	for _, o := range options {
		o(emu)
	}

	rom, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}

	ram := make([]byte, 1<<16)

	apu := NewAPU(ram)
	if !emu.enableApu {
		// stops apu from emitting samples
		// useful to avoid blocking in integ tests because nothing is consuming the samples
		apu.Disable()
	}

	mbc := NewMBC(rom)
	mmu := NewMMU(ram, mbc, emu.logger, apu.AudioRegisterWriteCallback)

	cpu := NewCPU(emu.debug, apu, mmu)
	ppu := NewPPU(ram, cpu.RunSync)

	emu.ppu = ppu
	emu.cpu = cpu
	emu.apu = apu
	emu.mmu = mmu

	return emu
}
