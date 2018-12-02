package backend

// Load load n into register dest
func (c *CPU) Load(dest Register, n byte) {
	c.reg[dest] = n
}

// LoadReg load value from register src into register dest
func (c *CPU) LoadReg(dest, src Register) {
	c.reg[dest] = c.reg[src]
}

// LoadHL load value from (HL)
func (c *CPU) LoadHL(dest Register) {
	HL := c.ReadHL()
	c.reg[dest] = c.ram[HL]
}

// StoreN store value N to (HL)
func (c *CPU) StoreN(n byte) {
	HL := c.ReadHL()
	c.ram[HL] = n
}

// StoreReg store value at R to (HL)
func (c *CPU) StoreReg(src Register) {
	c.StoreN(c.reg[src])
}

// LoadHigh load with 0xFF00 offset
func (c *CPU) LoadHigh(n byte) {
	c.reg[A] = c.ram[0xFF00+uint16(n)]
}

// StoreHigh store with 0xFF00 offset
func (c *CPU) StoreHigh(n byte) {
	c.ram[0xFF00+uint16(n)] = c.reg[A]
}

// LoadHLSPN implements LD HL, SP+n
func (c *CPU) LoadHLSPN(n byte) {
	v := c.SP + uint16(n)

	c.ResetFlag(ZFlag)
	c.ResetFlag(NFlag)
	c.MaybeFlagSetter(v&0xFFF < c.SP&0xFFF, HFlag)
	c.MaybeFlagSetter(v < c.SP, CFlag)

	c.Writedouble(H, L, v)
}
