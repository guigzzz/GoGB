package backend

import (
	"fmt"
)

// CPU represents the current cpu state
type CPU struct {
	reg [8]byte
	SP  uint16 // stack pointer
	PC  uint16 // program counter
	ram []byte // 64 KB ram
	IME bool   // interrupt master enable

	cartridgeROM        []byte
	cartridgeRAM        []byte
	cartridgeRAMEnabled bool
	ROMMode             bool
	selectedROMBank     byte // points to the currently switched rom bank
	selectedRAMBank     byte // points to the currently switched ram bank
	mbcType             byte // memory bank controller type (0, 1, etc)

	instructionCounter uint // to count instructions
}

// NewCPU creates a new cpu struct
// also copies the bootrom into ram from 0x0000 to 0x00FF (256 bytes)
func NewCPU(rom []byte) *CPU {
	c := new(CPU)

	c.ram = make([]byte, 1<<16)

	// copy first 16KB of data into the ram
	for i := 0; i < (1 << 14); i++ {
		c.ram[i] = rom[i]
	}

	c.cartridgeROM = rom
	c.cartridgeRAM = make([]byte, c.getCartridgeRAMSize())

	mbc := c.ram[0x147]
	if mbc > 1 {
		panic("GoGB currently only supports roms using mbc1")
	}
	c.mbcType = mbc

	c.selectedROMBank = 1
	c.selectedRAMBank = 0

	c.PC = 0x100
	c.SP = 0xFFFE

	return c
}

// NewTestCPU creates a barebone CPU specifically for tests
// designed to be fast
func NewTestCPU() *CPU {
	c := new(CPU)
	c.ram = make([]byte, 1<<16)

	c.selectedROMBank = 1
	c.selectedRAMBank = 0

	c.cartridgeROM = make([]byte, 1<<15)
	c.cartridgeRAM = make([]byte, 1<<12)

	return c
}

func (c *CPU) Runner() {
	for {
		c.CheckAndHandleInterrupts()
		c.DecodeAndExecuteNext()
	}
}

// GetRAM exposes the c.ram member
func (c *CPU) GetRAM() []byte {
	return c.ram
}

func (c *CPU) getCartridgeRAMSize() uint32 {
	switch c.ram[0x0149] {
	case 0:
		return 0
	case 1:
		return 1 << 11
	case 2:
		return 1 << 13
	case 3:
		return 1 << 15
	case 4:
		return 1 << 17
	case 5:
		return 1 << 16
	default:
		panic(fmt.Sprintf("Got unexpected RAM size index: %v", c.ram[0x0149]))
	}
}

func (c *CPU) String() string {
	ret := "CPU registers:\n" +
		fmt.Sprintf("A: 0x%0.2X, F: 0x%0.2X, (AF: 0x%0.4X)\n", c.reg[A], c.reg[F], c.ReadAF()) +
		fmt.Sprintf("B: 0x%0.2X, C: 0x%0.2X, (BC: 0x%0.4X)\n", c.reg[B], c.reg[C], c.ReadBC()) +
		fmt.Sprintf("D: 0x%0.2X, E: 0x%0.2X, (DE: 0x%0.4X)\n", c.reg[D], c.reg[E], c.ReadDE()) +
		fmt.Sprintf("H: 0x%0.2X, L: 0x%0.2X, (HL: 0x%0.4X), (HL): 0x%0.2X\n",
			c.reg[H], c.reg[L], c.ReadHL(), c.readMemory(c.ReadHL())) +
		fmt.Sprintf("SP: 0x%0.4X, PC: 0x%0.4X\n", c.SP, c.PC) +
		fmt.Sprintf("Z: %1b, N: %1b, H: %1b, C: %1b\n",
			c.ReadFlag(ZFlag), c.ReadFlag(NFlag), c.ReadFlag(HFlag), c.ReadFlag(CFlag))

	return ret
}

// DecodeAndExecuteNext fetches next instruction from memory stored at PC
func (c *CPU) DecodeAndExecuteNext() {
	op := c.readMemory(c.PC)
	oprow := (op & 0xF0) >> 4

	b := []byte{c.readMemory(c.PC), c.readMemory(c.PC + 1), c.readMemory(c.PC + 2)}

	switch {
	case oprow <= 3:
		// various instructions
		c.DecodeVariousUpper(b)
	case oprow <= 7:
		// various memory instructions
		c.DecodeMem(b)
	case oprow <= 11:
		// various ALU inctructions
		c.DecodeArith(b)
	default:
		// various instructions
		c.DecodeVariousLower(b)
	}

	c.PC += GetPCIncrement(op)
	c.instructionCounter++
}

// GetPCIncrement return PC increment for opcode
func GetPCIncrement(op byte) uint16 {
	switch op {
	case 0xCB: // CB Prefixed instructions
		return 2
	case 0x38: //{JR C r8 0x38 2 [12]}
		return 2
	case 0xe7: //{RST 20H  0xe7 1 [16]}
		return 0
	case 0xf6: //{OR d8  0xf6 2 [8]}
		return 2
	case 0xde: //{SBC A d8 0xde 2 [8]}
		return 2
	case 0xc2: //{JP NZ a16 0xc2 3 [16]}
		return 0
	case 0xce: //{ADC A d8 0xce 2 [8]}
		return 2
	case 0x16: //{LD D d8 0x16 2 [8]}
		return 2
	case 0xdf: //{RST 18H  0xdf 1 [16]}
		return 0
	case 0xd4: //{CALL NC a16 0xd4 3 [24]}
		return 0
	case 0xe8: //{ADD SP r8 0xe8 2 [16]}
		return 2
	case 0x1: //{LD BC d16 0x1 3 [12]}
		return 3
	case 0xfe: //{CP d8  0xfe 2 [8]}
		return 2
	case 0xe2: //{LD (C) A 0xe2 2 [8]}
		return 1
	case 0xe6: //{AND d8  0xe6 2 [8]}
		return 2
	case 0x31: //{LD SP d16 0x31 3 [12]}
		return 3
	case 0xfa: //{LD A (a16) 0xfa 3 [16]}
		return 3
	case 0x21: //{LD HL d16 0x21 3 [12]}
		return 3
	case 0x26: //{LD H d8 0x26 2 [8]}
		return 2
	case 0x3e: //{LD A d8 0x3e 2 [8]}
		return 2
	case 0x1e: //{LD E d8 0x1e 2 [8]}
		return 2
	case 0xc9: //{RET   0xc9 1 [16]}
		return 0
	case 0xca: //{JP Z a16 0xca 3 [16]}
		return 0
	case 0x10: //{STOP 0  0x10 2 [4]}
		return 2
	case 0xf7: //{RST 30H  0xf7 1 [16]}
		return 0
	case 0xc0: //{RET NZ  0xc0 1 [20]}
		return 0
	case 0x28: //{JR Z r8 0x28 2 [12]}
		return 2
	case 0xc7: //{RST 00H  0xc7 1 [16]}
		return 0
	case 0xd0: //{RET NC  0xd0 1 [20]}
		return 0
	case 0x18: //{JR r8  0x18 2 [12]}
		return 2
	case 0xcf: //{RST 08H  0xcf 1 [16]}
		return 0
	case 0xdc: //{CALL C a16 0xdc 3 [24]}
		return 0
	case 0xd7: //{RST 10H  0xd7 1 [16]}
		return 0
	case 0x36: //{LD (HL) d8 0x36 2 [12]}
		return 2
	case 0xc6: //{ADD A d8 0xc6 2 [8]}
		return 2
	case 0xda: //{JP C a16 0xda 3 [16]}
		return 0
	case 0x2e: //{LD L d8 0x2e 2 [8]}
		return 2
	case 0xcc: //{CALL Z a16 0xcc 3 [24]}
		return 0
	case 0x30: //{JR NC r8 0x30 2 [12]}
		return 2
	case 0xf0: //{LDH A (a8) 0xf0 2 [12]}
		return 2
	case 0xc4: //{CALL NZ a16 0xc4 3 [24]}
		return 0
	case 0x6: //{LD B d8 0x6 2 [8]}
		return 2
	case 0xf8: //{LD HL SP+r8 0xf8 2 [12]}
		return 2
	case 0xe9: //{JP (HL)  0xe9 1 [4]}
		return 0
	case 0xe: //{LD C d8 0xe 2 [8]}
		return 2
	case 0xd2: //{JP NC a16 0xd2 3 [16]}
		return 0
	case 0xd8: //{RET C  0xd8 1 [20]}
		return 0
	case 0xf2: //{LD A (C) 0xf2 2 [8]}
		return 1
	case 0xc3: //{JP a16  0xc3 3 [16]}
		return 0
	case 0xc8: //{RET Z  0xc8 1 [20]}
		return 0
	case 0x8: //{LD (a16) SP 0x8 3 [20]}
		return 3
	case 0xe0: //{LDH (a8) A 0xe0 2 [12]}
		return 2
	case 0xef: //{RST 28H  0xef 1 [16]}
		return 0
	case 0x20: //{JR NZ r8 0x20 2 [12]}
		return 2
	case 0xff: //{RST 38H  0xff 1 [16]}
		return 0
	case 0xd6: //{SUB d8  0xd6 2 [8]}
		return 2
	case 0xee: //{XOR d8  0xee 2 [8]}
		return 2
	case 0x11: //{LD DE d16 0x11 3 [12]}
		return 3
	case 0xcd: //{CALL a16  0xcd 3 [24]}
		return 0
	case 0xea: //{LD (a16) A 0xea 3 [16]}
		return 3
	case 0xd9:
		return 0
	default:
		return 1
	}
}

// FetchCycles returns the number of cycles that the current instruction should be ran for
// This is used to make sure the emulator runs at the correct speed
func (c *CPU) FetchCycles() byte {
	op := c.readMemory(c.PC)
	if op == 0xCB {
		return getCbprefixedCycles(c.ram[c.PC+1])
	}
	return getUnprefixedCycles(op)
}

func getUnprefixedCycles(op byte) byte {

	// Generated by: (Using opcodes.json)
	// for i := 0; i < 256; i++ {
	// 	if v, ok := d.Unprefixed[byte(i)]; ok {
	// 		fmt.Printf("case %v: //%v\nreturn %v\n", v.Addr, v, v.Cycles[0])
	// 	}
	// }

	switch op {
	case 0x0: //{NOP   0x0 1 [4]}
		return 4
	case 0x1: //{LD BC d16 0x1 3 [12]}
		return 12
	case 0x2: //{LD (BC) A 0x2 1 [8]}
		return 8
	case 0x3: //{INC BC  0x3 1 [8]}
		return 8
	case 0x4: //{INC B  0x4 1 [4]}
		return 4
	case 0x5: //{DEC B  0x5 1 [4]}
		return 4
	case 0x6: //{LD B d8 0x6 2 [8]}
		return 8
	case 0x7: //{RLCA   0x7 1 [4]}
		return 4
	case 0x8: //{LD (a16) SP 0x8 3 [20]}
		return 20
	case 0x9: //{ADD HL BC 0x9 1 [8]}
		return 8
	case 0xa: //{LD A (BC) 0xa 1 [8]}
		return 8
	case 0xb: //{DEC BC  0xb 1 [8]}
		return 8
	case 0xc: //{INC C  0xc 1 [4]}
		return 4
	case 0xd: //{DEC C  0xd 1 [4]}
		return 4
	case 0xe: //{LD C d8 0xe 2 [8]}
		return 8
	case 0xf: //{RRCA   0xf 1 [4]}
		return 4
	case 0x10: //{STOP 0  0x10 2 [4]}
		return 4
	case 0x11: //{LD DE d16 0x11 3 [12]}
		return 12
	case 0x12: //{LD (DE) A 0x12 1 [8]}
		return 8
	case 0x13: //{INC DE  0x13 1 [8]}
		return 8
	case 0x14: //{INC D  0x14 1 [4]}
		return 4
	case 0x15: //{DEC D  0x15 1 [4]}
		return 4
	case 0x16: //{LD D d8 0x16 2 [8]}
		return 8
	case 0x17: //{RLA   0x17 1 [4]}
		return 4
	case 0x18: //{JR r8  0x18 2 [12]}
		return 12
	case 0x19: //{ADD HL DE 0x19 1 [8]}
		return 8
	case 0x1a: //{LD A (DE) 0x1a 1 [8]}
		return 8
	case 0x1b: //{DEC DE  0x1b 1 [8]}
		return 8
	case 0x1c: //{INC E  0x1c 1 [4]}
		return 4
	case 0x1d: //{DEC E  0x1d 1 [4]}
		return 4
	case 0x1e: //{LD E d8 0x1e 2 [8]}
		return 8
	case 0x1f: //{RRA   0x1f 1 [4]}
		return 4
	case 0x20: //{JR NZ r8 0x20 2 [12]}
		return 12
	case 0x21: //{LD HL d16 0x21 3 [12]}
		return 12
	case 0x22: //{LD (HL+) A 0x22 1 [8]}
		return 8
	case 0x23: //{INC HL  0x23 1 [8]}
		return 8
	case 0x24: //{INC H  0x24 1 [4]}
		return 4
	case 0x25: //{DEC H  0x25 1 [4]}
		return 4
	case 0x26: //{LD H d8 0x26 2 [8]}
		return 8
	case 0x27: //{DAA   0x27 1 [4]}
		return 4
	case 0x28: //{JR Z r8 0x28 2 [12]}
		return 12
	case 0x29: //{ADD HL HL 0x29 1 [8]}
		return 8
	case 0x2a: //{LD A (HL+) 0x2a 1 [8]}
		return 8
	case 0x2b: //{DEC HL  0x2b 1 [8]}
		return 8
	case 0x2c: //{INC L  0x2c 1 [4]}
		return 4
	case 0x2d: //{DEC L  0x2d 1 [4]}
		return 4
	case 0x2e: //{LD L d8 0x2e 2 [8]}
		return 8
	case 0x2f: //{CPL   0x2f 1 [4]}
		return 4
	case 0x30: //{JR NC r8 0x30 2 [12]}
		return 12
	case 0x31: //{LD SP d16 0x31 3 [12]}
		return 12
	case 0x32: //{LD (HL-) A 0x32 1 [8]}
		return 8
	case 0x33: //{INC SP  0x33 1 [8]}
		return 8
	case 0x34: //{INC (HL)  0x34 1 [12]}
		return 12
	case 0x35: //{DEC (HL)  0x35 1 [12]}
		return 12
	case 0x36: //{LD (HL) d8 0x36 2 [12]}
		return 12
	case 0x37: //{SCF   0x37 1 [4]}
		return 4
	case 0x38: //{JR C r8 0x38 2 [12]}
		return 12
	case 0x39: //{ADD HL SP 0x39 1 [8]}
		return 8
	case 0x3a: //{LD A (HL-) 0x3a 1 [8]}
		return 8
	case 0x3b: //{DEC SP  0x3b 1 [8]}
		return 8
	case 0x3c: //{INC A  0x3c 1 [4]}
		return 4
	case 0x3d: //{DEC A  0x3d 1 [4]}
		return 4
	case 0x3e: //{LD A d8 0x3e 2 [8]}
		return 8
	case 0x3f: //{CCF   0x3f 1 [4]}
		return 4
	case 0x40: //{LD B B 0x40 1 [4]}
		return 4
	case 0x41: //{LD B C 0x41 1 [4]}
		return 4
	case 0x42: //{LD B D 0x42 1 [4]}
		return 4
	case 0x43: //{LD B E 0x43 1 [4]}
		return 4
	case 0x44: //{LD B H 0x44 1 [4]}
		return 4
	case 0x45: //{LD B L 0x45 1 [4]}
		return 4
	case 0x46: //{LD B (HL) 0x46 1 [8]}
		return 8
	case 0x47: //{LD B A 0x47 1 [4]}
		return 4
	case 0x48: //{LD C B 0x48 1 [4]}
		return 4
	case 0x49: //{LD C C 0x49 1 [4]}
		return 4
	case 0x4a: //{LD C D 0x4a 1 [4]}
		return 4
	case 0x4b: //{LD C E 0x4b 1 [4]}
		return 4
	case 0x4c: //{LD C H 0x4c 1 [4]}
		return 4
	case 0x4d: //{LD C L 0x4d 1 [4]}
		return 4
	case 0x4e: //{LD C (HL) 0x4e 1 [8]}
		return 8
	case 0x4f: //{LD C A 0x4f 1 [4]}
		return 4
	case 0x50: //{LD D B 0x50 1 [4]}
		return 4
	case 0x51: //{LD D C 0x51 1 [4]}
		return 4
	case 0x52: //{LD D D 0x52 1 [4]}
		return 4
	case 0x53: //{LD D E 0x53 1 [4]}
		return 4
	case 0x54: //{LD D H 0x54 1 [4]}
		return 4
	case 0x55: //{LD D L 0x55 1 [4]}
		return 4
	case 0x56: //{LD D (HL) 0x56 1 [8]}
		return 8
	case 0x57: //{LD D A 0x57 1 [4]}
		return 4
	case 0x58: //{LD E B 0x58 1 [4]}
		return 4
	case 0x59: //{LD E C 0x59 1 [4]}
		return 4
	case 0x5a: //{LD E D 0x5a 1 [4]}
		return 4
	case 0x5b: //{LD E E 0x5b 1 [4]}
		return 4
	case 0x5c: //{LD E H 0x5c 1 [4]}
		return 4
	case 0x5d: //{LD E L 0x5d 1 [4]}
		return 4
	case 0x5e: //{LD E (HL) 0x5e 1 [8]}
		return 8
	case 0x5f: //{LD E A 0x5f 1 [4]}
		return 4
	case 0x60: //{LD H B 0x60 1 [4]}
		return 4
	case 0x61: //{LD H C 0x61 1 [4]}
		return 4
	case 0x62: //{LD H D 0x62 1 [4]}
		return 4
	case 0x63: //{LD H E 0x63 1 [4]}
		return 4
	case 0x64: //{LD H H 0x64 1 [4]}
		return 4
	case 0x65: //{LD H L 0x65 1 [4]}
		return 4
	case 0x66: //{LD H (HL) 0x66 1 [8]}
		return 8
	case 0x67: //{LD H A 0x67 1 [4]}
		return 4
	case 0x68: //{LD L B 0x68 1 [4]}
		return 4
	case 0x69: //{LD L C 0x69 1 [4]}
		return 4
	case 0x6a: //{LD L D 0x6a 1 [4]}
		return 4
	case 0x6b: //{LD L E 0x6b 1 [4]}
		return 4
	case 0x6c: //{LD L H 0x6c 1 [4]}
		return 4
	case 0x6d: //{LD L L 0x6d 1 [4]}
		return 4
	case 0x6e: //{LD L (HL) 0x6e 1 [8]}
		return 8
	case 0x6f: //{LD L A 0x6f 1 [4]}
		return 4
	case 0x70: //{LD (HL) B 0x70 1 [8]}
		return 8
	case 0x71: //{LD (HL) C 0x71 1 [8]}
		return 8
	case 0x72: //{LD (HL) D 0x72 1 [8]}
		return 8
	case 0x73: //{LD (HL) E 0x73 1 [8]}
		return 8
	case 0x74: //{LD (HL) H 0x74 1 [8]}
		return 8
	case 0x75: //{LD (HL) L 0x75 1 [8]}
		return 8
	case 0x76: //{HALT   0x76 1 [4]}
		return 4
	case 0x77: //{LD (HL) A 0x77 1 [8]}
		return 8
	case 0x78: //{LD A B 0x78 1 [4]}
		return 4
	case 0x79: //{LD A C 0x79 1 [4]}
		return 4
	case 0x7a: //{LD A D 0x7a 1 [4]}
		return 4
	case 0x7b: //{LD A E 0x7b 1 [4]}
		return 4
	case 0x7c: //{LD A H 0x7c 1 [4]}
		return 4
	case 0x7d: //{LD A L 0x7d 1 [4]}
		return 4
	case 0x7e: //{LD A (HL) 0x7e 1 [8]}
		return 8
	case 0x7f: //{LD A A 0x7f 1 [4]}
		return 4
	case 0x80: //{ADD A B 0x80 1 [4]}
		return 4
	case 0x81: //{ADD A C 0x81 1 [4]}
		return 4
	case 0x82: //{ADD A D 0x82 1 [4]}
		return 4
	case 0x83: //{ADD A E 0x83 1 [4]}
		return 4
	case 0x84: //{ADD A H 0x84 1 [4]}
		return 4
	case 0x85: //{ADD A L 0x85 1 [4]}
		return 4
	case 0x86: //{ADD A (HL) 0x86 1 [8]}
		return 8
	case 0x87: //{ADD A A 0x87 1 [4]}
		return 4
	case 0x88: //{ADC A B 0x88 1 [4]}
		return 4
	case 0x89: //{ADC A C 0x89 1 [4]}
		return 4
	case 0x8a: //{ADC A D 0x8a 1 [4]}
		return 4
	case 0x8b: //{ADC A E 0x8b 1 [4]}
		return 4
	case 0x8c: //{ADC A H 0x8c 1 [4]}
		return 4
	case 0x8d: //{ADC A L 0x8d 1 [4]}
		return 4
	case 0x8e: //{ADC A (HL) 0x8e 1 [8]}
		return 8
	case 0x8f: //{ADC A A 0x8f 1 [4]}
		return 4
	case 0x90: //{SUB B  0x90 1 [4]}
		return 4
	case 0x91: //{SUB C  0x91 1 [4]}
		return 4
	case 0x92: //{SUB D  0x92 1 [4]}
		return 4
	case 0x93: //{SUB E  0x93 1 [4]}
		return 4
	case 0x94: //{SUB H  0x94 1 [4]}
		return 4
	case 0x95: //{SUB L  0x95 1 [4]}
		return 4
	case 0x96: //{SUB (HL)  0x96 1 [8]}
		return 8
	case 0x97: //{SUB A  0x97 1 [4]}
		return 4
	case 0x98: //{SBC A B 0x98 1 [4]}
		return 4
	case 0x99: //{SBC A C 0x99 1 [4]}
		return 4
	case 0x9a: //{SBC A D 0x9a 1 [4]}
		return 4
	case 0x9b: //{SBC A E 0x9b 1 [4]}
		return 4
	case 0x9c: //{SBC A H 0x9c 1 [4]}
		return 4
	case 0x9d: //{SBC A L 0x9d 1 [4]}
		return 4
	case 0x9e: //{SBC A (HL) 0x9e 1 [8]}
		return 8
	case 0x9f: //{SBC A A 0x9f 1 [4]}
		return 4
	case 0xa0: //{AND B  0xa0 1 [4]}
		return 4
	case 0xa1: //{AND C  0xa1 1 [4]}
		return 4
	case 0xa2: //{AND D  0xa2 1 [4]}
		return 4
	case 0xa3: //{AND E  0xa3 1 [4]}
		return 4
	case 0xa4: //{AND H  0xa4 1 [4]}
		return 4
	case 0xa5: //{AND L  0xa5 1 [4]}
		return 4
	case 0xa6: //{AND (HL)  0xa6 1 [8]}
		return 8
	case 0xa7: //{AND A  0xa7 1 [4]}
		return 4
	case 0xa8: //{XOR B  0xa8 1 [4]}
		return 4
	case 0xa9: //{XOR C  0xa9 1 [4]}
		return 4
	case 0xaa: //{XOR D  0xaa 1 [4]}
		return 4
	case 0xab: //{XOR E  0xab 1 [4]}
		return 4
	case 0xac: //{XOR H  0xac 1 [4]}
		return 4
	case 0xad: //{XOR L  0xad 1 [4]}
		return 4
	case 0xae: //{XOR (HL)  0xae 1 [8]}
		return 8
	case 0xaf: //{XOR A  0xaf 1 [4]}
		return 4
	case 0xb0: //{OR B  0xb0 1 [4]}
		return 4
	case 0xb1: //{OR C  0xb1 1 [4]}
		return 4
	case 0xb2: //{OR D  0xb2 1 [4]}
		return 4
	case 0xb3: //{OR E  0xb3 1 [4]}
		return 4
	case 0xb4: //{OR H  0xb4 1 [4]}
		return 4
	case 0xb5: //{OR L  0xb5 1 [4]}
		return 4
	case 0xb6: //{OR (HL)  0xb6 1 [8]}
		return 8
	case 0xb7: //{OR A  0xb7 1 [4]}
		return 4
	case 0xb8: //{CP B  0xb8 1 [4]}
		return 4
	case 0xb9: //{CP C  0xb9 1 [4]}
		return 4
	case 0xba: //{CP D  0xba 1 [4]}
		return 4
	case 0xbb: //{CP E  0xbb 1 [4]}
		return 4
	case 0xbc: //{CP H  0xbc 1 [4]}
		return 4
	case 0xbd: //{CP L  0xbd 1 [4]}
		return 4
	case 0xbe: //{CP (HL)  0xbe 1 [8]}
		return 8
	case 0xbf: //{CP A  0xbf 1 [4]}
		return 4
	case 0xc0: //{RET NZ  0xc0 1 [20]}
		return 20
	case 0xc1: //{POP BC  0xc1 1 [12]}
		return 12
	case 0xc2: //{JP NZ a16 0xc2 3 [16]}
		return 16
	case 0xc3: //{JP a16  0xc3 3 [16]}
		return 16
	case 0xc4: //{CALL NZ a16 0xc4 3 [24]}
		return 24
	case 0xc5: //{PUSH BC  0xc5 1 [16]}
		return 16
	case 0xc6: //{ADD A d8 0xc6 2 [8]}
		return 8
	case 0xc7: //{RST 00H  0xc7 1 [16]}
		return 16
	case 0xc8: //{RET Z  0xc8 1 [20]}
		return 20
	case 0xc9: //{RET   0xc9 1 [16]}
		return 16
	case 0xca: //{JP Z a16 0xca 3 [16]}
		return 16
	// case 0xcb: //{PREFIX CB  0xcb 1 [4]}
	// 	return 4
	case 0xcc: //{CALL Z a16 0xcc 3 [24]}
		return 24
	case 0xcd: //{CALL a16  0xcd 3 [24]}
		return 24
	case 0xce: //{ADC A d8 0xce 2 [8]}
		return 8
	case 0xcf: //{RST 08H  0xcf 1 [16]}
		return 16
	case 0xd0: //{RET NC  0xd0 1 [20]}
		return 20
	case 0xd1: //{POP DE  0xd1 1 [12]}
		return 12
	case 0xd2: //{JP NC a16 0xd2 3 [16]}
		return 16
	case 0xd4: //{CALL NC a16 0xd4 3 [24]}
		return 24
	case 0xd5: //{PUSH DE  0xd5 1 [16]}
		return 16
	case 0xd6: //{SUB d8  0xd6 2 [8]}
		return 8
	case 0xd7: //{RST 10H  0xd7 1 [16]}
		return 16
	case 0xd8: //{RET C  0xd8 1 [20]}
		return 20
	case 0xd9: //{RETI   0xd9 1 [16]}
		return 16
	case 0xda: //{JP C a16 0xda 3 [16]}
		return 16
	case 0xdc: //{CALL C a16 0xdc 3 [24]}
		return 24
	case 0xde: //{SBC A d8 0xde 2 [8]}
		return 8
	case 0xdf: //{RST 18H  0xdf 1 [16]}
		return 16
	case 0xe0: //{LDH (a8) A 0xe0 2 [12]}
		return 12
	case 0xe1: //{POP HL  0xe1 1 [12]}
		return 12
	case 0xe2: //{LD (C) A 0xe2 2 [8]}
		return 8
	case 0xe5: //{PUSH HL  0xe5 1 [16]}
		return 16
	case 0xe6: //{AND d8  0xe6 2 [8]}
		return 8
	case 0xe7: //{RST 20H  0xe7 1 [16]}
		return 16
	case 0xe8: //{ADD SP r8 0xe8 2 [16]}
		return 16
	case 0xe9: //{JP (HL)  0xe9 1 [4]}
		return 4
	case 0xea: //{LD (a16) A 0xea 3 [16]}
		return 16
	case 0xee: //{XOR d8  0xee 2 [8]}
		return 8
	case 0xef: //{RST 28H  0xef 1 [16]}
		return 16
	case 0xf0: //{LDH A (a8) 0xf0 2 [12]}
		return 12
	case 0xf1: //{POP AF  0xf1 1 [12]}
		return 12
	case 0xf2: //{LD A (C) 0xf2 2 [8]}
		return 8
	case 0xf3: //{DI   0xf3 1 [4]}
		return 4
	case 0xf5: //{PUSH AF  0xf5 1 [16]}
		return 16
	case 0xf6: //{OR d8  0xf6 2 [8]}
		return 8
	case 0xf7: //{RST 30H  0xf7 1 [16]}
		return 16
	case 0xf8: //{LD HL SP+r8 0xf8 2 [12]}
		return 12
	case 0xf9: //{LD SP HL 0xf9 1 [8]}
		return 8
	case 0xfa: //{LD A (a16) 0xfa 3 [16]}
		return 16
	case 0xfb: //{EI   0xfb 1 [4]}
		return 4
	case 0xfe: //{CP d8  0xfe 2 [8]}
		return 8
	case 0xff: //{RST 38H  0xff 1 [16]}
		return 16
	default:
		panic(fmt.Sprintf("Unprefixed Cycle - got unknown op: %X", op))
	}
}

func getCbprefixedCycles(op byte) byte {

	// Generated by: (using opcodes.json)
	// for i := 0; i < 256; i++ {
	// 	v := d.Cbprefixed[byte(i)]
	// 	fmt.Printf("case %v: //%v\nreturn %v\n", v.Addr, v, v.Cycles[0])
	// }

	switch op {
	case 0x0: //{RLC B  0x0 2 [8]}
		return 8
	case 0x1: //{RLC C  0x1 2 [8]}
		return 8
	case 0x2: //{RLC D  0x2 2 [8]}
		return 8
	case 0x3: //{RLC E  0x3 2 [8]}
		return 8
	case 0x4: //{RLC H  0x4 2 [8]}
		return 8
	case 0x5: //{RLC L  0x5 2 [8]}
		return 8
	case 0x6: //{RLC (HL)  0x6 2 [16]}
		return 16
	case 0x7: //{RLC A  0x7 2 [8]}
		return 8
	case 0x8: //{RRC B  0x8 2 [8]}
		return 8
	case 0x9: //{RRC C  0x9 2 [8]}
		return 8
	case 0xa: //{RRC D  0xa 2 [8]}
		return 8
	case 0xb: //{RRC E  0xb 2 [8]}
		return 8
	case 0xc: //{RRC H  0xc 2 [8]}
		return 8
	case 0xd: //{RRC L  0xd 2 [8]}
		return 8
	case 0xe: //{RRC (HL)  0xe 2 [16]}
		return 16
	case 0xf: //{RRC A  0xf 2 [8]}
		return 8
	case 0x10: //{RL B  0x10 2 [8]}
		return 8
	case 0x11: //{RL C  0x11 2 [8]}
		return 8
	case 0x12: //{RL D  0x12 2 [8]}
		return 8
	case 0x13: //{RL E  0x13 2 [8]}
		return 8
	case 0x14: //{RL H  0x14 2 [8]}
		return 8
	case 0x15: //{RL L  0x15 2 [8]}
		return 8
	case 0x16: //{RL (HL)  0x16 2 [16]}
		return 16
	case 0x17: //{RL A  0x17 2 [8]}
		return 8
	case 0x18: //{RR B  0x18 2 [8]}
		return 8
	case 0x19: //{RR C  0x19 2 [8]}
		return 8
	case 0x1a: //{RR D  0x1a 2 [8]}
		return 8
	case 0x1b: //{RR E  0x1b 2 [8]}
		return 8
	case 0x1c: //{RR H  0x1c 2 [8]}
		return 8
	case 0x1d: //{RR L  0x1d 2 [8]}
		return 8
	case 0x1e: //{RR (HL)  0x1e 2 [16]}
		return 16
	case 0x1f: //{RR A  0x1f 2 [8]}
		return 8
	case 0x20: //{SLA B  0x20 2 [8]}
		return 8
	case 0x21: //{SLA C  0x21 2 [8]}
		return 8
	case 0x22: //{SLA D  0x22 2 [8]}
		return 8
	case 0x23: //{SLA E  0x23 2 [8]}
		return 8
	case 0x24: //{SLA H  0x24 2 [8]}
		return 8
	case 0x25: //{SLA L  0x25 2 [8]}
		return 8
	case 0x26: //{SLA (HL)  0x26 2 [16]}
		return 16
	case 0x27: //{SLA A  0x27 2 [8]}
		return 8
	case 0x28: //{SRA B  0x28 2 [8]}
		return 8
	case 0x29: //{SRA C  0x29 2 [8]}
		return 8
	case 0x2a: //{SRA D  0x2a 2 [8]}
		return 8
	case 0x2b: //{SRA E  0x2b 2 [8]}
		return 8
	case 0x2c: //{SRA H  0x2c 2 [8]}
		return 8
	case 0x2d: //{SRA L  0x2d 2 [8]}
		return 8
	case 0x2e: //{SRA (HL)  0x2e 2 [16]}
		return 16
	case 0x2f: //{SRA A  0x2f 2 [8]}
		return 8
	case 0x30: //{SWAP B  0x30 2 [8]}
		return 8
	case 0x31: //{SWAP C  0x31 2 [8]}
		return 8
	case 0x32: //{SWAP D  0x32 2 [8]}
		return 8
	case 0x33: //{SWAP E  0x33 2 [8]}
		return 8
	case 0x34: //{SWAP H  0x34 2 [8]}
		return 8
	case 0x35: //{SWAP L  0x35 2 [8]}
		return 8
	case 0x36: //{SWAP (HL)  0x36 2 [16]}
		return 16
	case 0x37: //{SWAP A  0x37 2 [8]}
		return 8
	case 0x38: //{SRL B  0x38 2 [8]}
		return 8
	case 0x39: //{SRL C  0x39 2 [8]}
		return 8
	case 0x3a: //{SRL D  0x3a 2 [8]}
		return 8
	case 0x3b: //{SRL E  0x3b 2 [8]}
		return 8
	case 0x3c: //{SRL H  0x3c 2 [8]}
		return 8
	case 0x3d: //{SRL L  0x3d 2 [8]}
		return 8
	case 0x3e: //{SRL (HL)  0x3e 2 [16]}
		return 16
	case 0x3f: //{SRL A  0x3f 2 [8]}
		return 8
	case 0x40: //{BIT 0 B 0x40 2 [8]}
		return 8
	case 0x41: //{BIT 0 C 0x41 2 [8]}
		return 8
	case 0x42: //{BIT 0 D 0x42 2 [8]}
		return 8
	case 0x43: //{BIT 0 E 0x43 2 [8]}
		return 8
	case 0x44: //{BIT 0 H 0x44 2 [8]}
		return 8
	case 0x45: //{BIT 0 L 0x45 2 [8]}
		return 8
	case 0x46: //{BIT 0 (HL) 0x46 2 [16]}
		return 16
	case 0x47: //{BIT 0 A 0x47 2 [8]}
		return 8
	case 0x48: //{BIT 1 B 0x48 2 [8]}
		return 8
	case 0x49: //{BIT 1 C 0x49 2 [8]}
		return 8
	case 0x4a: //{BIT 1 D 0x4a 2 [8]}
		return 8
	case 0x4b: //{BIT 1 E 0x4b 2 [8]}
		return 8
	case 0x4c: //{BIT 1 H 0x4c 2 [8]}
		return 8
	case 0x4d: //{BIT 1 L 0x4d 2 [8]}
		return 8
	case 0x4e: //{BIT 1 (HL) 0x4e 2 [16]}
		return 16
	case 0x4f: //{BIT 1 A 0x4f 2 [8]}
		return 8
	case 0x50: //{BIT 2 B 0x50 2 [8]}
		return 8
	case 0x51: //{BIT 2 C 0x51 2 [8]}
		return 8
	case 0x52: //{BIT 2 D 0x52 2 [8]}
		return 8
	case 0x53: //{BIT 2 E 0x53 2 [8]}
		return 8
	case 0x54: //{BIT 2 H 0x54 2 [8]}
		return 8
	case 0x55: //{BIT 2 L 0x55 2 [8]}
		return 8
	case 0x56: //{BIT 2 (HL) 0x56 2 [16]}
		return 16
	case 0x57: //{BIT 2 A 0x57 2 [8]}
		return 8
	case 0x58: //{BIT 3 B 0x58 2 [8]}
		return 8
	case 0x59: //{BIT 3 C 0x59 2 [8]}
		return 8
	case 0x5a: //{BIT 3 D 0x5a 2 [8]}
		return 8
	case 0x5b: //{BIT 3 E 0x5b 2 [8]}
		return 8
	case 0x5c: //{BIT 3 H 0x5c 2 [8]}
		return 8
	case 0x5d: //{BIT 3 L 0x5d 2 [8]}
		return 8
	case 0x5e: //{BIT 3 (HL) 0x5e 2 [16]}
		return 16
	case 0x5f: //{BIT 3 A 0x5f 2 [8]}
		return 8
	case 0x60: //{BIT 4 B 0x60 2 [8]}
		return 8
	case 0x61: //{BIT 4 C 0x61 2 [8]}
		return 8
	case 0x62: //{BIT 4 D 0x62 2 [8]}
		return 8
	case 0x63: //{BIT 4 E 0x63 2 [8]}
		return 8
	case 0x64: //{BIT 4 H 0x64 2 [8]}
		return 8
	case 0x65: //{BIT 4 L 0x65 2 [8]}
		return 8
	case 0x66: //{BIT 4 (HL) 0x66 2 [16]}
		return 16
	case 0x67: //{BIT 4 A 0x67 2 [8]}
		return 8
	case 0x68: //{BIT 5 B 0x68 2 [8]}
		return 8
	case 0x69: //{BIT 5 C 0x69 2 [8]}
		return 8
	case 0x6a: //{BIT 5 D 0x6a 2 [8]}
		return 8
	case 0x6b: //{BIT 5 E 0x6b 2 [8]}
		return 8
	case 0x6c: //{BIT 5 H 0x6c 2 [8]}
		return 8
	case 0x6d: //{BIT 5 L 0x6d 2 [8]}
		return 8
	case 0x6e: //{BIT 5 (HL) 0x6e 2 [16]}
		return 16
	case 0x6f: //{BIT 5 A 0x6f 2 [8]}
		return 8
	case 0x70: //{BIT 6 B 0x70 2 [8]}
		return 8
	case 0x71: //{BIT 6 C 0x71 2 [8]}
		return 8
	case 0x72: //{BIT 6 D 0x72 2 [8]}
		return 8
	case 0x73: //{BIT 6 E 0x73 2 [8]}
		return 8
	case 0x74: //{BIT 6 H 0x74 2 [8]}
		return 8
	case 0x75: //{BIT 6 L 0x75 2 [8]}
		return 8
	case 0x76: //{BIT 6 (HL) 0x76 2 [16]}
		return 16
	case 0x77: //{BIT 6 A 0x77 2 [8]}
		return 8
	case 0x78: //{BIT 7 B 0x78 2 [8]}
		return 8
	case 0x79: //{BIT 7 C 0x79 2 [8]}
		return 8
	case 0x7a: //{BIT 7 D 0x7a 2 [8]}
		return 8
	case 0x7b: //{BIT 7 E 0x7b 2 [8]}
		return 8
	case 0x7c: //{BIT 7 H 0x7c 2 [8]}
		return 8
	case 0x7d: //{BIT 7 L 0x7d 2 [8]}
		return 8
	case 0x7e: //{BIT 7 (HL) 0x7e 2 [16]}
		return 16
	case 0x7f: //{BIT 7 A 0x7f 2 [8]}
		return 8
	case 0x80: //{RES 0 B 0x80 2 [8]}
		return 8
	case 0x81: //{RES 0 C 0x81 2 [8]}
		return 8
	case 0x82: //{RES 0 D 0x82 2 [8]}
		return 8
	case 0x83: //{RES 0 E 0x83 2 [8]}
		return 8
	case 0x84: //{RES 0 H 0x84 2 [8]}
		return 8
	case 0x85: //{RES 0 L 0x85 2 [8]}
		return 8
	case 0x86: //{RES 0 (HL) 0x86 2 [16]}
		return 16
	case 0x87: //{RES 0 A 0x87 2 [8]}
		return 8
	case 0x88: //{RES 1 B 0x88 2 [8]}
		return 8
	case 0x89: //{RES 1 C 0x89 2 [8]}
		return 8
	case 0x8a: //{RES 1 D 0x8a 2 [8]}
		return 8
	case 0x8b: //{RES 1 E 0x8b 2 [8]}
		return 8
	case 0x8c: //{RES 1 H 0x8c 2 [8]}
		return 8
	case 0x8d: //{RES 1 L 0x8d 2 [8]}
		return 8
	case 0x8e: //{RES 1 (HL) 0x8e 2 [16]}
		return 16
	case 0x8f: //{RES 1 A 0x8f 2 [8]}
		return 8
	case 0x90: //{RES 2 B 0x90 2 [8]}
		return 8
	case 0x91: //{RES 2 C 0x91 2 [8]}
		return 8
	case 0x92: //{RES 2 D 0x92 2 [8]}
		return 8
	case 0x93: //{RES 2 E 0x93 2 [8]}
		return 8
	case 0x94: //{RES 2 H 0x94 2 [8]}
		return 8
	case 0x95: //{RES 2 L 0x95 2 [8]}
		return 8
	case 0x96: //{RES 2 (HL) 0x96 2 [16]}
		return 16
	case 0x97: //{RES 2 A 0x97 2 [8]}
		return 8
	case 0x98: //{RES 3 B 0x98 2 [8]}
		return 8
	case 0x99: //{RES 3 C 0x99 2 [8]}
		return 8
	case 0x9a: //{RES 3 D 0x9a 2 [8]}
		return 8
	case 0x9b: //{RES 3 E 0x9b 2 [8]}
		return 8
	case 0x9c: //{RES 3 H 0x9c 2 [8]}
		return 8
	case 0x9d: //{RES 3 L 0x9d 2 [8]}
		return 8
	case 0x9e: //{RES 3 (HL) 0x9e 2 [16]}
		return 16
	case 0x9f: //{RES 3 A 0x9f 2 [8]}
		return 8
	case 0xa0: //{RES 4 B 0xa0 2 [8]}
		return 8
	case 0xa1: //{RES 4 C 0xa1 2 [8]}
		return 8
	case 0xa2: //{RES 4 D 0xa2 2 [8]}
		return 8
	case 0xa3: //{RES 4 E 0xa3 2 [8]}
		return 8
	case 0xa4: //{RES 4 H 0xa4 2 [8]}
		return 8
	case 0xa5: //{RES 4 L 0xa5 2 [8]}
		return 8
	case 0xa6: //{RES 4 (HL) 0xa6 2 [16]}
		return 16
	case 0xa7: //{RES 4 A 0xa7 2 [8]}
		return 8
	case 0xa8: //{RES 5 B 0xa8 2 [8]}
		return 8
	case 0xa9: //{RES 5 C 0xa9 2 [8]}
		return 8
	case 0xaa: //{RES 5 D 0xaa 2 [8]}
		return 8
	case 0xab: //{RES 5 E 0xab 2 [8]}
		return 8
	case 0xac: //{RES 5 H 0xac 2 [8]}
		return 8
	case 0xad: //{RES 5 L 0xad 2 [8]}
		return 8
	case 0xae: //{RES 5 (HL) 0xae 2 [16]}
		return 16
	case 0xaf: //{RES 5 A 0xaf 2 [8]}
		return 8
	case 0xb0: //{RES 6 B 0xb0 2 [8]}
		return 8
	case 0xb1: //{RES 6 C 0xb1 2 [8]}
		return 8
	case 0xb2: //{RES 6 D 0xb2 2 [8]}
		return 8
	case 0xb3: //{RES 6 E 0xb3 2 [8]}
		return 8
	case 0xb4: //{RES 6 H 0xb4 2 [8]}
		return 8
	case 0xb5: //{RES 6 L 0xb5 2 [8]}
		return 8
	case 0xb6: //{RES 6 (HL) 0xb6 2 [16]}
		return 16
	case 0xb7: //{RES 6 A 0xb7 2 [8]}
		return 8
	case 0xb8: //{RES 7 B 0xb8 2 [8]}
		return 8
	case 0xb9: //{RES 7 C 0xb9 2 [8]}
		return 8
	case 0xba: //{RES 7 D 0xba 2 [8]}
		return 8
	case 0xbb: //{RES 7 E 0xbb 2 [8]}
		return 8
	case 0xbc: //{RES 7 H 0xbc 2 [8]}
		return 8
	case 0xbd: //{RES 7 L 0xbd 2 [8]}
		return 8
	case 0xbe: //{RES 7 (HL) 0xbe 2 [16]}
		return 16
	case 0xbf: //{RES 7 A 0xbf 2 [8]}
		return 8
	case 0xc0: //{SET 0 B 0xc0 2 [8]}
		return 8
	case 0xc1: //{SET 0 C 0xc1 2 [8]}
		return 8
	case 0xc2: //{SET 0 D 0xc2 2 [8]}
		return 8
	case 0xc3: //{SET 0 E 0xc3 2 [8]}
		return 8
	case 0xc4: //{SET 0 H 0xc4 2 [8]}
		return 8
	case 0xc5: //{SET 0 L 0xc5 2 [8]}
		return 8
	case 0xc6: //{SET 0 (HL) 0xc6 2 [16]}
		return 16
	case 0xc7: //{SET 0 A 0xc7 2 [8]}
		return 8
	case 0xc8: //{SET 1 B 0xc8 2 [8]}
		return 8
	case 0xc9: //{SET 1 C 0xc9 2 [8]}
		return 8
	case 0xca: //{SET 1 D 0xca 2 [8]}
		return 8
	case 0xcb: //{SET 1 E 0xcb 2 [8]}
		return 8
	case 0xcc: //{SET 1 H 0xcc 2 [8]}
		return 8
	case 0xcd: //{SET 1 L 0xcd 2 [8]}
		return 8
	case 0xce: //{SET 1 (HL) 0xce 2 [16]}
		return 16
	case 0xcf: //{SET 1 A 0xcf 2 [8]}
		return 8
	case 0xd0: //{SET 2 B 0xd0 2 [8]}
		return 8
	case 0xd1: //{SET 2 C 0xd1 2 [8]}
		return 8
	case 0xd2: //{SET 2 D 0xd2 2 [8]}
		return 8
	case 0xd3: //{SET 2 E 0xd3 2 [8]}
		return 8
	case 0xd4: //{SET 2 H 0xd4 2 [8]}
		return 8
	case 0xd5: //{SET 2 L 0xd5 2 [8]}
		return 8
	case 0xd6: //{SET 2 (HL) 0xd6 2 [16]}
		return 16
	case 0xd7: //{SET 2 A 0xd7 2 [8]}
		return 8
	case 0xd8: //{SET 3 B 0xd8 2 [8]}
		return 8
	case 0xd9: //{SET 3 C 0xd9 2 [8]}
		return 8
	case 0xda: //{SET 3 D 0xda 2 [8]}
		return 8
	case 0xdb: //{SET 3 E 0xdb 2 [8]}
		return 8
	case 0xdc: //{SET 3 H 0xdc 2 [8]}
		return 8
	case 0xdd: //{SET 3 L 0xdd 2 [8]}
		return 8
	case 0xde: //{SET 3 (HL) 0xde 2 [16]}
		return 16
	case 0xdf: //{SET 3 A 0xdf 2 [8]}
		return 8
	case 0xe0: //{SET 4 B 0xe0 2 [8]}
		return 8
	case 0xe1: //{SET 4 C 0xe1 2 [8]}
		return 8
	case 0xe2: //{SET 4 D 0xe2 2 [8]}
		return 8
	case 0xe3: //{SET 4 E 0xe3 2 [8]}
		return 8
	case 0xe4: //{SET 4 H 0xe4 2 [8]}
		return 8
	case 0xe5: //{SET 4 L 0xe5 2 [8]}
		return 8
	case 0xe6: //{SET 4 (HL) 0xe6 2 [16]}
		return 16
	case 0xe7: //{SET 4 A 0xe7 2 [8]}
		return 8
	case 0xe8: //{SET 5 B 0xe8 2 [8]}
		return 8
	case 0xe9: //{SET 5 C 0xe9 2 [8]}
		return 8
	case 0xea: //{SET 5 D 0xea 2 [8]}
		return 8
	case 0xeb: //{SET 5 E 0xeb 2 [8]}
		return 8
	case 0xec: //{SET 5 H 0xec 2 [8]}
		return 8
	case 0xed: //{SET 5 L 0xed 2 [8]}
		return 8
	case 0xee: //{SET 5 (HL) 0xee 2 [16]}
		return 16
	case 0xef: //{SET 5 A 0xef 2 [8]}
		return 8
	case 0xf0: //{SET 6 B 0xf0 2 [8]}
		return 8
	case 0xf1: //{SET 6 C 0xf1 2 [8]}
		return 8
	case 0xf2: //{SET 6 D 0xf2 2 [8]}
		return 8
	case 0xf3: //{SET 6 E 0xf3 2 [8]}
		return 8
	case 0xf4: //{SET 6 H 0xf4 2 [8]}
		return 8
	case 0xf5: //{SET 6 L 0xf5 2 [8]}
		return 8
	case 0xf6: //{SET 6 (HL) 0xf6 2 [16]}
		return 16
	case 0xf7: //{SET 6 A 0xf7 2 [8]}
		return 8
	case 0xf8: //{SET 7 B 0xf8 2 [8]}
		return 8
	case 0xf9: //{SET 7 C 0xf9 2 [8]}
		return 8
	case 0xfa: //{SET 7 D 0xfa 2 [8]}
		return 8
	case 0xfb: //{SET 7 E 0xfb 2 [8]}
		return 8
	case 0xfc: //{SET 7 H 0xfc 2 [8]}
		return 8
	case 0xfd: //{SET 7 L 0xfd 2 [8]}
		return 8
	case 0xfe: //{SET 7 (HL) 0xfe 2 [16]}
		return 16
	case 0xff: //{SET 7 A 0xff 2 [8]}
		return 8
	default:
		panic(fmt.Sprintf("Cbprefixed Cycle - got unknown op: %X", op))
	}
}
