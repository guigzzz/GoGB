package backend

import (
	"image"
	"io"
)

type Emulator struct {
	ppu *PPU
	cpu *CPU
	mmu *MMU
	apu *APU
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

func newEmulatorForTests(rom []byte) *Emulator {
	return NewEmulator(rom, false, false)
}

func NewEmulator(rom []byte, debug, enableApu bool) *Emulator {
	return newEmulator(rom, NewNullLogger(), debug, enableApu)
}

func newEmulator(rom []byte, logger Logger, debug, enableApu bool) *Emulator {
	ram := make([]byte, 1<<16)

	apu := NewAPU(ram)
	if !enableApu {
		// stops apu from emitting samples
		// useful to avoid blocking in integ tests because nothing is consuming the samples
		apu.Disable()
	}

	mbc := NewMBC(rom)
	mmu := NewMMU(ram, mbc, logger, apu.AudioRegisterWriteCallback)

	cpu := NewCPU(debug, apu, mmu)
	ppu := NewPPU(ram, cpu)

	return &Emulator{ppu, cpu, mmu, apu}
}
