package main

import (
	"log"

	"github.com/guigzzz/GoGB/backend"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type Game struct {
	e *backend.Emulator
}

func (g *Game) Update() error {

	emu := g.e

	for key, value := range keyMap {
		if inpututil.IsKeyJustPressed(key) {
			emu.SetKeyIsPressed(value, true)
		}

		if inpututil.IsKeyJustReleased(key) {
			emu.SetKeyIsPressed(value, false)
		}
	}

	emu.RunForAFrame()

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	image := ebiten.NewImageFromImage(g.e.GetImage())
	screen.DrawImage(image, &ebiten.DrawImageOptions{})
}

const (
	width  = 160
	height = 144
)

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return width, height
}

func RunGame(emu *backend.Emulator) {
	game := &Game{emu}

	audioContext := audio.NewContext(48000)
	player, err := audioContext.NewPlayer(emu.GetAudioStream())
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
