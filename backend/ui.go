package backend

import (
	"fmt"
	"image"
	"image/color"
	"log"

	"golang.org/x/exp/shiny/screen"
	"golang.org/x/mobile/event/key"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/paint"
)

var (
	width  = 160
	height = 144
)

// Run contains the main routine
func Run(s screen.Screen) {
	w, err := s.NewWindow(&screen.NewWindowOptions{
		Width: width, Height: height,
		Title: "GoGB",
	})
	if err != nil {
		log.Fatal(err)
	}
	defer w.Release()

	size0 := image.Point{width, height}
	b, err := s.NewBuffer(size0)
	if err != nil {
		log.Fatal(err)
	}
	defer b.Release()

	// var sz size.Event
	for {
		e := w.NextEvent()

		format := "got %#v\n"
		if _, ok := e.(fmt.Stringer); ok {
			format = "got %v\n"
		}
		fmt.Printf(format, e)

		switch e := e.(type) {
		case lifecycle.Event:
			if e.To == lifecycle.StageDead {
				return
			}

		case key.Event:
			if e.Code == key.CodeEscape {
				return
			}

		case paint.Event:
			drawGradient(b.RGBA())
			w.Upload(image.Point{0, 0}, b, b.Bounds())
			w.Publish()

		// case size.Event:
		// 	sz = e

		case error:
			log.Print(e)
		}
	}
}

func drawGradient(m *image.RGBA) {
	b := m.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			// if x%64 == 0 || y%64 == 0 {
			// 	m.SetRGBA(x, y, color.RGBA{0xff, 0xff, 0xff, 0xff})
			// } else if x%64 == 63 || y%64 == 63 {
			// 	m.SetRGBA(x, y, color.RGBA{0x00, 0x00, 0xff, 0xff})
			// } else {
			m.SetRGBA(x, y, color.RGBA{uint8(x), uint8(y), 0x00, 0xff})
			// }
		}
	}

	// Round off the corners.
	// const radius = 64
	// lox := b.Min.X + radius - 1
	// loy := b.Min.Y + radius - 1
	// hix := b.Max.X - radius
	// hiy := b.Max.Y - radius
	// for y := 0; y < radius; y++ {
	// 	for x := 0; x < radius; x++ {
	// 		if x*x+y*y <= radius*radius {
	// 			continue
	// 		}
	// 		m.SetRGBA(lox-x, loy-y, color.RGBA{})
	// 		m.SetRGBA(hix+x, loy-y, color.RGBA{})
	// 		m.SetRGBA(lox-x, hiy+y, color.RGBA{})
	// 		m.SetRGBA(hix+x, hiy+y, color.RGBA{})
	// 	}
	// }
}
