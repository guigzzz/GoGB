package backend

////////////
// SHIFTS //
////////////

func (c *CPU) ShiftLeftArithmetic(n byte) byte {
	if n&0x80 > 0 {
		c.SetFlag(CFlag)
	}
	nv := n << 1

	c.ResetFlag(NFlag)
	c.ResetFlag(HFlag)
	if nv == 0 {
		c.SetFlag(ZFlag)
	}
	return nv
}

func (c *CPU) ShiftLeftArithmeticReg(r Register) {
	c.reg[r] = c.ShiftLeftArithmetic(c.reg[r])
}

func (c *CPU) ShiftLeftArithmeticHL() {
	HL := c.ReadHL()
	c.ram[HL] = c.ShiftLeftArithmetic(c.ram[HL])
}

func (c *CPU) ShiftRightLogical(n byte) byte {
	if n&0x01 > 0 {
		c.SetFlag(CFlag)
	}
	nv := n >> 1

	c.ResetFlag(NFlag)
	c.ResetFlag(HFlag)
	if nv == 0 {
		c.SetFlag(ZFlag)
	}
	return nv
}

func (c *CPU) ShiftRightLogicalReg(r Register) {
	c.reg[r] = c.ShiftRightLogical(c.reg[r])
}

func (c *CPU) ShiftRightLogicalHL() {
	HL := c.ReadHL()
	c.ram[HL] = c.ShiftRightLogical(c.ram[HL])
}

func (c *CPU) ShiftRightArithmetic(n byte) byte {
	if n&0x01 > 0 {
		c.SetFlag(CFlag)
	}

	// golang performs arithmetic shift if shift applied to signed value
	// https://medium.com/learning-the-go-programming-language/bit-hacking-with-go-e0acee258827
	// hence cast byte to signed 8-bit int and perform shift
	nv := int8(n) >> 1

	c.ResetFlag(NFlag)
	c.ResetFlag(HFlag)
	if nv == 0 {
		c.SetFlag(ZFlag)
	}

	return byte(nv)
}

func (c *CPU) ShiftRightArithmeticReg(r Register) {
	c.reg[r] = c.ShiftRightArithmetic(c.reg[r])
}

func (c *CPU) ShiftRightArithmeticHL() {
	HL := c.ReadHL()
	c.ram[HL] = c.ShiftRightArithmetic(c.ram[HL])
}
