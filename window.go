package main

import (
	"log"

	"github.com/guigzzz/GoGB/backend"
	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/inpututil"
)

type Game struct {
	p *backend.PPU
	c *backend.CPU
}

func (g *Game) Update(screen *ebiten.Image) error {

	g.c.KeyPressedMapLock.Lock()
	defer g.c.KeyPressedMapLock.Unlock()

	for key, value := range keyMap {
		if inpututil.IsKeyJustPressed(key) {
			g.c.KeyPressedMap[value] = true
		}

		if inpututil.IsKeyJustReleased(key) {
			g.c.KeyPressedMap[value] = false
		}
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	image, _ := ebiten.NewImageFromImage(g.p.Image, ebiten.FilterDefault)
	screen.DrawImage(image, &ebiten.DrawImageOptions{})
}

const (
	width  = 160
	height = 144
)

// Layout takes the outside size (e.g., the window size) and returns the (logical) screen size.
// If you don't have to adjust the screen size with the outside size, just return a fixed size.
func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return width, height
}

func RunGame(p *backend.PPU, c *backend.CPU) {
	game := &Game{p: p, c: c}
	// Specify the window size as you like. Here, a doubled size is specified.
	ebiten.SetWindowSize(width*2, height*2)
	ebiten.SetWindowTitle("GoGB")
	// Call ebiten.RunGame to start your game loop.
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

var keyMap = map[ebiten.Key]string{
	ebiten.KeyS: "down",
	ebiten.KeyW: "up",
	ebiten.KeyA: "left",
	ebiten.KeyD: "right",

	ebiten.KeyU: "start",
	ebiten.KeyI: "select",
	ebiten.KeyK: "B",
	ebiten.KeyJ: "A",
}
