package backend

// opcode grid for reference
// http://www.pastraiser.com/cpu/gameboy/gameboy_opcodes.html

// opcode explanations
// http://www.chrisantonellis.com/files/gameboy/gb-instructions.txt

//////////////////
// VariousUpper //
//////////////////

// DecodeVariousUpper decodes various instructions on first 4 rows
func (c *CPU) DecodeVariousUpper(first, second, third byte) (pcIncrement, cycleIncrement int) {
	op := first
	oprow := int((op & 0xF0) >> 4)
	opcol := int(op & 0xF)

	switch opcol {
	case 0:
		switch oprow {
		case 0: // NOP
			// panic("No Op - Unimplemented")
			return 1, 4
		case 1: // STOP
			// fmt.Println("Warning: Got STOP instruction, this is probably a bug")
			return 2, 4
		case 2: // JR NZ,r8
			return c.JumpRelativeNZ(second)
		case 3: // JR NC,r8
			return c.JumpRelativeNC(second)
		}
	case 1:
		v := PackBytes(third, second)
		switch oprow {
		case 0: // LD BC,d16
			c.Writedouble(B, C, v)
		case 1: // LD DE, d16
			c.Writedouble(D, E, v)
		case 2: // LD HL, d16
			c.Writedouble(H, L, v)
		case 3: // LD SP, d16
			c.SP = v
		}
		return 3, 12
	case 2:
		switch oprow {
		case 0: // LD (BC), A
			c.writeMemory(c.ReadBC(), c.reg[A])
		case 1: // LD (DE), A
			c.writeMemory(c.ReadDE(), c.reg[A])
		case 2: // LD (HL+), A
			c.writeMemory(c.ReadHL(), c.reg[A])
			c.IncRegs(H, L)
		case 3: // LD (HL-), A
			c.writeMemory(c.ReadHL(), c.reg[A])
			c.DecRegs(H, L)
		}
		return 1, 8
	case 3:
		switch oprow {
		case 0: // INC BC
			c.IncRegs(B, C)
		case 1: // INC DE
			c.IncRegs(D, E)
		case 2: // INC HL
			c.IncRegs(H, L)
		case 3: // INC SP
			c.IncSP()
		}
		return 1, 8
	case 4:
		switch oprow {
		case 0: // INC B
			c.Inc(B)
			return 1, 4
		case 1: // INC D
			c.Inc(D)
			return 1, 4
		case 2: // INC H
			c.Inc(H)
			return 1, 4
		case 3: // INC (HL)
			c.IncHL()
			return 1, 12
		}

	case 5:
		switch oprow {
		case 0: // DEC B
			c.Dec(B)
			return 1, 4
		case 1: // DEC D
			c.Dec(D)
			return 1, 4
		case 2: // DEC H
			c.Dec(H)
			return 1, 4
		case 3: // DEC (HL)
			c.DecHL()
			return 1, 12
		}
	case 6:
		switch oprow {
		case 0: // LD B, d8
			c.Load(B, second)
			return 2, 8
		case 1: // LD D, d8
			c.Load(D, second)
			return 2, 8
		case 2: // LD H, d8
			c.Load(H, second)
			return 2, 8
		case 3: // LD (HL), d8
			c.StoreN(second)
			return 2, 12
		}
	case 7:
		switch oprow {
		case 0: // RLCA
			c.RotateLeftCReg(A)
			c.ResetFlag(ZFlag)
		case 1: // RLA
			c.RotateLeftReg(A)
			c.ResetFlag(ZFlag)
		case 2: // DAA
			c.DAA()
		case 3: // SCF
			c.ResetFlag(NFlag)
			c.ResetFlag(HFlag)
			c.SetFlag(CFlag)
		}
		return 1, 4
	case 8:
		switch oprow {
		case 0: // LD (a16),SP
			c.StoreSPNN(PackBytes(third, second))
			return 3, 20
		case 1: // JR r8
			c.JumpRelative(second)
			return 2, 12
		case 2: // JR Z,r8
			return c.JumpRelativeZ(second)
		case 3: // JR C,r8
			return c.JumpRelativeC(second)
		}
	case 9:
		switch oprow {
		case 0: // ADD HL, BC
			c.AddHL16(c.ReadBC())
		case 1: // ADD HL, DE
			c.AddHL16(c.ReadDE())
		case 2: // ADD HL, HL
			c.AddHL16(c.ReadHL())
		case 3: // ADD HL, SP
			c.AddHL16(c.SP)
		}
		return 1, 8
	case 10:
		switch oprow {
		case 0: // LD A, (BC)
			c.Load(A, c.readMemory(c.ReadBC()))
		case 1: // LD A, (DE)
			c.Load(A, c.readMemory(c.ReadDE()))
		case 2: // LD A, (HL+)
			c.Load(A, c.readMemory(c.ReadHL()))
			c.IncRegs(H, L)
		case 3: // LD A, (HL-)
			c.Load(A, c.readMemory(c.ReadHL()))
			c.DecRegs(H, L)
		}
		return 1, 8
	case 11:
		switch oprow {
		case 0: // DEC BC
			c.DecRegs(B, C)
		case 1: // DEC DE
			c.DecRegs(D, E)
		case 2: // DEC HL
			c.DecRegs(H, L)
		case 3: // DEC SP
			c.DecSP()
		}
		return 1, 8
	case 12:
		switch oprow {
		case 0: // INC C
			c.Inc(C)
		case 1: // INC E
			c.Inc(E)
		case 2: // INC L
			c.Inc(L)
		case 3: // INC A
			c.Inc(A)
		}
		return 1, 4
	case 13:
		switch oprow {
		case 0: // DEC C
			c.Dec(C)
		case 1: // DEC E
			c.Dec(E)
		case 2: // DEC L
			c.Dec(L)
		case 3: // DEC A
			c.Dec(A)
		}
		return 1, 4
	case 14:
		switch oprow {
		case 0: // LD C, d8
			c.Load(C, second)
		case 1: // LD E, d8
			c.Load(E, second)
		case 2: // LD L, d8
			c.Load(L, second)
		case 3: // LD A, d8
			c.Load(A, second)
		}
		return 2, 8
	case 15:
		switch oprow {
		case 0: // RRCA = RRC A
			c.RotateRightCReg(A)
			c.ResetFlag(ZFlag)
		case 1: // RRA = RR A
			c.RotateRightReg(A)
			c.ResetFlag(ZFlag)
		case 2: // CPL = complement A = ~A
			c.reg[A] = ^c.reg[A]
			c.SetFlag(NFlag)
			c.SetFlag(HFlag)
		case 3: // CCF = complement carry flag = ~CFlag
			if c.IsFlagSet(CFlag) {
				c.ResetFlag(CFlag)
			} else {
				c.SetFlag(CFlag)
			}
			c.ResetFlag(NFlag)
			c.ResetFlag(HFlag)
		}
		return 1, 4
	}

	panic("unreachable")
}

////////////
// Memory //
////////////

// DecodeMem decodes the various LD instructioons
func (c *CPU) DecodeMem(op byte) (cycleIncrement int) {
	oprow := int((op&0xF0)>>4) - 4
	opcol := int(op & 0xF)

	switch opcol {

	case 0, 1, 2, 3, 4, 5:
		return c.decodeLDArg1(oprow, Register(opcol))
	case 6:
		switch oprow {
		case 0:
			c.LoadHL(B)
			return 8
		case 1:
			c.LoadHL(D)
			return 8
		case 2:
			c.LoadHL(H)
			return 8
		case 3:
			c.halt()
			return 4
		}
	case 7:
		return c.decodeLDArg1(oprow, A)

	case 8, 9, 10, 11, 12, 13:
		return c.decodeLDArg2(oprow, Register(opcol-8))
	case 14:
		switch oprow {
		case 0:
			c.LoadHL(C)
		case 1:
			c.LoadHL(E)
		case 2:
			c.LoadHL(L)
		case 3:
			c.LoadHL(A)
		}
		return 8
	case 15:
		return c.decodeLDArg2(oprow, A)
	}

	panic("unreachable")
}

func (c *CPU) decodeLDArg1(oprow int, src Register) int {
	switch oprow {
	case 0:
		c.LoadReg(B, src)
		return 4
	case 1:
		c.LoadReg(D, src)
		return 4
	case 2:
		c.LoadReg(H, src)
		return 4
	case 3:
		c.StoreReg(src)
		return 8
	}

	panic("unreachable")
}

func (c *CPU) decodeLDArg2(oprow int, src Register) int {
	switch oprow {
	case 0:
		c.LoadReg(C, src)
	case 1:
		c.LoadReg(E, src)
	case 2:
		c.LoadReg(L, src)
	case 3:
		c.LoadReg(A, src)
	}
	return 4
}

////////////////
// Arithmetic //
////////////////

// DecodeArith decodes various arithmetic instructions
func (c *CPU) DecodeArith(op byte) (cycleIncrement int) {
	oprow := int((op&0xF0)>>4) - 8
	opcol := int(op & 0xF)

	switch opcol {

	// first half
	case 0, 1, 2, 3, 4, 5:
		return c.decodeArithArg1(oprow, Register(opcol))
	case 6:
		switch oprow {
		case 0:
			c.AddHL(false)
		case 1:
			c.SubHL(false)
		case 2:
			c.AndHL()
		case 3:
			c.OrHL()
		}
		return 8
	case 7:
		return c.decodeArithArg1(oprow, A)

	// second half
	case 8, 9, 10, 11, 12, 13:
		return c.decodeArithArg2(oprow, Register(opcol-8))
	case 14:
		switch oprow {
		case 0:
			c.AddHL(true)
		case 1:
			c.SubHL(true)
		case 2:
			c.XorHL()
		case 3:
			c.CpHL()
		}
		return 8
	case 15:
		return c.decodeArithArg2(oprow, A)
	}

	panic("unreachable")
}

func (c *CPU) decodeArithArg1(oprow int, src Register) (cycleIncrement int) {
	switch oprow {
	case 0:
		c.AddReg(src, false)
	case 1:
		c.SubReg(src, false)
	case 2:
		c.AndReg(src)
	case 3:
		c.OrReg(src)
	}
	return 4
}

func (c *CPU) decodeArithArg2(oprow int, src Register) (cycleIncrement int) {
	switch oprow {
	case 0:
		c.AddReg(src, true)
	case 1:
		c.SubReg(src, true)
	case 2:
		c.XorReg(src)
	case 3:
		c.CpReg(src)
	}
	return 4
}

//////////////////
// VariousLower //
//////////////////

// DecodeVariousLower decodes various instructions on last 4 rows
func (c *CPU) DecodeVariousLower(first, second, third byte) (pcIncrement, cycleIncrement int) {
	op := first
	oprow := int((op&0xF0)>>4) - 12
	opcol := int(op & 0xF)

	switch opcol {
	case 0:
		switch oprow {
		case 0: // RET NZ
			return c.RetNZ()
		case 1: // RET NC
			return c.RetNC()
		case 2: // LDH (a8),A
			c.StoreHigh(second)
			return 2, 12
		case 3: // LDH A,(a8)
			c.LoadHigh(second)
			return 2, 12
		}
	case 1:
		switch oprow {
		case 0: // POP BC
			c.popDouble(B, C)
		case 1: // POP DE
			c.popDouble(D, E)
		case 2: // POP HL
			c.popDouble(H, L)
		case 3: // POP AF
			c.popDouble(A, F)
			c.reg[F] &= 0xF0
		}
		return 1, 12
	case 2:
		switch oprow {
		case 0: // JP NZ,a16
			v := PackBytes(third, second)
			return c.JumpNZ(v)
		case 1: // JP NC,a16
			v := PackBytes(third, second)
			return c.JumpNC(v)
		case 2: // LD (C),A
			c.StoreHigh(c.reg[C])
			return 1, 8
		case 3: // LD A,(C)
			c.LoadHigh(c.reg[C])
			return 1, 8
		}
	case 3:
		switch oprow {
		case 0: // JP a16
			v := PackBytes(third, second)
			c.Jump(v)
			return 0, 16
		case 1, 2: // NONE
			panic("ERROR - byte decoded to unused instruction -> there is a bug somewhere")
		case 3: // DI
			c.IME = false
			return 1, 4
		}
	case 4:
		v := PackBytes(third, second)
		switch oprow {
		case 0: // CALL NZ,a16
			return c.CallNZ(v)
		case 1: // CALL NC,a16
			return c.CallNC(v)
		case 2, 3: // NONE
			panic("ERROR - byte decoded to unused instruction -> there is a bug somewhere")
		}
	case 5:
		switch oprow {
		case 0: // PUSH BC
			c.pushDouble(B, C)
		case 1: // PUSH DE
			c.pushDouble(D, E)
		case 2: // PUSH HL
			c.pushDouble(H, L)
		case 3: // PUSH AF
			c.pushDouble(A, F)
		}
		return 1, 16
	case 6:
		switch oprow {
		case 0: // ADD A,d8
			c.Addn(second, false)
		case 1: // SUB d8
			c.Subn(second, false)
		case 2: // AND d8
			c.Andn(second)
		case 3: // OR d8
			c.Orn(second)
		}
		return 2, 8
	case 7:
		switch oprow {
		case 0: // RST 00H
			c.Rst(0x00)
		case 1: // RST 10H
			c.Rst(0x10)
		case 2: // RST 20H
			c.Rst(0x20)
		case 3: // RST 30H
			c.Rst(0x30)
		}
		return 0, 16
	case 8:
		switch oprow {
		case 0: // RET Z
			return c.RetZ()
		case 1: // RET C
			return c.RetC()
		case 2: // ADD SP,r8
			c.AddSP8(second)
			return 2, 16
		case 3: // LD HL,SP+r8
			c.LoadHLSPN(second)
			return 2, 12
		}
	case 9:
		switch oprow {
		case 0: // RET
			c.Ret()
			return 0, 16
		case 1: // RETI
			c.IME = true
			c.Ret()
			return 0, 16
		case 2: // JP (HL)
			c.Jump(c.ReadHL())
			return 0, 4
		case 3: // LD SP,HL
			c.SP = c.ReadHL()
			return 1, 8
		}
	case 10:
		v := PackBytes(third, second)
		switch oprow {
		case 0: // JP Z,a16
			return c.JumpZ(v)
		case 1: // JP C,a16
			return c.JumpC(v)
		case 2: // LD (a16),A
			c.writeMemory(v, c.reg[A])
			return 3, 16
		case 3: // LD A,(a16)
			c.reg[A] = c.readMemory(v)
			return 3, 16
		}
	case 11:
		switch oprow {
		case 0: // PREFIX CB
			cycles := c.DecodePrefixCB(second)
			return 2, cycles
		case 1, 2: // NONE
			panic("ERROR - byte decoded to unused instruction -> there is a bug somewhere")
		case 3: // EI
			c.IME = true
			return 1, 4
		}
	case 12:
		v := PackBytes(third, second)
		switch oprow {
		case 0: // CALL Z,a16
			return c.CallZ(v)
		case 1: // CALL C,a16
			return c.CallC(v)
		case 2, 3: // NONE
			panic("ERROR - byte decoded to unused instruction -> there is a bug somewhere")
		}
	case 13:
		switch oprow {
		case 0: // CALL a16
			v := PackBytes(third, second)
			c.Call(v)
			return 0, 24
		case 1, 2, 3: // NONE
			panic("ERROR - byte decoded to unused instruction -> there is a bug somewhere")
		}
	case 14:
		switch oprow {
		case 0: // ADC A,d8
			c.Addn(second, true)
		case 1: // SBC A,d8
			c.Subn(second, true)
		case 2: // XOR d8
			c.Xorn(second)
		case 3: // CP d8
			c.Cpn(second)
		}
		return 2, 8
	case 15:
		switch oprow {
		case 0: // RST 08H
			c.Rst(0x08)
		case 1: // RST 18H
			c.Rst(0x18)
		case 2: // RST 28H
			c.Rst(0x28)
		case 3: // RST 38H
			c.Rst(0x38)
		}
		return 0, 16
	}

	panic("unreachable")
}

// DecodePrefixCB decodes instructions prefixed with CB such as bit, res, set, etc.
// returns the number of CPU cycles we ran for
func (c *CPU) DecodePrefixCB(op byte) int {
	oprow := int((op & 0xF0) >> 4)
	opcol := int(op & 0xF)

	switch opcol {
	case 0, 1, 2, 3, 4, 5: // B, C, D, E, H, L
		return c.decodePrefixCBRow1(oprow, Register(opcol))
	case 6: // (HL)
		return c.decodePrefixCBRowHL1(oprow)
	case 7: // A
		return c.decodePrefixCBRow1(oprow, A)
	case 8, 9, 10, 11, 12, 13: // B, C, D, E, H, L
		return c.decodePrefixCBRow2(oprow, Register(opcol-8))
	case 14: // (HL)
		return c.decodePrefixCBRowHL2(oprow)
	case 15: // A
		return c.decodePrefixCBRow2(oprow, A)
	default:
		panic("unreachable")
	}
}

// generic decode function for registers B, C, D, E, H, L (columns 0 -> 7)
func (c *CPU) decodePrefixCBRow1(oprow int, r Register) int {
	switch oprow {
	case 0: // RLC
		c.RotateLeftCReg(r)
	case 1: // RL
		c.RotateLeftReg(r)
	case 2: // SLA
		c.ShiftLeftArithmeticReg(r)
	case 3: // SWAP
		c.SwapReg(r)
	case 4, 5, 6, 7: // BIT 0, 2, 4, 6
		bitnum := byte((oprow - 4) * 2)
		c.Bit(r, bitnum)
	case 8, 9, 10, 11: // RES 0, 2, 4, 6
		bitnum := byte((oprow - 8) * 2)
		c.Res(r, bitnum)
	case 12, 13, 14, 15: // SET 0, 2, 4, 6
		bitnum := byte((oprow - 12) * 2)
		c.Set(r, bitnum)
	}

	// all these instructions run in 8 cycles
	return 8
}

// generic decode function for registers B, C, D, E, H, L (columns 8 -> 15)
func (c *CPU) decodePrefixCBRow2(oprow int, r Register) int {
	switch oprow {
	case 0: // RRC
		c.RotateRightCReg(r)
	case 1: // RR
		c.RotateRightReg(r)
	case 2: // SRA
		c.ShiftRightArithmeticReg(r)
	case 3: // SRL
		c.ShiftRightLogicalReg(r)
	case 4, 5, 6, 7: // BIT 1, 3, 5, 7
		bitnum := byte((oprow-4)*2 + 1)
		c.Bit(r, bitnum)
	case 8, 9, 10, 11: // RES 1, 3, 5, 7
		bitnum := byte((oprow-8)*2 + 1)
		c.Res(r, bitnum)
	case 12, 13, 14, 15: // SET 1, 3, 5, 7
		bitnum := byte((oprow-12)*2 + 1)
		c.Set(r, bitnum)
	}

	// all these instructions run in 8 cycles
	return 8
}

// decode function specific to (HL) (columns 0 -> 7)
func (c *CPU) decodePrefixCBRowHL1(oprow int) int {
	switch oprow {
	case 0: // RLC
		c.RotateLeftCHL()
		return 16
	case 1: // RL
		c.RotateLeftHL()
		return 16
	case 2: // SLA
		c.ShiftLeftArithmeticHL()
		return 16
	case 3: // SWAP
		c.SwapHL()
		return 16
	case 4, 5, 6, 7: // BIT 0, 2, 4, 6
		bitnum := byte((oprow - 4) * 2)
		c.BitHL(bitnum)
		return 12
	case 8, 9, 10, 11: // RES 0, 2, 4, 6
		bitnum := byte((oprow - 8) * 2)
		c.ResHL(bitnum)
		return 16
	case 12, 13, 14, 15: // SET 0, 2, 4, 6
		bitnum := byte((oprow - 12) * 2)
		c.SetHL(bitnum)
		return 16
	default:
		panic("unreachable")
	}
}

// decode function specific to (HL) (columns 8 -> 15)
func (c *CPU) decodePrefixCBRowHL2(oprow int) int {
	switch oprow {
	case 0: // RRC
		c.RotateRightCHL()
		return 16
	case 1: // RR
		c.RotateRightHL()
		return 16
	case 2: // SRA
		c.ShiftRightArithmeticHL()
		return 16
	case 3: // SRL
		c.ShiftRightLogicalHL()
		return 16
	case 4, 5, 6, 7: // BIT 1, 3, 5, 7
		bitnum := byte((oprow-4)*2 + 1)
		c.BitHL(bitnum)
		return 12
	case 8, 9, 10, 11: // RES 1, 3, 5, 7
		bitnum := byte((oprow-8)*2 + 1)
		c.ResHL(bitnum)
		return 16
	case 12, 13, 14, 15: // SET 1, 3, 5, 7
		bitnum := byte((oprow-12)*2 + 1)
		c.SetHL(bitnum)
		return 16
	default:
		panic("unreachable")
	}
}
