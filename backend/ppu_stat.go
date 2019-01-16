package backend

import "fmt"

func (p *PPU) interruptHBlankEnabled() bool {
	return p.ram[0xFF41]&(1<<3) > 0
}

func (p *PPU) interruptVBlankEnabled() bool {
	return p.ram[0xFF41]&(1<<4) > 0
}

func (p *PPU) interruptOAMEnabled() bool {
	return p.ram[0xFF41]&(1<<5) > 0
}

func (p *PPU) interruptLYCEnabled() bool {
	return p.ram[0xFF41]&(1<<6) > 0
}

func (p *PPU) writeLY(lineNumber byte) {
	p.ram[0xFF44] = lineNumber

	if p.coincidence() {
		p.ram[0xFF41] |= 1 << 2
		if p.interruptLYCEnabled() {
			p.dispatchLCDStatInterrupt()
		}
	} else {
		p.ram[0xFF41] &^= 1 << 2
	}
}

func (p *PPU) dispatchVBlankInterrupt() {
	p.ram[0xFF0F] |= 1
}

func (p *PPU) dispatchLCDStatInterrupt() {
	p.ram[0xFF0F] |= 2
}

func (p *PPU) coincidence() bool {
	return p.ram[0xFF45] == p.ram[0xFF44]
}

func (p *PPU) shouldRaiseSTATInterrupt(mode byte) bool {
	return p.interruptHBlankEnabled() && mode == 0 ||
		p.interruptVBlankEnabled() && mode == 1 ||
		p.interruptOAMEnabled() && mode == 2 ||
		p.interruptLYCEnabled() && p.coincidence()
}

var modeStringToNumber = map[string]byte{
	"HBlank": 0,
	"VBlank": 1,
	"OAM":    2,
}

func (p *PPU) setControllerMode(modeString string) {

	if mode, ok := modeStringToNumber[modeString]; ok {

		irq := p.irq
		p.irq = p.shouldRaiseSTATInterrupt(mode)

		if !irq && p.irq {
			p.dispatchLCDStatInterrupt()
		}

		p.ram[0xFF41] = 0x80 | p.ram[0xFF41]&0x7C | mode

	} else {

		if modeString != "PixelTransfer" {
			panic(fmt.Sprintln("Got unexpected LCD controller mode: ", modeString))
		} else {
			p.ram[0xFF41] = 0x80 | p.ram[0xFF41]&0x7C | 3
		}
	}
}
