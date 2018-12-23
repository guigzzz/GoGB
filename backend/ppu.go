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
	ram          [1 << 16]byte   // reference to memory shared with CPU
	Image        *image.RGBA     // represents the current screen
	ImageMutex   *sync.RWMutex   // to ensure safety when writing to screen buffer
	screenBuffer [144 * 160]byte // contains the pixels to draw on next refresh
}

// NewPPU creates a new PPU object
func NewPPU(c CPU) *PPU {
	p := new(PPU)
	p.ram = c.ram
	p.Image = image.NewRGBA(image.Rectangle{image.Point{0, 0}, image.Point{160, 144}})
	p.ImageMutex = new(sync.RWMutex)
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

func (p *PPU) getBackgroundPixels(lineNumber byte) [144]byte {

	tileMap := p.getBackgroundTileMap()
	tileData, interpretIndexAsSigned := p.getBackgroundTileData()

	scrollY, scrollX := p.getScroll()
	rowInTile := scrollY % 8
	tileRow := (scrollY + lineNumber) / 8
	pixels := [144]byte{}

	for i := byte(0); i < 144; i++ {

		// compute in which background tile we fall in (in a 32 x 32 grid)
		tileColumn := (scrollX + i) / 8
		tileIndex := tileRow*32 + tileColumn

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
		lineDataIndex := tileMapIndex*16 + 2*rowInTile
		lineData := tileData[lineDataIndex : lineDataIndex+2]

		msb := lineData[1] >> (6 - i%8)
		lsb := lineData[0] >> (7 - i%8)

		pixels[i] = mapColorToPalette(p.getBGPalette(), msb|lsb)
	}

	return pixels
}

func (p *PPU) getWindowPixels() {

}

func (p *PPU) getSpritePixels() {

}

func (p *PPU) lineByLineRender(canRenderLine *time.Ticker, canRenderScreen chan struct{}) {

	// 114 clocks per line
	// OAM search 20 clocks
	// pixel transfer 43 blocks
	// hblank 51 clocks

	// 144 lines + 10 vblank

	lineNumber := byte(0)

	for range canRenderLine.C {

		switch {
		case lineNumber < 144:
			pixels := p.getBackgroundPixels(lineNumber)
			// window pixels := getWindowPixels
			// spritePixels := getSpritePixels

			for i, pixel := range pixels {
				p.screenBuffer[int(lineNumber)*160+i] = pixel
			}
			lineNumber++

			if lineNumber == 144 {
				canRenderScreen <- struct{}{}
			}

		case lineNumber < 154:
			lineNumber++

		case lineNumber == 154:
			lineNumber = 0
		}
	}
}

func (p *PPU) writeBufferToImage() {
	white := color.RGBA{255, 255, 255, 255}
	lightgray := color.RGBA{192, 192, 192, 255}
	gray := color.RGBA{128, 128, 128, 255}
	black := color.RGBA{0, 0, 0, 255}

	p.ImageMutex.Lock()
	defer p.ImageMutex.Unlock()

	for i := 0; i < 144; i++ {
		for j := 0; j < 160; j++ {
			switch p.screenBuffer[i*160+j] {
			case 3:
				p.Image.SetRGBA(j, i, black)
			case 2:
				p.Image.SetRGBA(j, i, gray)
			case 1:
				p.Image.SetRGBA(j, i, lightgray)
			case 0:
				p.Image.SetRGBA(j, i, white)
			default:
				panic("Got unexpected color")
			}
		}
	}

	red := color.RGBA{255, 0, 0, 255}
	for j := 0; j < 160; j++ {
		p.Image.SetRGBA(j, 0, red)
		p.Image.SetRGBA(j, 1, red)
		p.Image.SetRGBA(j, 142, red)
		p.Image.SetRGBA(j, 143, red)
	}
	for i := 0; i < 144; i++ {
		p.Image.SetRGBA(0, i, red)
		p.Image.SetRGBA(1, i, red)
		p.Image.SetRGBA(158, i, red)
		p.Image.SetRGBA(159, i, red)
	}
}

func (p *PPU) Renderer() {

	canRenderScreenChan := make(chan struct{})
	lineTicker := time.NewTicker(108719 * time.Nanosecond)

	go p.lineByLineRender(lineTicker, canRenderScreenChan)

	for range canRenderScreenChan {
		p.writeBufferToImage()
	}
}
