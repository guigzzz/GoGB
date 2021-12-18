package backend

import (
	"fmt"
	"image"
	"image/color"
	"sort"
	"sync"
	"time"
)

// PPU represents the pixel processing unit
// contains references to ram sections containing video relevant data
type PPU struct {
	ram          []byte          // reference to memory shared with CPU
	Image        *image.RGBA     // represents the current screen
	ImageMutex   *sync.RWMutex   // to ensure safety when writing to screen buffer
	screenBuffer [144 * 160]byte // contains the pixels to draw on next refresh

	cpu *CPU

	irq bool
}

// NewPPU creates a new PPU object
func NewPPU(c *CPU) *PPU {
	p := new(PPU)
	p.ram = c.ram
	p.Image = image.NewRGBA(image.Rectangle{image.Point{0, 0}, image.Point{160, 144}})
	p.ImageMutex = new(sync.RWMutex)
	p.cpu = c
	return p
}

const (
	bgDisplay                  = iota // (0=Off, 1=On)
	objDisplayEnable                  // (0=Off, 1=On)
	objSize                           // (0=8x8, 1=8x16)
	bgTileMapDisplaySelect            // (0=9800-9BFF, 1=9C00-9FFF)
	bgWindowTileDataSelect            // (0=8800-97FF, 1=8000-8FFF)
	windowDisplayEnable               // (0=Off, 1=On)
	windowTileMapDisplaySelect        // (0=9800-9BFF, 1=9C00-9FFF)
	lcdDisplayEnable                  // (0=Off, 1=On)
)

func (p *PPU) LCDCBitSet(bitnum uint) bool {
	if bitnum > 7 {
		panic(fmt.Sprintf("Got unexpected bit number %d higher than 7 (max for byte)", bitnum))
	}
	return p.ram[0xFF40]&(1<<bitnum) > 0
}

func (p *PPU) getBackgroundTileData() ([]byte, bool) {
	if p.LCDCBitSet(bgWindowTileDataSelect) {
		return p.ram[0x8000 : 0x8FFF+1], false
	}
	return p.ram[0x8800 : 0x97FF+1], true
}

func (p *PPU) getWindowTileData() ([]byte, bool) {
	return p.getBackgroundTileData()
}

func (p *PPU) getBackgroundTileMap() []byte {
	if p.LCDCBitSet(bgTileMapDisplaySelect) {
		return p.ram[0x9C00 : 0x9FFF+1]
	}
	return p.ram[0x9800 : 0x9BFF+1]
}

func (p *PPU) getWindowTileMap() []byte {
	if p.LCDCBitSet(windowTileMapDisplaySelect) {
		return p.ram[0x9C00 : 0x9FFF+1]
	}
	return p.ram[0x9800 : 0x9BFF+1]
}

func mapColorToPalette(palette byte, color byte) byte {
	return (palette >> (color * 2)) & 0x3
}

func (p *PPU) getBGPalette() byte {
	return p.ram[0xFF47]
}

func (p *PPU) getScroll() (byte, byte) {
	return p.ram[0xFF42], p.ram[0xFF43]
}

func (p *PPU) getBackgroundPixels(lineNumber byte) [160]byte {

	pixels := [160]byte{}
	if !p.LCDCBitSet(bgDisplay) {
		return pixels
	}

	tileMap := p.getBackgroundTileMap()
	tileData, interpretIndexAsSigned := p.getBackgroundTileData()

	scrollY, scrollX := p.getScroll()
	rowInTile := (scrollY + lineNumber) % 8
	tileRow := (scrollY + lineNumber) / 8

	for i := byte(0); i < 160; i++ {

		// compute in which background tile we fall in (in a 32 x 32 grid)
		tileColumn := (scrollX + i) / 8
		tileIndex := uint(tileRow)*32 + uint(tileColumn)

		// get the tile data index for that tile
		tileMapIndex := tileMap[tileIndex]

		// if we are using 0x8000 to 0x8FFF
		// then 0-127 maps to 8000-87FF and 128-255 maps to 8800-8FFF
		//
		// if we are using the 0x8800 to 0x97FF
		// then 0-127 maps to 9000-97FF whereas 128-255 maps to 8800-8FFF
		//
		// we can just flip the MSB of the data index in the 0x8800 to 0x97FF case
		if interpretIndexAsSigned {
			tileMapIndex ^= 0x80
		}

		// 16 bytes per tile, 8 lines of 8 pixels per tiles
		// meaning 2 bytes per line
		lineDataIndex := uint(tileMapIndex)*16 + 2*uint(rowInTile)
		lineData := tileData[lineDataIndex : lineDataIndex+2]

		pixelInLine := (scrollX + i) % 8
		msb := (lineData[1] >> (7 - pixelInLine)) & 1
		lsb := (lineData[0] >> (7 - pixelInLine)) & 1

		colorCode := (msb << 1) | lsb

		pixels[i] = mapColorToPalette(p.getBGPalette(), colorCode)
	}

	return pixels
}

func (p *PPU) getWindowPosition() (byte, byte) {
	return p.ram[0xFF4A], p.ram[0xFF4B] - 7
}

func (p *PPU) getWindowPixels(lineNumber byte) [160]byte {

	pixels := [160]byte{4}
	for i := 0; i < 160; i++ {
		pixels[i] = 4
	}
	if !p.LCDCBitSet(windowDisplayEnable) {
		return pixels
	}

	yPos, xPos := p.getWindowPosition()
	if yPos > lineNumber || xPos > 159 {
		return pixels
	}

	tileMap := p.getWindowTileMap()
	tileData, interpretIndexAsSigned := p.getWindowTileData()

	rowInTile := (lineNumber - yPos) % 8
	tileRow := (lineNumber - yPos) / 8

	for i := 0; i < 160-int(xPos); i++ {

		// compute in which background tile we fall in (in a 32 x 32 grid)
		tileColumn := i / 8
		tileIndex := uint(tileRow)*32 + uint(tileColumn)

		// get the tile data index for that tile
		tileMapIndex := tileMap[tileIndex]

		// if we are using 0x8000 to 0x8FFF
		// then 0-127 maps to 8000-87FF and 128-255 maps to 8800-8FFF
		//
		// if we are using the 0x8800 to 0x97FF
		// then 0-127 maps to 9000-97FF whereas 128-255 maps to 8800-8FFF
		//
		// we can just flip the MSB of the data index in the 0x8800 to 0x97FF case
		if interpretIndexAsSigned {
			tileMapIndex ^= 0x80
		}

		// 16 bytes per tile, 8 lines of 8 pixels per tiles
		// meaning 2 bytes per line
		lineDataIndex := uint(tileMapIndex)*16 + 2*uint(rowInTile)
		lineData := tileData[lineDataIndex : lineDataIndex+2]

		pixelInLine := uint(i) % 8
		msb := (lineData[1] >> (7 - pixelInLine)) & 1
		lsb := (lineData[0] >> (7 - pixelInLine)) & 1

		colorCode := (msb << 1) | lsb

		pixels[int(xPos)+i] = mapColorToPalette(p.getBGPalette(), colorCode)
	}

	return pixels

}

func (p *PPU) getSpriteData() []byte {
	return p.ram[0x8000 : 0x8FFF+1]
}

func (p *PPU) getSpriteAttributes() []byte {
	return p.ram[0xFE00 : 0xFE9F+1]
}

func (p *PPU) getSpritePalette(paletteNumber byte) byte {
	if paletteNumber&0x10 > 0 {
		return p.ram[0xFF49]
	}
	return p.ram[0xFF48]
}

func (p *PPU) getSpriteHeight() byte {
	if p.LCDCBitSet(objSize) {
		return 16
	}
	return 8
}

type Sprite struct {
	position  int
	xPos      byte
	yPos      byte
	tileIndex byte
	palette   byte
	xFlipped  bool
	yFlipped  bool
	priority  bool
}

func (p *PPU) getSpritePixels(lineNumber byte) ([160]byte, [160]byte, [160]bool) {

	if !p.LCDCBitSet(objDisplayEnable) {
		return [160]byte{}, [160]byte{}, [160]bool{}
	}

	// Implements OAM searching and sprite rendering
	attributes := p.getSpriteAttributes()
	spriteHeight := p.getSpriteHeight()
	sprites := make([]Sprite, 0, 10)

	for i := 0; i < 40 && len(sprites) < 10; i++ {

		yPos := attributes[4*i]
		if yPos > lineNumber+16 || lineNumber+16 >= yPos+spriteHeight {
			continue
		}

		xPos := attributes[4*i+1]
		if xPos == 0 {
			continue
		}

		tileIndex := attributes[4*i+2]
		flags := attributes[4*i+3]
		xFlipped := flags&0x20 > 0
		yFlipped := flags&0x40 > 0
		palette := p.getSpritePalette(flags)
		priority := flags&0x80 == 0

		sprites = append(sprites, Sprite{i, xPos, yPos, tileIndex, palette, xFlipped, yFlipped, priority})
	}

	sort.Slice(sprites, func(i, j int) bool {
		if sprites[i].xPos < sprites[j].xPos {
			return true
		} else if sprites[i].xPos == sprites[j].xPos && sprites[i].position < sprites[j].position {
			return true
		}
		return false
	})

	pixels := [160]byte{}
	palettes := [160]byte{}
	priorities := [160]bool{}
	tileData := p.getSpriteData()
	for _, s := range sprites {
		var rowInTile byte
		if s.yPos < 16 {
			rowInTile = 16 - s.yPos + lineNumber
		} else {
			rowInTile = lineNumber - (s.yPos - 16)
		}

		if spriteHeight == 16 {
			if rowInTile >= 8 {
				s.tileIndex |= 1
				rowInTile -= 8
			} else {
				s.tileIndex &= 0xFE
			}
		}

		lineDataIndex := uint(s.tileIndex)*16 + 2*uint(rowInTile)
		if s.yFlipped {
			lineDataIndex = uint(s.tileIndex)*16 + 2*7 - 2*uint(rowInTile)
		}
		lsbs := tileData[lineDataIndex]
		msbs := tileData[lineDataIndex+1]
		if s.xFlipped {
			lsbs = reverse(lsbs)
			msbs = reverse(msbs)
		}

		for l := 0; l < 8; l++ {
			if s.xPos < 8-byte(l) {
				continue
			}
			msb := (msbs >> (7 - byte(l))) & 1
			lsb := (lsbs >> (7 - byte(l))) & 1

			colorCode := (msb << 1) | lsb

			pos := s.xPos - 8 + byte(l)
			if pos <= 159 && pixels[pos] == 0 && colorCode > 0 {
				pixels[pos] = colorCode
				palettes[pos] = s.palette
				priorities[pos] = s.priority
			}
		}
	}

	return pixels, palettes, priorities
}

func reverse(in byte) byte {
	return (in & 0x1 << 7) | (in & 0x2 << 5) | (in & 0x4 << 3) | (in & 0x8 << 1) |
		(in & 0x10 >> 1) | (in & 0x20 >> 3) | (in & 0x40 >> 5) | (in & 0x80 >> 7)
}

func (p *PPU) RunCPU(cycles int) {
	p.cpu.RunSync(cycles)
}

func (p *PPU) runEmulatorForAFrame(canRenderScreen chan struct{}) {

	if !p.LCDCBitSet(lcdDisplayEnable) {
		p.RunCPU(154 * 114 * 4)
		return
	}

	for lineNumber := byte(0); lineNumber < 144; lineNumber++ {

		p.writeLY(lineNumber)
		p.setControllerMode("OAM")
		p.RunCPU(20 * 4) // OAM search allowance

		// OAM search

		p.setControllerMode("PixelTransfer")

		p.RunCPU(43 * 4) // pixel transfer allowance

		// pixel transfer
		background := p.getBackgroundPixels(lineNumber)
		window := p.getWindowPixels(lineNumber)
		sprites, palettes, priorities := p.getSpritePixels(lineNumber)

		for i := range background {
			if window[i] < 4 {
				p.screenBuffer[int(lineNumber)*160+i] = window[i]
			} else {
				p.screenBuffer[int(lineNumber)*160+i] = background[i]
			}

			if sprites[i] > 0 {
				if priorities[i] || p.screenBuffer[int(lineNumber)*160+i] == 0 {
					p.screenBuffer[int(lineNumber)*160+i] = mapColorToPalette(palettes[i], sprites[i])
				}
			}
		}

		p.setControllerMode("HBlank")
		p.RunCPU(51 * 4) // HBlank allowance

		// do nothing
	}

	canRenderScreen <- struct{}{}
	p.writeLY(144)
	p.dispatchVBlankInterrupt()
	p.setControllerMode("VBlank")
	p.RunCPU(114 * 4) // VBlank row allowance

	for lineNumber := byte(145); lineNumber < 154; lineNumber++ {
		p.writeLY(lineNumber)
		p.RunCPU(114 * 4) // VBlank row allowance
	}
}

func getPixelColor(value byte) color.RGBA {
	white := color.RGBA{255, 255, 255, 255}
	lightgray := color.RGBA{192, 192, 192, 255}
	gray := color.RGBA{128, 128, 128, 255}
	black := color.RGBA{0, 0, 0, 255}

	switch value {
	case 3:
		return black
	case 2:
		return gray
	case 1:
		return lightgray
	case 0:
		return white
	default:
		panic("Got unexpected color")
	}
}

func (p *PPU) writeBufferToImage() {

	p.ImageMutex.Lock()
	defer p.ImageMutex.Unlock()

	for i := 0; i < 144; i++ {
		for j := 0; j < 160; j++ {
			p.Image.SetRGBA(j, i, getPixelColor(p.screenBuffer[i*160+j]))
		}
	}
}

func (p *PPU) renderer(canRenderScreenChan chan struct{}) {
	for range canRenderScreenChan {
		p.writeBufferToImage()
	}
}

func (p *PPU) Renderer() {

	canRenderScreenChan := make(chan struct{})
	frameTicker := time.NewTicker(time.Second / 60)

	go p.renderer(canRenderScreenChan)

	for range frameTicker.C {
		p.runEmulatorForAFrame(canRenderScreenChan)
	}
}

// for testing, run emulation as fast as possible
func (p *PPU) FastRenderer() {
	canRenderScreenChan := make(chan struct{})

	go p.renderer(canRenderScreenChan)

	for {
		p.runEmulatorForAFrame(canRenderScreenChan)
	}
}
