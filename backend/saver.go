package backend

import (
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
)

const SAVES = "saves"
const EXTENSION = ".json.gz"

func setupSaveDirectory() {

	_, err := os.Stat(SAVES)
	if errors.Is(err, fs.ErrNotExist) {
		os.Mkdir(SAVES, os.ModePerm)
	}

}

func makeSavePathForRomPath(romPath string) string {
	base := path.Base(filepath.ToSlash(romPath))
	return path.Join(SAVES, base+EXTENSION)
}

func SaveExistsForRom(romPath string) bool {
	save := filepath.FromSlash(makeSavePathForRomPath(romPath))

	_, err := os.Stat(save)
	if err == nil {
		fmt.Println("Found save file!")
		return true
	}

	fmt.Println("Save not found.")
	return false
}

func LoadSave(romPath string) *Emulator {

	f, err := os.OpenFile(makeSavePathForRomPath(romPath), os.O_RDONLY, 0644)
	if err != nil {
		panic(err)
	}

	decompress, err := gzip.NewReader(f)
	if err != nil {
		panic(err)
	}

	defer decompress.Close()

	decoder := json.NewDecoder(decompress)

	var state EmulatorState
	err = decoder.Decode(&state)
	if err != nil {
		panic(err)
	}

	logger := NewNullLogger()

	cpuState := state.Cpu

	apu := NewAPU(cpuState.Ram)
	mmu := NewMMU(cpuState.Ram, cpuState.Mbc.mbc, logger, apu.AudioRegisterWriteCallback)

	cpu := new(CPU)
	cpu.reg = cpuState.Reg
	cpu.SP = cpuState.SP
	cpu.PC = cpuState.PC
	cpu.IME = cpuState.IME
	cpu.haltMode = cpuState.HaltMode
	cpu.cycleCounter = cpuState.CycleCounter
	cpu.apu = apu
	cpu.mmu = mmu

	ppu := NewPPU(cpuState.Ram, cpu.RunSync)

	return &Emulator{ppu, cpu, mmu, apu, true, logger, false}
}

type CPUState struct {
	Reg [8]byte
	SP  uint16 // stack pointer
	PC  uint16 // program counter
	Ram []byte // 64 KB ram
	IME bool   // interrupt master enable

	Mbc MbcWrapper // memory bank controller

	HaltMode     byte
	CycleCounter uint64
}

type EmulatorState struct {
	Cpu CPUState
}

func DumpEmulatorState(romPath string, emu *Emulator) {
	setupSaveDirectory()

	save := makeSavePathForRomPath(romPath)
	fmt.Println("Writing save file: " + save)

	c := emu.cpu
	m := emu.mmu

	cpuState := CPUState{}
	cpuState.Reg = c.reg
	cpuState.SP = c.SP
	cpuState.PC = c.PC
	cpuState.Ram = m.ram
	cpuState.IME = c.IME
	cpuState.Mbc = MbcWrapper{m.mbc}
	cpuState.HaltMode = c.haltMode
	cpuState.CycleCounter = c.cycleCounter

	state := EmulatorState{cpuState}

	bytes, err := json.MarshalIndent(state, "", "\t")
	if err != nil {
		fmt.Println("Can't serialize", state)
	}

	f, err := os.OpenFile(save, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}

	compress := gzip.NewWriter(f)
	defer compress.Close()

	_, err = compress.Write(bytes)
	if err != nil {
		panic(err)
	}
}
