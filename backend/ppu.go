package backend

import (
	"fmt"
	"image"
	"image/color"
	"sync"
	"time"
)

// Bit 7 - LCD Display Enable             (0=Off, 1=On)
// Bit 6 - Window Tile Map Display Select (0=9800-9BFF, 1=9C00-9FFF)
// Bit 5 - Window Display Enable          (0=Off, 1=On)
// Bit 4 - BG & Window Tile Data Select   (0=8800-97FF, 1=8000-8FFF)
// Bit 3 - BG Tile Map Display Select     (0=9800-9BFF, 1=9C00-9FFF)
// Bit 2 - OBJ (Sprite) Size              (0=8x8, 1=8x16)
// Bit 1 - OBJ (Sprite) Display Enable    (0=Off, 1=On)
// Bit 0 - BG Display (for CGB see below) (0=Off, 1=On)

// PPU represents the pixel processing unit
// contains references to ram sections containing video relevant data
type PPU struct {
	ram          []byte          // reference to memory shared with CPU
	Image        *image.RGBA     // represents the current screen
	ImageMutex   *sync.RWMutex   // to ensure safety when writing to screen buffer
	screenBuffer [144 * 160]byte // contains the pixels to draw on next refresh
	bus          *Bus
}

// NewPPU creates a new PPU object
func NewPPU(c *CPU, bus *Bus) *PPU {
	p := new(PPU)
	p.ram = c.ram
	p.Image = image.NewRGBA(image.Rectangle{image.Point{0, 0}, image.Point{160, 144}})
	p.ImageMutex = new(sync.RWMutex)
	p.bus = bus
	return p
}

const (
	bgDisplay                  = iota // (0=Off, 1=On)
	bgJDisplayEnable                  // (0=Off, 1=On)
	objSize                           // (0=8x8, 1=8x16)
	bgTileMapDisplaySelect            // (0=9800-9BFF, 1=9C00-9FFF)
	bgWindowTileDataSelect            // (0=8800-97FF, 1=8000-8FFF)
	windowDisplayEnable               // (0=Off, 1=On)
	windowTileMapDisplaySelect        // (0=9800-9BFF, 1=9C00-9FFF)
	lcdDisplayEnable                  // (0=Off, 1=On)
)

func (p *PPU) lcdControlRegisterIsBitSet(bitnum uint) bool {
	if bitnum > 7 {
		panic(fmt.Sprintf("Got unexpected bit number %d higher than 7 (max for byte)", bitnum))
	}
	return p.ram[0xFF40]&(1<<bitnum) > 0
}

func (p *PPU) getBackgroundTileData() ([]byte, bool) {
	if p.lcdControlRegisterIsBitSet(bgWindowTileDataSelect) {
		return p.ram[0x8000 : 0x8FFF+1], false
	}
	return p.ram[0x8800 : 0x97FF+1], true
}

func (p *PPU) getWindowTileData() ([]byte, bool) {
	return p.getBackgroundTileData()
}

func (p *PPU) getBackgroundTileMap() []byte {
	if p.lcdControlRegisterIsBitSet(bgTileMapDisplaySelect) {
		return p.ram[0x9C00 : 0x9FFF+1]
	}
	return p.ram[0x9800 : 0x9BFF+1]
}

func (p *PPU) getWindowTileMap() []byte {
	if p.lcdControlRegisterIsBitSet(windowTileMapDisplaySelect) {
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
	if !p.lcdControlRegisterIsBitSet(bgDisplay) {
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

	pixels := [160]byte{}
	if !p.lcdControlRegisterIsBitSet(windowDisplayEnable) {
		return pixels
	}

	yPos, xPos := p.getWindowPosition()

	tileMap := p.getWindowTileMap()
	tileData, interpretIndexAsSigned := p.getWindowTileData()

	rowInTile := (yPos + lineNumber) % 8
	tileRow := (yPos + lineNumber) / 8

	for i := byte(0); i < 160; i++ {

		// compute in which background tile we fall in (in a 32 x 32 grid)
		tileColumn := (xPos + i) / 8
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

		pixelInLine := (xPos + i) % 8
		msb := (lineData[1] >> (7 - pixelInLine)) & 1
		lsb := (lineData[0] >> (7 - pixelInLine)) & 1

		colorCode := (msb << 1) | lsb

		pixels[i] = mapColorToPalette(p.getBGPalette(), colorCode)
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
	if p.lcdControlRegisterIsBitSet(objSize) {
		return 16
	}
	return 8
}

func (p *PPU) getSpritePixels(lineNumber byte) [160]byte {

	// Implements OAM searching and sprite rendering
	colorCodes := [160]byte{}
	palettes := [160]byte{}

	attributes := p.getSpriteAttributes()
	tileData := p.getSpriteData()
	spriteHeight := p.getSpriteHeight()

	numSprites := 0

	for i := 0; i < 40 && numSprites < 10; i++ {

		yPos := attributes[4*i]
		if yPos > lineNumber+16 || lineNumber+16 >= yPos+spriteHeight {
			continue
		}

		xPos := attributes[4*i+1]
		if xPos == 0 {
			continue
		}

		numSprites++
		tileIndex := attributes[4*i+2]
		flags := attributes[4*i+3]
		xFlipped := flags&0x20 > 0
		yFlipped := flags&0x40 > 0
		palette := p.getSpritePalette(flags)

		var rowInTile byte
		if yPos < 16 {
			rowInTile = 16 - yPos + lineNumber
		} else {
			rowInTile = lineNumber - (yPos - 16)
		}

		if spriteHeight == 16 {
			if rowInTile >= 8 {
				tileIndex |= 1
				rowInTile -= 8
			} else {
				tileIndex &= 0xFE
			}
		}

		lineDataIndex := uint(tileIndex)*16 + 2*uint(rowInTile)
		if yFlipped {
			lineDataIndex = uint(tileIndex)*16 + 2*7 - 2*uint(rowInTile)
		}
		lineData := tileData[lineDataIndex : lineDataIndex+2]
		if xFlipped {
			lineData[0] = reverse(lineData[0])
			lineData[1] = reverse(lineData[1])
		}

		for l := 0; l < 8; l++ {
			if xPos < 8-byte(l) {
				continue
			}
			msb := (lineData[1] >> (7 - byte(l))) & 1
			lsb := (lineData[0] >> (7 - byte(l))) & 1

			colorCode := (msb << 1) | lsb

			pos := xPos - 8 + byte(l)
			if pos <= 159 && colorCode > 0 && colorCodes[pos] == 0 {

				colorCodes[pos] = colorCode
				palettes[pos] = palette

			}
		}
	}

	pixels := [160]byte{}
	for i, c := range colorCodes {
		if c > 0 {
			pixels[i] = mapColorToPalette(palettes[i], c)
		}
	}

	// sprite data @ 8000-8FFF
	// sprite attributes in OAM @ OAM @ FE00-FE9F
	// OAM is divided into 40 4-byte blocks each of which corresponds to a sprite

	// Byte0  Y position on the screen
	// Byte1  X position on the screen
	// Byte2  Pattern number 0-255 [notice that unlike tile numbers, sprite
	// 		pattern numbers are unsigned]
	// Byte3  Flags:
	// 		Bit7  Priority
	// 			Sprite is displayed in front of the window if this bit
	// 			is set to 1. Otherwise, sprite is shown behind the
	// 			window but in front of the background.
	// 		Bit6  Y flip
	// 			Sprite pattern is flipped vertically if this bit is
	// 			set to 1.
	// 		Bit5  X flip
	// 			Sprite pattern is flipped horizontally if this bit is
	// 			set to 1.
	// 		Bit4  Palette number
	// 			Sprite colors are taken from OBJ1PAL (FF49) if this bit is
	// 			set to 1 and from OBJ0PAL (FF48) otherwise.

	// Todo check sprite size
	// if in 8x16 mode then pattern number of upper tile is pattern_number & 0xFE
	// pattern number of lower tile is pattern_number | 0x01

	// For each sprite in RAM, check if it has pixels that need to be drawn on that line
	// and that we have so far drawn less than 10

	return pixels
}

func reverse(in byte) byte {
	return (in & 0x1 << 7) | (in & 0x2 << 5) | (in & 0x4 << 3) | (in & 0x8 << 1) |
		(in & 0x10 >> 1) | (in & 0x20 >> 3) | (in & 0x40 >> 5) | (in & 0x80 >> 7)
}

func (p *PPU) writeLY(lineNumber byte) {
	p.ram[0xFF44] = lineNumber

	if lineNumber == p.ram[0xFF45] { // CMPLINE
		p.ram[0xFF41] |= 1 << 2

		// If scanline coincidence interrupt is enabled
		if p.ram[0xFF41]&0x40 > 0 {
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

func (p *PPU) setControllerMode(mode string) {
	switch mode {
	case "VBlank":
		p.dispatchVBlankInterrupt()
		p.ram[0xFF41] = 0x80 | p.ram[0xFF41]&0x7C | 0x1
	case "HBlank":
		p.ram[0xFF41] = 0x80 | p.ram[0xFF41]&0x7C | 0x0
		if p.ram[0xFF41]&0x8 > 0 {
			p.dispatchLCDStatInterrupt()
		}
	case "OAM":
		p.ram[0xFF41] = 0x80 | p.ram[0xFF41]&0x7C | 0x2
		// if p.ram[0xFF41]&0x20 > 0 {
		// 	p.dispatchLCDStatInterrupt()
		// }
	case "PixelTransfer":
		p.ram[0xFF41] = 0x80 | p.ram[0xFF41]&0x7C | 0x3
	default:
		panic(fmt.Sprintln("Got unexpected LCD controller mode: ", mode))
	}
}

func (p *PPU) lineByLineRender(frameTicker *time.Ticker, canRenderScreen chan struct{}) {

	for range frameTicker.C {

		for lineNumber := byte(0); lineNumber < 144; lineNumber++ {

			<-p.bus.cpuDoneChannel
			p.writeLY(lineNumber)
			p.setControllerMode("OAM")
			p.bus.allowanceChannel <- 20 * 4 // OAM search allowance

			// OAM search

			<-p.bus.cpuDoneChannel
			p.setControllerMode("PixelTransfer")
			p.bus.allowanceChannel <- 43 * 4 // pixel transfer allowance

			// pixel transfer
			background := p.getBackgroundPixels(lineNumber)
			// window pixels := getWindowPixels
			spritePixels := p.getSpritePixels(lineNumber)

			for i := range background {
				if spritePixels[i] > 0 {
					p.screenBuffer[int(lineNumber)*160+i] = spritePixels[i]
				} else {
					p.screenBuffer[int(lineNumber)*160+i] = background[i]
				}
			}

			<-p.bus.cpuDoneChannel
			p.setControllerMode("HBlank")
			p.bus.allowanceChannel <- 51 * 4 // HBlank allowance

			// do nothing
		}

		canRenderScreen <- struct{}{}
		<-p.bus.cpuDoneChannel
		p.writeLY(144)
		p.dispatchVBlankInterrupt()
		p.setControllerMode("VBlank")
		p.bus.allowanceChannel <- 114 * 4 // VBlank row allowance

		for lineNumber := byte(145); lineNumber < 154; lineNumber++ {
			<-p.bus.cpuDoneChannel
			p.writeLY(lineNumber)
			p.bus.allowanceChannel <- 114 * 4 // VBlank row allowance
		}
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

func (p *PPU) Renderer() {

	canRenderScreenChan := make(chan struct{})
	frameTicker := time.NewTicker(16666 * time.Microsecond)

	go p.lineByLineRender(frameTicker, canRenderScreenChan)

	for range canRenderScreenChan {
		p.writeBufferToImage()
	}
}

func (p *PPU) DumpBackground() image.Image {
	image := image.NewRGBA(image.Rectangle{image.Point{0, 0}, image.Point{255, 255}})

	tileMap := p.getBackgroundTileMap()
	tileData, interpretIndexAsSigned := p.getBackgroundTileData()

	for i := 0; i < 32; i++ {
		for j := 0; j < 32; j++ {

			index := i*32 + j
			tileMapIndex := tileMap[index]

			if interpretIndexAsSigned {
				tileMapIndex ^= 0x80
			}

			data := tileData[uint(tileMapIndex)*16 : uint(tileMapIndex+1)*16]

			for k := 0; k < 8; k++ {

				lineData := data[k*2 : (k+1)*2]

				for l := 0; l < 8; l++ {

					msb := (lineData[1] >> (7 - byte(l))) & 1
					lsb := (lineData[0] >> (7 - byte(l))) & 1

					colorCode := (msb << 1) | lsb

					pixel := mapColorToPalette(p.getBGPalette(), colorCode)
					image.SetRGBA(j*8+l, i*8+k, getPixelColor(pixel))
				}
			}

		}
	}

	scrollY, scrollX := p.getScroll()

	red := color.RGBA{255, 0, 0, 255}
	for i := byte(0); i < 160; i++ {
		image.SetRGBA(int(scrollX+i), int(scrollY), red)
		image.SetRGBA(int(scrollX+i), int(scrollY+143), red)
	}
	for j := byte(0); j < 144; j++ {
		image.SetRGBA(int(scrollX), int(scrollY+j), red)
		image.SetRGBA(int(scrollX+159), int(scrollY+j), red)
	}

	return image
}
