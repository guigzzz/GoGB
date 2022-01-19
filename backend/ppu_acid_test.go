package backend

import (
	"image"
	"image/draw"
	"image/png"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

const romPath = "../rom/dmg-acid2.gb"
const outDmgAcid = "out/dmg-acid.png"
const refDmgAcid = "ref/dmg-acid.png"

func imageToRGBA(src image.Image) *image.RGBA {

	// No conversion needed if image is an *image.RGBA.
	if dst, ok := src.(*image.RGBA); ok {
		return dst
	}

	// Use the image/draw package to convert to *image.RGBA.
	b := src.Bounds()
	dst := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(dst, dst.Bounds(), src, b.Min, draw.Src)
	return dst
}

func getImage(path string) *image.RGBA {
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}

	defer f.Close()

	image, err := png.Decode(f)
	if err != nil {
		panic(err)
	}

	return imageToRGBA(image)
}

func TestRunDmgAcid2(t *testing.T) {

	rom, err := ioutil.ReadFile(romPath)
	if err != nil {
		panic(err)
	}

	cpu := NewCPU(rom, false, NewNullLogger(), NullApuFactory)
	ppu := NewPPU(cpu)

	for i := 0; i < 100; i++ {
		ppu.RunEmulatorForAFrame()
	}

	ref := getImage(refDmgAcid)
	if !assert.Equal(t, ppu.Image.Pix, ref.Pix) {
		ppu.dumpScreenToPng(outDmgAcid)
	}
}
