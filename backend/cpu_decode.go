package backend

// opcode grid for reference
// http://www.pastraiser.com/cpu/gameboy/gameboy_opcodes.html

// opcode explanations
// http://www.chrisantonellis.com/files/gameboy/gb-instructions.txt

//////////////////
// VariousUpper //
//////////////////

// DecodeVariousUpper decodes various instructions on first 4 rows
func (c *CPU) DecodeVariousUpper(b []byte) {
	op := b[0]
	oprow := int((op & 0xF0) >> 4)
	opcol := int(op & 0xF)

	switch opcol {
	case 0:
		switch oprow {
		case 0: // NOP
			// panic("No Op - Unimplemented")
		case 1: // STOP
			// fmt.Println("Warning: Got STOP instruction, this is probably a bug")
		case 2: // JR NZ,r8
			c.JumpRelativeNZ(b[1])
		case 3: // JR NC,r8
			c.JumpRelativeNC(b[1])
		}
	case 1:
		v := PackBytes(b[2], b[1])
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
	case 4:
		switch oprow {
		case 0: // INC B
			c.Inc(B)
		case 1: // INC D
			c.Inc(D)
		case 2: // INC H
			c.Inc(H)
		case 3: // INC (HL)
			c.IncHL()
		}
	case 5:
		switch oprow {
		case 0: // DEC B
			c.Dec(B)
		case 1: // DEC D
			c.Dec(D)
		case 2: // DEC H
			c.Dec(H)
		case 3: // DEC (HL)
			c.DecHL()
		}
	case 6:
		switch oprow {
		case 0: // LD B, d8
			c.Load(B, b[1])
		case 1: // LD D, d8
			c.Load(D, b[1])
		case 2: // LD H, d8
			c.Load(H, b[1])
		case 3: // LD (HL), d8
			c.StoreN(b[1])
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
	case 8:
		switch oprow {
		case 0: // LD (a16),SP
			c.StoreSPNN(PackBytes(b[2], b[1]))
		case 1: // JR r8
			c.JumpRelative(b[1])
		case 2: // JR Z,r8
			c.JumpRelativeZ(b[1])
		case 3: // JR C,r8
			c.JumpRelativeC(b[1])
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
	case 14:
		switch oprow {
		case 0: // LD C, d8
			c.Load(C, b[1])
		case 1: // LD E, d8
			c.Load(E, b[1])
		case 2: // LD L, d8
			c.Load(L, b[1])
		case 3: // LD A, d8
			c.Load(A, b[1])
		}
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

	}
}

////////////
// Memory //
////////////

// DecodeMem decodes the various LD instructioons
func (c *CPU) DecodeMem(b []byte) {
	op := b[0]
	oprow := int((op&0xF0)>>4) - 4
	opcol := int(op & 0xF)

	switch opcol {

	case 0, 1, 2, 3, 4, 5:
		c.decodeLDArg1(oprow, Register(opcol))
	case 6:
		switch oprow {
		case 0:
			c.LoadHL(B)
		case 1:
			c.LoadHL(D)
		case 2:
			c.LoadHL(H)
		case 3:
			panic("HALT - Unimplemented")
		}
	case 7:
		c.decodeLDArg1(oprow, A)

	case 8, 9, 10, 11, 12, 13:
		c.decodeLDArg2(oprow, Register(opcol-8))
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
	case 15:
		c.decodeLDArg2(oprow, A)

	}
}

func (c *CPU) decodeLDArg1(oprow int, src Register) {
	switch oprow {
	case 0:
		c.LoadReg(B, src)
	case 1:
		c.LoadReg(D, src)
	case 2:
		c.LoadReg(H, src)
	case 3:
		c.StoreReg(src)
	}
}

func (c *CPU) decodeLDArg2(oprow int, src Register) {
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
}

////////////////
// Arithmetic //
////////////////

// DecodeArith decodes various arithmetic instructions
func (c *CPU) DecodeArith(b []byte) {
	op := b[0]
	oprow := int((op&0xF0)>>4) - 8
	opcol := int(op & 0xF)

	switch opcol {

	// first half
	case 0, 1, 2, 3, 4, 5:
		c.decodeArithArg1(oprow, Register(opcol))
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
	case 7:
		c.decodeArithArg1(oprow, A)

	// second half
	case 8, 9, 10, 11, 12, 13:
		c.decodeArithArg2(oprow, Register(opcol-8))
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
	case 15:
		c.decodeArithArg2(oprow, A)
	}
}

func (c *CPU) decodeArithArg1(oprow int, src Register) {
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
}

func (c *CPU) decodeArithArg2(oprow int, src Register) {
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
}

//////////////////
// VariousLower //
//////////////////

// DecodeVariousLower decodes various instructions on last 4 rows
func (c *CPU) DecodeVariousLower(b []byte) {
	op := b[0]
	oprow := int((op&0xF0)>>4) - 12
	opcol := int(op & 0xF)

	switch opcol {
	case 0:
		switch oprow {
		case 0: // RET NZ
			c.RetNZ()
		case 1: // RET NC
			c.RetNC()
		case 2: // LDH (a8),A
			c.StoreHigh(b[1])
		case 3: // LDH A,(a8)
			c.LoadHigh(b[1])
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
	case 2:
		switch oprow {
		case 0: // JP NZ,a16
			v := PackBytes(b[2], b[1])
			c.JumpNZ(v)
		case 1: // JP NC,a16
			v := PackBytes(b[2], b[1])
			c.JumpNC(v)
		case 2: // LD (C),A
			c.StoreHigh(c.reg[C])
		case 3: // LD A,(C)
			c.LoadHigh(c.reg[C])
		}
	case 3:
		switch oprow {
		case 0: // JP a16
			v := PackBytes(b[2], b[1])
			c.Jump(v)
		case 1, 2: // NONE
			panic("ERROR - byte decoded to unused instruction -> there is a bug somewhere")
		case 3: // DI
			c.IME = false
		}
	case 4:
		v := PackBytes(b[2], b[1])
		switch oprow {
		case 0: // CALL NZ,a16
			c.CallNZ(v)
		case 1: // CALL NC,a16
			c.CallNC(v)
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
	case 6:
		switch oprow {
		case 0: // ADD A,d8
			c.Addn(b[1], false)
		case 1: // SUB d8
			c.Subn(b[1], false)
		case 2: // AND d8
			c.Andn(b[1])
		case 3: // OR d8
			c.Orn(b[1])
		}
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
	case 8:
		switch oprow {
		case 0: // RET Z
			c.RetZ()
		case 1: // RET C
			c.RetC()
		case 2: // ADD SP,r8
			c.AddSP8(b[1])
		case 3: // LD HL,SP+r8
			c.LoadHLSPN(b[1])
		}
	case 9:
		switch oprow {
		case 0: // RET
			c.Ret()
		case 1: // RETI
			c.IME = true
			c.Ret()
		case 2: // JP (HL)
			c.Jump(c.ReadHL())
		case 3: // LD SP,HL
			c.SP = c.ReadHL()
		}
	case 10:
		v := PackBytes(b[2], b[1])
		switch oprow {
		case 0: // JP Z,a16
			c.JumpZ(v)
		case 1: // JP C,a16
			c.JumpC(v)
		case 2: // LD (a16),A
			c.writeMemory(v, c.reg[A])
		case 3: // LD A,(a16)
			c.reg[A] = c.readMemory(v)
		}
	case 11:
		switch oprow {
		case 0: // PREFIX CB
			c.DecodePrefixCB(b[1])
		case 1, 2: // NONE
			panic("ERROR - byte decoded to unused instruction -> there is a bug somewhere")
		case 3: // EI
			c.IME = true
		}
	case 12:
		v := PackBytes(b[2], b[1])
		switch oprow {
		case 0: // CALL Z,a16
			c.CallZ(v)
		case 1: // CALL C,a16
			c.CallC(v)
		case 2, 3: // NONE
			panic("ERROR - byte decoded to unused instruction -> there is a bug somewhere")
		}
	case 13:
		switch oprow {
		case 0: // CALL a16
			v := PackBytes(b[2], b[1])
			c.Call(v)
		case 1, 2, 3: // NONE
			panic("ERROR - byte decoded to unused instruction -> there is a bug somewhere")
		}
	case 14:
		switch oprow {
		case 0: // ADC A,d8
			c.Addn(b[1], true)
		case 1: // SBC A,d8
			c.Subn(b[1], true)
		case 2: // XOR d8
			c.Xorn(b[1])
		case 3: // CP d8
			c.Cpn(b[1])
		}
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
	}
}

// DecodePrefixCB decodes instructions prefixed with CB such as bit, res, set, etc.
func (c *CPU) DecodePrefixCB(op byte) {
	oprow := int((op & 0xF0) >> 4)
	opcol := int(op & 0xF)

	switch opcol {
	case 0, 1, 2, 3, 4, 5: // B, C, D, E, H, L
		c.decodePrefixCBRow1(oprow, Register(opcol))
	case 6: // (HL)
		c.decodePrefixCBRowHL1(oprow)
	case 7: // A
		c.decodePrefixCBRow1(oprow, A)
	case 8, 9, 10, 11, 12, 13: // B, C, D, E, H, L
		c.decodePrefixCBRow2(oprow, Register(opcol-8))
	case 14: // (HL)
		c.decodePrefixCBRowHL2(oprow)
	case 15: // A
		c.decodePrefixCBRow2(oprow, A)
	}
}

// generic decode function for registers B, C, D, E, H, L (columns 0 -> 7)
func (c *CPU) decodePrefixCBRow1(oprow int, r Register) {
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
}

// generic decode function for registers B, C, D, E, H, L (columns 8 -> 15)
func (c *CPU) decodePrefixCBRow2(oprow int, r Register) {
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
}

// decode function specific to (HL) (columns 0 -> 7)
func (c *CPU) decodePrefixCBRowHL1(oprow int) {
	switch oprow {
	case 0: // RLC
		c.RotateLeftCHL()
	case 1: // RL
		c.RotateLeftHL()
	case 2: // SLA
		c.ShiftLeftArithmeticHL()
	case 3: // SWAP
		c.SwapHL()
	case 4, 5, 6, 7: // BIT 0, 2, 4, 6
		bitnum := byte((oprow - 4) * 2)
		c.BitHL(bitnum)
	case 8, 9, 10, 11: // RES 0, 2, 4, 6
		bitnum := byte((oprow - 8) * 2)
		c.ResHL(bitnum)
	case 12, 13, 14, 15: // SET 0, 2, 4, 6
		bitnum := byte((oprow - 12) * 2)
		c.SetHL(bitnum)
	}
}

// decode function specific to (HL) (columns 8 -> 15)
func (c *CPU) decodePrefixCBRowHL2(oprow int) {
	switch oprow {
	case 0: // RRC
		c.RotateRightCHL()
	case 1: // RR
		c.RotateRightHL()
	case 2: // SRA
		c.ShiftRightArithmeticHL()
	case 3: // SRL
		c.ShiftRightLogicalHL()
	case 4, 5, 6, 7: // BIT 1, 3, 5, 7
		bitnum := byte((oprow-4)*2 + 1)
		c.BitHL(bitnum)
	case 8, 9, 10, 11: // RES 1, 3, 5, 7
		bitnum := byte((oprow-8)*2 + 1)
		c.ResHL(bitnum)
	case 12, 13, 14, 15: // SET 1, 3, 5, 7
		bitnum := byte((oprow-12)*2 + 1)
		c.SetHL(bitnum)
	}
}
