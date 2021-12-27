package main

import (
	"fmt"
	"log"

	"github.com/guigzzz/GoGB/backend"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type Game struct {
	p *backend.PPU
	c *backend.CPU
}

func (g *Game) Update() error {

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
	image := ebiten.NewImageFromImage(g.p.Image)
	screen.DrawImage(image, &ebiten.DrawImageOptions{})

	msg := fmt.Sprintf("TPS: %0.2f", ebiten.CurrentTPS())
	ebitenutil.DebugPrint(screen, msg)
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

	audioContext := audio.NewContext(48000)
	player, err := audioContext.NewPlayer(c.GetAPU().ToReadCloser())
	if err != nil {
		panic(err)
	}
	player.SetVolume(1.0)

	go player.Play()

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
