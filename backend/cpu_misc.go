package backend

///// PUSH & POP /////

func (c *CPU) push(n byte) {
	c.SP--
	c.ram[c.SP] = n
}

func (c *CPU) pop() byte {
	ret := c.ram[c.SP]
	c.SP++
	return ret
}

// PushReg register value to stack
func (c *CPU) PushReg(r Register) {
	c.push(c.reg[r])
}

// PopReg pop value from stack into register
func (c *CPU) PopReg(r Register) {
	c.reg[r] = c.pop()
}

// pushPC push PCh to (SP-1) and PCl to (SP-2)
func (c *CPU) pushPC() {
	c.push(byte(c.PC & 0xFF00 >> 8))
	c.push(byte(c.PC & 0xFF))
}

///// JUMP /////

// Jump implements basic target jumps
func (c *CPU) Jump(v uint16) {
	c.PC = v
}

// JumpNZ jump if not zero
func (c *CPU) JumpNZ(target uint16) {
	if !c.IsFlagSet(ZFlag) {
		c.Jump(target)
	}
}

// JumpZ jump if zero
func (c *CPU) JumpZ(target uint16) {
	if c.IsFlagSet(ZFlag) {
		c.Jump(target)
	}
}

// JumpNC jump if not carry
func (c *CPU) JumpNC(target uint16) {
	if !c.IsFlagSet(CFlag) {
		c.Jump(target)
	}
}

// JumpC jump if carry
func (c *CPU) JumpC(target uint16) {
	if c.IsFlagSet(CFlag) {
		c.Jump(target)
	}
}

///// JUMP RELATIVE /////

// JumpRelative implements basic relative jumps
func (c *CPU) JumpRelative(v byte) {
	sv := int8(v) // interpret value as signed
	if sv < 0 {
		c.PC -= uint16(-sv)
	} else {
		c.PC += uint16(sv)
	}
}

// JumpRelativeNZ jump if non zero
func (c *CPU) JumpRelativeNZ(v byte) {
	if !c.IsFlagSet(ZFlag) {
		c.JumpRelative(v)
	}
}

// JumpRelativeZ jump if zero
func (c *CPU) JumpRelativeZ(v byte) {
	if c.IsFlagSet(ZFlag) {
		c.JumpRelative(v)
	}
}

// JumpRelativeNC jump if not carry
func (c *CPU) JumpRelativeNC(v byte) {
	if !c.IsFlagSet(CFlag) {
		c.JumpRelative(v)
	}
}

// JumpRelativeC jump if carry
func (c *CPU) JumpRelativeC(v byte) {
	if c.IsFlagSet(CFlag) {
		c.JumpRelative(v)
	}
}

///// RET /////

func (c *CPU) Ret() {
	v := c.pop()
	addr := PackBytes(c.pop(), v)
	c.Jump(addr)
}

func (c *CPU) RetNZ() {
	if !c.IsFlagSet(ZFlag) {
		c.Ret()
	}
}

func (c *CPU) RetZ() {
	if c.IsFlagSet(ZFlag) {
		c.Ret()
	}
}

func (c *CPU) RetNC() {
	if !c.IsFlagSet(CFlag) {
		c.Ret()
	}
}

func (c *CPU) RetC() {
	if c.IsFlagSet(CFlag) {
		c.Ret()
	}
}

///// CALL /////

func (c *CPU) Call(v uint16) {
	// pre-emptively increment PC such that on return,
	// PC points to the next instruction (in memory)
	c.PC += 3
	c.pushPC()
	c.Jump(v)
}

func (c *CPU) CallNZ(v uint16) {
	if !c.IsFlagSet(ZFlag) {
		c.Call(v)
	}
}

func (c *CPU) CallZ(v uint16) {
	if c.IsFlagSet(ZFlag) {
		c.Call(v)
	}
}

func (c *CPU) CallNC(v uint16) {
	if !c.IsFlagSet(CFlag) {
		c.Call(v)
	}
}

func (c *CPU) CallC(v uint16) {
	if c.IsFlagSet(CFlag) {
		c.Call(v)
	}
}

///// RST /////

func (c *CPU) Rst(v byte) {
	c.pushPC()
	c.PC = uint16(v)
}
