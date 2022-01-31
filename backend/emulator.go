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

func newEmulatorForTests(path string) *Emulator {
	return NewEmulator(path, false, false)
}

func NewEmulator(path string, debug, enableApu bool) *Emulator {
	return newEmulator(path, NewNullLogger(), debug, enableApu)
}

func newEmulator(path string, logger Logger, debug, enableApu bool) *Emulator {
	rom, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}

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
	ppu := NewPPU(ram, cpu.RunSync)

	return &Emulator{ppu, cpu, mmu, apu}
}
