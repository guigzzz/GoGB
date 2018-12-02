package backend

import (
	"fmt"
	"image"
)

// PPU represents the pixel processing unit
// contains references to ram sections containing video relevant data
type PPU struct {
	// TilePatternT    []byte // Tile Pattern Table
	// BackgroundTileT []byte // Background Tile Table
	// WindowTileT     []byte // Window Tile Table
	// OAM             []byte // Object Attribute Memory
	ram        [1 << 16]byte   // reference to memory shared with CPU
	image      *image.RGBA     // represents the current screen buffer
	background [256 * 256]byte // contains the whole background
}

// NewPPU creates a new PPU object
func NewPPU(c CPU) PPU {
	p := PPU{}
	// p.TilePatternT = c.ram[0x8000:0x9800]    // 0x17FF bytes
	// p.BackgroundTileT = c.ram[0x9800:0x9C00] // 32 * 32 bytes
	// p.WindowTileT = c.ram[0x9C00:0xA000]     // 32 * 32 bytes
	// p.OAM = c.ram[0xFE00:0xFEA0]             // 40 * 4 bytes
	p.ram = c.ram
	p.image = image.NewRGBA(image.Rectangle{image.Point{0, 0}, image.Point{160, 144}})
	return p
}

func (p *PPU) DrawBackground() {
	// lcdc := p.ram[0xFF40] // LCD Control Register

	// // Bit 7 - LCD Display Enable             (0=Off, 1=On)
	// // Bit 6 - Window Tile Map Display Select (0=9800-9BFF, 1=9C00-9FFF)
	// // Bit 5 - Window Display Enable          (0=Off, 1=On)
	// // Bit 4 - BG & Window Tile Data Select   (0=8800-97FF, 1=8000-8FFF)
	// // Bit 3 - BG Tile Map Display Select     (0=9800-9BFF, 1=9C00-9FFF)
	// // Bit 2 - OBJ (Sprite) Size              (0=8x8, 1=8x16)
	// // Bit 1 - OBJ (Sprite) Display Enable    (0=Off, 1=On)
	// // Bit 0 - BG Display (for CGB see below) (0=Off, 1=On)

	// var mask byte = 1 << 3
	// var bgTileMapSelect []byte
	// if lcdc&mask == 0 {
	// 	bgTileMapSelect = p.ram[0x9800 : 0x9BFF+1]
	// } else {
	// 	bgTileMapSelect = p.ram[0x9C00 : 0x9FFF+1]
	// }

	// mask = 1 << 4
	// var bgTileMapData []byte
	// var indexOffset int
	// if lcdc&mask == 0 {
	// 	// index will be signed -128 -> 127
	// 	bgTileMapData = p.ram[0x8800 : 0x9BFF+1]
	// 	indexOffset = -128

	// } else {
	// 	// index will be unsigned 0 -> 255
	// 	bgTileMapData = p.ram[0x8000 : 0x8FFF+1]
	// 	indexOffset = 0
	// }

	// palette := p.ram[0xFF47]

	// for i, b := range bgTileMapSelect {
	// 	ind := int(b)
	// 	tile := bgTileMapData[(indexOffset+ind)*16 : (indexOffset+ind+1)*16] // 16 bytes per tile
	// }

}

func (p *PPU) DrawLine() {

	// fetch relevant memory
	// TilePatternT := p.ram[0x8000:0x9800]    // 0x17FF bytes
	// BackgroundTileT := p.ram[0x9800:0x9C00] // 32 * 32 bytes
	// WindowTileT := p.ram[0x9C00:0xA000]     // 32 * 32 bytes
	// OAM := p.ram[0xFE00:0xFEA0]             // 40 * 4 bytes

	// ScrollY := p.ram[0xFF42]
	// ScrollX := p.ram[0xFF43]

	fmt.Println(len(p.image.Pix) / 4)

	// p.image

}
