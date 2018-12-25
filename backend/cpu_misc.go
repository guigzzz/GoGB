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
	} else {
		c.PC += 3
	}
}

// JumpZ jump if zero
func (c *CPU) JumpZ(target uint16) {
	if c.IsFlagSet(ZFlag) {
		c.Jump(target)
	} else {
		c.PC += 3
	}
}

// JumpNC jump if not carry
func (c *CPU) JumpNC(target uint16) {
	if !c.IsFlagSet(CFlag) {
		c.Jump(target)
	} else {
		c.PC += 3
	}
}

// JumpC jump if carry
func (c *CPU) JumpC(target uint16) {
	if c.IsFlagSet(CFlag) {
		c.Jump(target)
	} else {
		c.PC += 3
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

// Ret pop two bytes from stack, build address and jump to that address
func (c *CPU) Ret() {
	v := c.pop()
	addr := PackBytes(c.pop(), v)
	c.Jump(addr)
}

// RetNZ return if not zero
func (c *CPU) RetNZ() {
	if !c.IsFlagSet(ZFlag) {
		c.Ret()
	} else {
		c.PC++
	}
}

// RetZ return if zero
func (c *CPU) RetZ() {
	if c.IsFlagSet(ZFlag) {
		c.Ret()
	} else {
		c.PC++
	}
}

// RetNC return if not carry
func (c *CPU) RetNC() {
	if !c.IsFlagSet(CFlag) {
		c.Ret()
	} else {
		c.PC++
	}
}

// RetC return if carry
func (c *CPU) RetC() {
	if c.IsFlagSet(CFlag) {
		c.Ret()
	} else {
		c.PC++
	}
}

///// CALL /////

// Call push PC to stack and jump to specified address
func (c *CPU) Call(v uint16) {
	// pre-emptively increment PC such that on return,
	// PC points to the next instruction (in memory)
	c.PC += 3
	c.pushPC()
	c.Jump(v)
}

// CallNZ call if not zero
func (c *CPU) CallNZ(v uint16) {
	if !c.IsFlagSet(ZFlag) {
		c.Call(v)
	} else {
		c.PC += 3
	}
}

// CallZ call if zero
func (c *CPU) CallZ(v uint16) {
	if c.IsFlagSet(ZFlag) {
		c.Call(v)
	} else {
		c.PC += 3
	}
}

// CallNC call if not carry
func (c *CPU) CallNC(v uint16) {
	if !c.IsFlagSet(CFlag) {
		c.Call(v)
	} else {
		c.PC += 3
	}
}

// CallC call if carry
func (c *CPU) CallC(v uint16) {
	if c.IsFlagSet(CFlag) {
		c.Call(v)
	} else {
		c.PC += 3
	}
}

///// RST /////

// Rst restart cpu, v can only be:
// 0x00, 0x08, 0x10, 0x18, 0x20, 0x28, 0x30, 0x38
func (c *CPU) Rst(v byte) {
	c.pushPC()
	c.PC = uint16(v)
}
