package backend

///// PUSH & POP /////

func (c *CPU) push2(h, l byte) {
	c.SP--
	c.writeMemory(c.SP, h)
	c.SP--
	c.writeMemory(c.SP, l)
}

func (c *CPU) pop2() (byte, byte) {
	l := c.readMemory(c.SP)
	c.SP++
	h := c.readMemory(c.SP)
	c.SP++
	return h, l
}

func (c *CPU) pushPC() {
	c.push2(byte(c.PC>>8), byte(c.PC&0xFF))
}

func (c *CPU) popPC() uint16 {
	top, bottom := c.pop2()
	return PackBytes(top, bottom)
}

func (c *CPU) pushDouble(h, l Register) {
	c.push2(c.reg[h], c.reg[l])
}

func (c *CPU) popDouble(h, l Register) {
	c.reg[h], c.reg[l] = c.pop2()
}

///// JUMP /////

// Jump implements basic target jumps
func (c *CPU) Jump(v uint16) {
	c.PC = v
}

// JumpNZ jump if not zero
func (c *CPU) JumpNZ(target uint16) (pcInc, cycleInc int) {
	if !c.IsFlagSet(ZFlag) {
		c.Jump(target)
		return 0, 16
	} else {
		return 3, 12
	}
}

// JumpZ jump if zero
func (c *CPU) JumpZ(target uint16) (pcInc, cycleInc int) {
	if c.IsFlagSet(ZFlag) {
		c.Jump(target)
		return 0, 16
	} else {
		return 3, 12
	}
}

// JumpNC jump if not carry
func (c *CPU) JumpNC(target uint16) (pcInc, cycleInc int) {
	if !c.IsFlagSet(CFlag) {
		c.Jump(target)
		return 0, 16
	} else {
		return 3, 12
	}
}

// JumpC jump if carry
func (c *CPU) JumpC(target uint16) (pcInc, cycleInc int) {
	if c.IsFlagSet(CFlag) {
		c.Jump(target)
		return 0, 16
	} else {
		return 3, 12
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
func (c *CPU) JumpRelativeNZ(v byte) (pcInc, cycleInc int) {
	if !c.IsFlagSet(ZFlag) {
		c.JumpRelative(v)
		return 2, 12
	} else {
		return 2, 8
	}
}

// JumpRelativeZ jump if zero
func (c *CPU) JumpRelativeZ(v byte) (pcInc, cycleInc int) {
	if c.IsFlagSet(ZFlag) {
		c.JumpRelative(v)
		return 2, 12
	} else {
		return 2, 8
	}
}

// JumpRelativeNC jump if not carry
func (c *CPU) JumpRelativeNC(v byte) (pcInc, cycleInc int) {
	if !c.IsFlagSet(CFlag) {
		c.JumpRelative(v)
		return 2, 12
	} else {
		return 2, 8
	}
}

// JumpRelativeC jump if carry
func (c *CPU) JumpRelativeC(v byte) (pcInc, cycleInc int) {
	if c.IsFlagSet(CFlag) {
		c.JumpRelative(v)
		return 2, 12
	} else {
		return 2, 8
	}
}

///// RET /////

// Ret pop two bytes from stack, build address and jump to that address
func (c *CPU) Ret() {
	c.PC = c.popPC()
}

// RetNZ return if not zero
func (c *CPU) RetNZ() (pcInc, cycleInc int) {
	if !c.IsFlagSet(ZFlag) {
		c.Ret()
		return 0, 20
	} else {
		// branch not taken ?
		return 1, 8
	}
}

// RetZ return if zero
func (c *CPU) RetZ() (pcInc, cycleInc int) {
	if c.IsFlagSet(ZFlag) {
		c.Ret()
		return 0, 20
	} else {
		return 1, 8
	}
}

// RetNC return if not carry
func (c *CPU) RetNC() (pcInc, cycleInc int) {
	if !c.IsFlagSet(CFlag) {
		c.Ret()
		return 0, 20
	} else {
		return 1, 8
	}
}

// RetC return if carry
func (c *CPU) RetC() (pcInc, cycleInc int) {
	if c.IsFlagSet(CFlag) {
		c.Ret()
		return 0, 20
	} else {
		return 1, 8
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
func (c *CPU) CallNZ(v uint16) (pcInc, cycleInc int) {
	if !c.IsFlagSet(ZFlag) {
		c.Call(v)
		return 0, 24
	} else {
		return 3, 12
	}
}

// CallZ call if zero
func (c *CPU) CallZ(v uint16) (pcInc, cycleInc int) {
	if c.IsFlagSet(ZFlag) {
		c.Call(v)
		return 0, 24
	} else {
		return 3, 12
	}
}

// CallNC call if not carry
func (c *CPU) CallNC(v uint16) (pcInc, cycleInc int) {
	if !c.IsFlagSet(CFlag) {
		c.Call(v)
		return 0, 24
	} else {
		return 3, 12
	}
}

// CallC call if carry
func (c *CPU) CallC(v uint16) (pcInc, cycleInc int) {
	if c.IsFlagSet(CFlag) {
		c.Call(v)
		return 0, 24
	} else {
		return 3, 12
	}
}

///// RST /////

// Rst restart cpu, v can only be:
// 0x00, 0x08, 0x10, 0x18, 0x20, 0x28, 0x30, 0x38
func (c *CPU) Rst(v byte) {
	c.PC++
	c.pushPC()
	c.PC = uint16(v)
}
