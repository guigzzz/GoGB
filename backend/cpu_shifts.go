package backend

////////////
// SHIFTS //
////////////

// ShiftLeftArithmetic implements an arithmetic left shift on a scalar n
func (c *CPU) ShiftLeftArithmetic(n byte) byte {
	nv := n << 1

	c.MaybeFlagSetter(n&0x80 > 0, CFlag)
	c.ResetFlag(NFlag)
	c.ResetFlag(HFlag)
	c.MaybeFlagSetter(nv == 0, ZFlag)

	return nv
}

// ShiftLeftArithmeticReg performs a left arithmetic shift on a register value
func (c *CPU) ShiftLeftArithmeticReg(r Register) {
	c.reg[r] = c.ShiftLeftArithmetic(c.reg[r])
}

// ShiftLeftArithmeticHL performs a left arithmetic shift on a memory value
func (c *CPU) ShiftLeftArithmeticHL() {
	HL := c.ReadHL()
	c.ram[HL] = c.ShiftLeftArithmetic(c.ram[HL])
}

// ShiftRightLogical implements a logical right shift on a scalar n
func (c *CPU) ShiftRightLogical(n byte) byte {
	nv := n >> 1

	c.MaybeFlagSetter(n&0x01 > 0, CFlag)
	c.ResetFlag(NFlag)
	c.ResetFlag(HFlag)
	c.MaybeFlagSetter(nv == 0, ZFlag)

	return nv
}

// ShiftRightLogicalReg implements a logical right shift on a register value
func (c *CPU) ShiftRightLogicalReg(r Register) {
	c.reg[r] = c.ShiftRightLogical(c.reg[r])
}

// ShiftRightLogicalHL implements a logical right shift on a memory value
func (c *CPU) ShiftRightLogicalHL() {
	HL := c.ReadHL()
	c.ram[HL] = c.ShiftRightLogical(c.ram[HL])
}

// ShiftRightArithmetic implements an arithmetic right shift on a scalar value
// i.e. the sign bit is extended
func (c *CPU) ShiftRightArithmetic(n byte) byte {
	// golang performs arithmetic shift if shift applied to signed value
	// https://medium.com/learning-the-go-programming-language/bit-hacking-with-go-e0acee258827
	// hence cast byte to signed 8-bit int and perform shift
	nv := int8(n) >> 1

	c.MaybeFlagSetter(n&0x01 > 0, CFlag)
	c.ResetFlag(NFlag)
	c.ResetFlag(HFlag)
	c.MaybeFlagSetter(nv == 0, ZFlag)

	return byte(nv)
}

// ShiftRightArithmeticReg implements a right arithmetic shift on a register value
func (c *CPU) ShiftRightArithmeticReg(r Register) {
	c.reg[r] = c.ShiftRightArithmetic(c.reg[r])
}

// ShiftRightArithmeticHL implements a right arithmetic shift on a memory value
func (c *CPU) ShiftRightArithmeticHL() {
	HL := c.ReadHL()
	c.ram[HL] = c.ShiftRightArithmetic(c.ram[HL])
}
