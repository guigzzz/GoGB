package backend

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
)

const SAVES = "saves"
const JSON = ".json"

func setupSaveDirectory() {

	_, err := os.Stat(SAVES)
	if errors.Is(err, fs.ErrNotExist) {
		os.Mkdir(SAVES, os.ModePerm)
	}

}

func makeSavePathForRomPath(romPath string) string {
	base := path.Base(filepath.ToSlash(romPath))
	return path.Join(SAVES, base+JSON)
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

func LoadSave(romPath string) (*PPU, *CPU, *APU, *MMU) {

	save, err := os.ReadFile(makeSavePathForRomPath(romPath))
	if err != nil {
		panic(err)
	}

	var state EmulatorState
	err = json.Unmarshal(save, &state)
	if err != nil {
		panic(err)
	}

	cpuState := state.Cpu

	apu := NewAPU(cpuState.Ram)
	mmu := NewMMU(cpuState.Ram, cpuState.Mbc.mbc, NewPrintLogger(), apu.AudioRegisterWriteCallback)

	cpu := new(CPU)
	cpu.reg = cpuState.Reg
	cpu.SP = cpuState.SP
	cpu.PC = cpuState.PC
	cpu.IME = cpuState.IME
	cpu.haltMode = cpuState.HaltMode
	cpu.cycleCounter = cpuState.CycleCounter
	cpu.apu = apu
	cpu.mmu = mmu

	ppu := NewPPU(cpuState.Ram, cpu)

	return ppu, cpu, apu, mmu
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

func DumpEmulatorState(romPath string, p *PPU, c *CPU, m *MMU) {
	setupSaveDirectory()

	save := makeSavePathForRomPath(romPath)
	fmt.Println("Writing save file: " + save)

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

	err = os.WriteFile(save, bytes, 0644)
	if err != nil {
		panic(err)
	}
}
