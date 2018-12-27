package backend

///////////////////
// ADD + SUB OPS //
///////////////////

// Addn performs A += n where n is an 8 bit number
func (c *CPU) Addn(n byte, carry bool) {

	carry = carry && c.IsFlagSet(CFlag)
	res := c.reg[A] + n
	c.MaybeFlagSetter(c.reg[A]&0xF+n&0xF > 0xF, HFlag)
	c.MaybeFlagSetter(n > 0 && c.reg[A] >= res, CFlag)

	c.reg[A] = res
	if carry {
		if c.reg[A]&0xF == 0xF {
			c.SetFlag(HFlag)
		}
		if c.reg[A] == 0xFF {
			c.SetFlag(CFlag)
		}
		c.reg[A]++
	}

	c.MaybeFlagSetter(c.reg[A] == 0, ZFlag)
	c.ResetFlag(NFlag)
}

// AddHL performs A += (HL) where (HL) is the 8 bit number stored @ (HL)
func (c *CPU) AddHL(carry bool) {
	HL := c.ReadHL()
	c.Addn(c.readMemory(HL), carry)
}

// AddReg performs A += R where R is the 8 bit number stored in register R
func (c *CPU) AddReg(src Register, carry bool) {
	c.Addn(c.reg[src], carry)
}

// Subn performs A -= n where n is an 8 bit number
func (c *CPU) Subn(n byte, carry bool) {

	carry = carry && c.IsFlagSet(CFlag)
	c.MaybeFlagSetter(c.reg[A]&0xF < n&0xF, HFlag)
	c.MaybeFlagSetter(c.reg[A] < n, CFlag)

	c.reg[A] -= n

	if carry {
		if c.reg[A]&0xF == 0 {
			c.SetFlag(HFlag)
		}
		if c.reg[A] == 0 {
			c.SetFlag(CFlag)
		}
		c.reg[A]--
	}

	c.MaybeFlagSetter(c.reg[A] == 0, ZFlag)
	c.SetFlag(NFlag)
}

// SubHL performs A -= (HL) where (HL) is the 8 bit number pointed to by HL
func (c *CPU) SubHL(carry bool) {
	HL := c.ReadHL()
	c.Subn(c.readMemory(HL), carry)
}

// SubReg performs A -= R where R is the 8 bit number stored in register R
func (c *CPU) SubReg(src Register, carry bool) {
	c.Subn(c.reg[src], carry)
}

// Cpn performs A - n but without writing to A (same as TST on ARM?)
func (c *CPU) Cpn(n byte) {
	res := c.reg[A] - n

	c.MaybeFlagSetter(res == 0, ZFlag)
	c.SetFlag(NFlag)
	c.MaybeFlagSetter(c.reg[A]&0xF < n&0xF, HFlag)
	c.MaybeFlagSetter(c.reg[A] < n, CFlag)
}

// CpHL performs A - (HL)
func (c *CPU) CpHL() {
	HL := c.ReadHL()
	c.Cpn(c.readMemory(HL))
}

// CpReg performs A - R
func (c *CPU) CpReg(src Register) {
	c.Cpn(c.reg[src])
}

/////////////////
// BITWISE OPS //
/////////////////

// Andn performs A and n
func (c *CPU) Andn(n byte) {
	c.reg[A] &= n

	c.MaybeFlagSetter(c.reg[A] == 0, ZFlag)
	c.ResetFlag(NFlag)
	c.SetFlag(HFlag)
	c.ResetFlag(CFlag)
}

// AndHL performs A and (HL)
func (c *CPU) AndHL() {
	HL := c.ReadHL()
	c.Andn(c.readMemory(HL))
}

// AndReg performs A and R
func (c *CPU) AndReg(src Register) {
	c.Andn(c.reg[src])
}

// Orn performs A or n
func (c *CPU) Orn(n byte) {
	c.reg[A] |= n

	c.MaybeFlagSetter(c.reg[A] == 0, ZFlag)
	c.ResetFlag(NFlag)
	c.ResetFlag(HFlag)
	c.ResetFlag(CFlag)
}

// OrHL performs A or (HL)
func (c *CPU) OrHL() {
	HL := c.ReadHL()
	c.Orn(c.readMemory(HL))
}

// OrReg performs A or R
func (c *CPU) OrReg(src Register) {
	c.Orn(c.reg[src])
}

// Xorn performs A xor n
func (c *CPU) Xorn(n byte) {
	c.reg[A] ^= n

	c.MaybeFlagSetter(c.reg[A] == 0, ZFlag)
	c.ResetFlag(NFlag)
	c.ResetFlag(HFlag)
	c.ResetFlag(CFlag)
}

// XorHL performs A xor (HL)
func (c *CPU) XorHL() {
	HL := c.ReadHL()
	c.Xorn(c.readMemory(HL))
}

// XorReg performs A xor R
func (c *CPU) XorReg(src Register) {
	c.Xorn(c.reg[src])
}

// helper function that swaps nibbles of a byte
// also sets CPU flags according to result
func (c *CPU) swp(n byte) byte {

	res := (n >> 4) | (n << 4)

	c.MaybeFlagSetter(res == 0, ZFlag)
	c.ResetFlag(NFlag)
	c.ResetFlag(HFlag)
	c.ResetFlag(CFlag)

	return res
}

// SwapHL swaps nibbles at (HL)
func (c *CPU) SwapHL() {
	HL := c.ReadHL()
	c.writeMemory(HL, c.swp(c.readMemory(HL)))
}

// SwapReg swaps nibbles of register r
func (c *CPU) SwapReg(r Register) {
	c.reg[r] = c.swp(c.reg[r])
}

///////////////////
// INC + DEC OPS //
///////////////////

///// 8-BIT /////

// Inc implements 8-bit increment
func (c *CPU) Inc(r Register) {

	res := c.reg[r] + 1

	c.MaybeFlagSetter(res == 0, ZFlag)
	c.ResetFlag(NFlag)
	c.MaybeFlagSetter(c.reg[r]&0xF > res&0xF, HFlag)
	// C not affected

	c.reg[r] = res

}

// IncHL implements 8-bit increment on (HL)
func (c *CPU) IncHL() {
	HL := c.ReadHL()
	res := c.readMemory(HL) + 1

	c.MaybeFlagSetter(res == 0, ZFlag)
	c.ResetFlag(NFlag)
	c.MaybeFlagSetter(c.readMemory(HL)&0xF > res&0xF, HFlag)
	// C not affected

	c.writeMemory(HL, res)
}

// Dec implements 8-bit decrement
func (c *CPU) Dec(r Register) {

	res := c.reg[r] - 1

	c.MaybeFlagSetter(res == 0, ZFlag)
	c.SetFlag(NFlag)
	c.MaybeFlagSetter(c.reg[r]&0xF < res&0xF, HFlag)
	// C not affected

	c.reg[r] = res
}

// DecHL implements 8-bit decrement on (HL)
func (c *CPU) DecHL() {

	HL := c.ReadHL()
	res := c.readMemory(HL) - 1

	c.MaybeFlagSetter(res == 0, ZFlag)
	c.SetFlag(NFlag)
	c.MaybeFlagSetter(c.readMemory(HL)&0xF < res&0xF, HFlag)
	// C not affected

	c.writeMemory(HL, res)
}

///// 16-BIT /////

func (c *CPU) AddHL16(n uint16) {

	HL := c.ReadHL()
	res := HL + n

	// Z unaffected
	c.ResetFlag(NFlag)
	c.MaybeFlagSetter(HL&0xFFF+n&0xFFF > 0xFFF, HFlag)
	c.MaybeFlagSetter(n > 0 && HL >= res, CFlag)

	c.Writedouble(H, L, res)
}

func (c *CPU) AddSP8(n byte) {

	c.ResetFlag(ZFlag)
	c.ResetFlag(NFlag)
	c.MaybeFlagSetter(byte(c.SP&0xF)+n&0xF > 0xF, HFlag)
	c.MaybeFlagSetter(c.SP&0xFF+uint16(n) > 0xFF, CFlag)

	sv := int8(n)
	if sv < 0 {
		c.SP -= uint16(-sv)
	} else {
		c.SP += uint16(sv)
	}
}

func (c *CPU) IncRegs(h, l Register) {
	// check if carry from low reg to high reg
	if c.reg[l] == 0xFF {
		c.reg[h]++
	}
	c.reg[l]++
}

func (c *CPU) IncSP() {
	c.SP++
}

func (c *CPU) DecRegs(h, l Register) {
	if c.reg[l] == 0 {
		c.reg[h]--
	}
	c.reg[l]--
}

func (c *CPU) DecSP() {
	c.SP--
}

///////////////////
// Bit, Set, Res //
///////////////////

// Bit tests bit bitnum of value contained in register r
func (c *CPU) Bit(r Register, bitnum byte) {
	c.MaybeFlagSetter(c.reg[r]&(1<<bitnum) == 0, ZFlag)
	c.ResetFlag(NFlag)
	c.SetFlag(HFlag)

}

// BitHL tests bit bitnum of value contained in (HL)
func (c *CPU) BitHL(bitnum byte) {
	v := c.readMemory(c.ReadHL())
	c.MaybeFlagSetter(v&(1<<bitnum) == 0, ZFlag)
	c.ResetFlag(NFlag)
	c.SetFlag(HFlag)
}

// Res resets bit bitnum of value in register r
func (c *CPU) Res(r Register, bitnum byte) {
	c.reg[r] &^= 1 << bitnum // &^ == AND NOT ---> c.reg[r] = c.reg[r] & !mask
}

// ResHL resets bit bitnum of value in (HL)
func (c *CPU) ResHL(bitnum byte) {
	HL := c.ReadHL()
	res := c.readMemory(HL) &^ (1 << bitnum)
	c.writeMemory(HL, res)
}

// Set sets bit bitnum of value in register r
func (c *CPU) Set(r Register, bitnum byte) {
	c.reg[r] |= 1 << bitnum
}

// SetHL sets bit bitnum of value in (HL)
func (c *CPU) SetHL(bitnum byte) {
	HL := c.ReadHL()
	res := c.readMemory(HL) | (1 << bitnum)
	c.writeMemory(HL, res)
}

/////////
// DAA //
/////////

func (c *CPU) DAA() {

	v := uint16(c.reg[A])

	if !c.IsFlagSet(NFlag) {
		// previous op was addition
		if c.IsFlagSet(HFlag) || v&0xF > 0x9 {
			v += 0x6
		}
		if c.IsFlagSet(CFlag) || v > 0x9F {
			v += 0x60
			c.SetFlag(CFlag)
		}
	} else {
		// previous op was subtraction
		if c.IsFlagSet(HFlag) {
			v -= 0x6
		}

		if c.IsFlagSet(CFlag) {
			v -= 0x60
		}
	}

	c.MaybeFlagSetter(v&0xFF == 0, ZFlag)
	// N unaffected
	c.ResetFlag(HFlag)

	c.reg[A] = byte(v)
}
