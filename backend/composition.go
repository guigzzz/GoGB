package backend

func composeForTests(rom []byte) (*PPU, *CPU, *MMU, *APU) {
	return Compose(rom, false, false)
}

func Compose(rom []byte, debug, enableApu bool) (*PPU, *CPU, *MMU, *APU) {
	return compose(rom, NewNullLogger(), debug, enableApu)
}

func compose(rom []byte, logger Logger, debug, enableApu bool) (*PPU, *CPU, *MMU, *APU) {
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

	return ppu, cpu, mmu, apu
}
