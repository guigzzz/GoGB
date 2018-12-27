package backend

// Register maps register "name" to index in CPU.reg
type Register int

// Register enums
const (
	B = iota
	C
	D
	E
	H
	L
	A
	F
)

///// REGISTER UTILS /////

// ReadAF helper function to read BC
func (c *CPU) ReadAF() uint16 {
	return c.Readdouble(A, F)
}

// ReadBC helper function to read BC
func (c *CPU) ReadBC() uint16 {
	return c.Readdouble(B, C)
}

// ReadDE helper function to read DE
func (c *CPU) ReadDE() uint16 {
	return c.Readdouble(D, E)
}

// ReadHL helper function to read HL
func (c *CPU) ReadHL() uint16 {
	return c.Readdouble(H, L)
}

// Readdouble is a helper function used to read a double register
func (c *CPU) Readdouble(h, l Register) uint16 {
	return PackBytes(c.reg[h], c.reg[l])
}

// PackBytes packs two bytes into a single uint16
func PackBytes(h, l byte) uint16 {
	return (uint16(h) << 8) | uint16(l)
}

// Writedouble is a helper function used to split a uint16 into high and low bytes
// then writing those bytes to the relevant registers
func (c *CPU) Writedouble(h, l Register, n uint16) {
	c.reg[h] = byte(n >> 8)
	c.reg[l] = byte(n & 0xFF)
}

// Flag maps flag "name" to bit index in F register
type Flag int

// Flag enums
const (
	CFlag = 1 << (4 + iota)
	HFlag
	NFlag
	ZFlag
)

///// FLAG UTILS /////

// SetFlag sets the bit in F register corresponding to the Flag f
func (c *CPU) SetFlag(f Flag) {
	c.reg[F] |= byte(f)
}

// IsFlagSet checks if given flag f is set
func (c *CPU) IsFlagSet(f Flag) bool {
	return c.reg[F]&byte(f) > 0
}

// ResetFlag resets the bit in F register corresponding to the Flag f
func (c *CPU) ResetFlag(f Flag) {
	c.reg[F] &^= byte(f)
}

// ReadFlag is used for operations that require reading carry bit (i.e. ADC)
// returns value in byte, 1 if bit is set, 0 otherwise
func (c *CPU) ReadFlag(f Flag) byte {
	if c.IsFlagSet(f) {
		return 1
	}
	return 0
}

// MaybeFlagSetter is a helper function used to set a certain flag if that flag's condition is true
// else reset the flag
func (c *CPU) MaybeFlagSetter(condition bool, f Flag) {
	if condition {
		c.SetFlag(f)
	} else {
		c.ResetFlag(f)
	}
}
