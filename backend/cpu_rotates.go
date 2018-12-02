package backend

/////////////
// Rotates //
/////////////

///// RLX /////

// RotateLeftn helper function for RL
func (c *CPU) rotateLeftn(n byte) byte {
	carryset := c.ReadFlag(CFlag)
	res := (n << 1) | carryset
	// TODO: Set Flags

	if res == 0 {
		c.SetFlag(ZFlag)
	}
	c.ResetFlag(NFlag)
	c.ResetFlag(HFlag)
	if n&0x80 > 0 {
		c.SetFlag(CFlag)
	}

	return res
}

// RotateLeftReg implements RL for register args
func (c *CPU) RotateLeftReg(r Register) {
	c.reg[r] = c.rotateLeftn(c.reg[r])
}

// RotateLeftHL implements RL for (HL) arg
func (c *CPU) RotateLeftHL() {
	HL := c.ReadHL()
	c.ram[HL] = c.rotateLeftn(c.ram[HL])
}

// RotateLeftCn helper function for RLC
func (c *CPU) rotateLeftCn(n byte) byte {
	carry := byte(0)
	if n&0x80 > 0 {
		carry = 1
	}
	res := (n << 1) | carry

	// TODO: Set Flags
	if res == 0 {
		c.SetFlag(ZFlag)
	}
	c.ResetFlag(NFlag)
	c.ResetFlag(HFlag)
	if n&0x80 > 0 {
		c.SetFlag(CFlag)
	}

	return res
}

// RotateLeftCReg implements RLC for register args
func (c *CPU) RotateLeftCReg(r Register) {
	c.reg[r] = c.rotateLeftCn(c.reg[r])
}

// RotateLeftCHL implements RLC for (HL) arg
func (c *CPU) RotateLeftCHL() {
	HL := c.ReadHL()
	c.ram[HL] = c.rotateLeftCn(c.ram[HL])
}

///// RRX /////

// rotateRightn helper function for RR
func (c *CPU) rotateRightn(n byte) byte {
	carryset := c.ReadFlag(CFlag)
	res := (n >> 1) | carryset

	if res == 0 {
		c.SetFlag(ZFlag)
	}
	c.ResetFlag(NFlag)
	c.ResetFlag(HFlag)
	if n&0x01 > 0 {
		c.SetFlag(CFlag)
	}

	return res
}

// RotateRightReg implements RR for register args
func (c *CPU) RotateRightReg(r Register) {
	c.reg[r] = c.rotateRightn(c.reg[r])
}

// RotateRightHL implements RR for (HL) arg
func (c *CPU) RotateRightHL() {
	HL := c.ReadHL()
	c.ram[HL] = c.rotateRightn(c.ram[HL])
}

// RotateRightCn helper function for RRC
func (c *CPU) rotateRightCn(n byte) byte {
	carry := byte(0)
	if n&0x01 > 0 {
		carry = 1
	}
	res := (n >> 1) | carry

	// TODO: Set Flags
	if res == 0 {
		c.SetFlag(ZFlag)
	}
	c.ResetFlag(NFlag)
	c.ResetFlag(HFlag)
	if n&0x01 > 0 {
		c.SetFlag(CFlag)
	}

	return res
}

// RotateRightCReg implements RRC for register args
func (c *CPU) RotateRightCReg(r Register) {
	c.reg[r] = c.rotateRightCn(c.reg[r])
}

// RotateRightCHL implements RRC for (HL) arg
func (c *CPU) RotateRightCHL() {
	HL := c.ReadHL()
	c.ram[HL] = c.rotateRightCn(c.ram[HL])
}
