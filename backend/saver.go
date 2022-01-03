package backend

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
)

const SAVES = "saves"
const JSON = ".json"

func setupSaveDirectory() {

	_, err := os.Stat(SAVES)
	if os.IsNotExist(err) {
		os.Mkdir(SAVES, os.ModePerm)
	}

}

func makeSavePathForRomPath(romPath string) string {
	base := path.Base(filepath.ToSlash(romPath))
	return path.Join(SAVES, base+JSON)
}

func SaveExistsForRom(romPath string) bool {
	// save := filepath.FromSlash(makeSavePathForRomPath(romPath))
	// fmt.Println("Trying load save file: " + save)
	// _, err := os.Stat(save)
	// if os.IsExist(err) {
	// 	fmt.Println("Found save file!")
	// 	return true
	// }

	// return false
	return true
}

func LoadSave(romPath string) (p *PPU, c *CPU) {

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

	c = new(CPU)
	c.reg = cpuState.Reg
	c.SP = cpuState.SP
	c.PC = cpuState.PC
	c.ram = cpuState.Ram
	c.IME = cpuState.IME
	c.mbc = cpuState.Mbc
	c.haltMode = cpuState.HaltMode
	c.cycleCounter = cpuState.CycleCounter

	return NewPPU(c), c
}

type MBCState struct {
}

type CPUState struct {
	Reg [8]byte
	SP  uint16 // stack pointer
	PC  uint16 // program counter
	Ram []byte // 64 KB ram
	IME bool   // interrupt master enable

	Mbc MBC // memory bank controller

	HaltMode     byte
	CycleCounter uint64
}

type EmulatorState struct {
	Cpu CPUState
}

func DumpEmulatorState(romPath string, p *PPU, c *CPU) {
	setupSaveDirectory()

	save := makeSavePathForRomPath(romPath)
	fmt.Println("Writing save file: " + save)

	cpuState := CPUState{}
	cpuState.Reg = c.reg
	cpuState.SP = c.SP
	cpuState.PC = c.PC
	cpuState.Ram = c.ram
	cpuState.IME = c.IME
	cpuState.Mbc = c.mbc
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
