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

	for key, value := range keyMap {
		if inpututil.IsKeyJustPressed(key) {
			g.c.KeyPressedMap[value] = true
		}

		if inpututil.IsKeyJustReleased(key) {
			g.c.KeyPressedMap[value] = false
		}
	}

	g.p.RunEmulatorForAFrame()

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

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return width, height
}

func RunGame(p *backend.PPU, c *backend.CPU) {
	game := &Game{p: p, c: c}
	ebiten.SetWindowSize(width*2, height*2)
	ebiten.SetWindowTitle("GoGB")
	ebiten.SetMaxTPS(60)
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
