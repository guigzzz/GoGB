package main

import (
	"log"

	"github.com/guigzzz/GoGB/backend"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type Game struct {
	p *backend.PPU

	keyPressedMap map[string]bool
}

func (g *Game) Update() error {

	for key, value := range keyMap {
		if inpututil.IsKeyJustPressed(key) {
			g.keyPressedMap[value] = true
		}

		if inpututil.IsKeyJustReleased(key) {
			g.keyPressedMap[value] = false
		}
	}

	g.p.RunEmulatorForAFrame()

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	image := ebiten.NewImageFromImage(g.p.Image)
	screen.DrawImage(image, &ebiten.DrawImageOptions{})
}

const (
	width  = 160
	height = 144
)

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return width, height
}

func RunGame(p *backend.PPU, a *backend.APU, keyPressedMap map[string]bool) {
	game := &Game{p, keyPressedMap}

	audioContext := audio.NewContext(48000)
	player, err := audioContext.NewPlayer(a.ToReadCloser())
	if err != nil {
		panic(err)
	}
	player.SetVolume(1.0)

	go player.Play()

	ebiten.SetWindowSize(width*4, height*4)
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
