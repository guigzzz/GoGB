package main

import (
	"log"

	"github.com/guigzzz/GoGB/backend"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type Game struct {
	e               *backend.Emulator
	speedMultiplier float32
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

	if inpututil.IsKeyJustPressed(ebiten.KeyX) {
		g.UpdateMaxTps(0.1)
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyZ) {
		g.UpdateMaxTps(-0.1)
	}

	emu.RunForAFrame()

	return nil
}

func (g *Game) UpdateMaxTps(increment float32) {
	g.speedMultiplier += increment
	println("New speed: ", g.speedMultiplier)
	ebiten.SetMaxTPS((int)(g.speedMultiplier * 60))
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
	game := &Game{e: emu, speedMultiplier: 1}

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
