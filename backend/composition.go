package backend

func composeForTests(rom []byte) (*PPU, *CPU, *MMU, APU) {
	return Compose(rom, false, false)
}

func Compose(rom []byte, debug, useRealApu bool) (*PPU, *CPU, *MMU, APU) {
	ram := make([]byte, 1<<16)

	var apu APU = nil
	if useRealApu {
		apu = NewAPU(ram)
	} else {
		apu = &NullAPU{}
	}

	mbc := NewMBC(rom)
	mmu := NewMMU(ram, mbc, NewNullLogger(), apu.AudioRegisterWriteCallback)

	cpu := NewCPU(debug, apu, mmu)
	ppu := NewPPU(ram, cpu)

	return ppu, cpu, mmu, apu
}
