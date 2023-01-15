package backend

import (
	"image"
	"image/png"
	"io"
	"os"
	"path/filepath"
)

type Emulator struct {
	ppu *PPU
	cpu *CPU
	mbc MBC
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

func WithRom(path string) func(*Emulator) {
	return func(e *Emulator) {
		rom, err := os.ReadFile(path)
		if err != nil {
			panic(err)
		}
		e.mbc = NewMBC(rom)
	}
}

func WithAudio(audio bool) func(*Emulator) {
	return func(e *Emulator) {
		e.enableApu = audio
	}
}

func NewTestMBC() MBC {
	return NewMBC0(make([]byte, 1<<15))
}

func WithNoRom() func(*Emulator) {
	return func(e *Emulator) {
		e.mbc = NewTestMBC()
	}
}

func NewEmulator(options ...func(*Emulator)) *Emulator {
	emu := new(Emulator)
	emu.enableApu = true
	emu.debug = false
	emu.logger = NewNullLogger()

	for _, o := range options {
		o(emu)
	}

	ram := make([]byte, 1<<16)

	apu := NewAPU(ram)
	if !emu.enableApu {
		// stops apu from emitting samples
		// useful to avoid blocking in integ tests because nothing is consuming the samples
		apu.Disable()
	}

	mmu := NewMMU(ram, emu.mbc, emu.logger, apu.AudioRegisterWriteCallback)

	cpu := NewCPU(emu.debug, apu, mmu)
	ppu := NewPPU(ram, cpu.RunSync)

	emu.ppu = ppu
	emu.cpu = cpu
	emu.apu = apu
	emu.mmu = mmu

	return emu
}

func createOutputFile(path string) *os.File {

	folder, _ := filepath.Split(path)
	_, err := os.Stat(folder)
	if os.IsNotExist(err) {
		os.MkdirAll(folder, 0777)
	}

	f, err := os.Create(path)
	if err != nil {
		panic(err)
	}

	return f
}

func (e *Emulator) dumpScreenToPng(path string) {
	f := createOutputFile(path)
	defer f.Close()

	png.Encode(f, e.GetImage())
}
