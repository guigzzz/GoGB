package main

import (
	"fmt"
	"image"
	"sync"
	"time"
	"unsafe"

	"github.com/go-gl/gl/v3.2-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/golang-ui/nuklear/nk"
	"github.com/guigzzz/GoGB/backend"
	"github.com/xlab/closer"
)

const (
	maxVertexBuffer  = 512 * 1024
	maxElementBuffer = 128 * 1024
)

type ScreenRenderer struct {
	win *glfw.Window
	ctx *nk.Context

	frameTex            uint32
	imageToDisplay      *image.RGBA
	imageToDisplayMutex *sync.RWMutex
}

func NewScreenRenderer(p *backend.PPU, width, height int) *ScreenRenderer {
	s := new(ScreenRenderer)

	s.imageToDisplay = p.Image
	s.imageToDisplayMutex = p.ImageMutex

	fmt.Println("init GLFW")
	if err := glfw.Init(); err != nil {
		closer.Fatalln(err)
	}
	glfw.WindowHint(glfw.ContextVersionMajor, 3)
	glfw.WindowHint(glfw.ContextVersionMinor, 2)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)

	fmt.Println("creating GLFW window")
	win, err := glfw.CreateWindow(width, height, "Nuklear Demo", nil, nil)
	if err != nil {
		closer.Fatalln(err)
	}
	win.MakeContextCurrent()

	s.win = win

	fmt.Printf("glfw: created window %dx%d\n", width, height)

	if err := gl.Init(); err != nil {
		closer.Fatalln("opengl: init failed:", err)
	}
	gl.Viewport(0, 0, int32(width), int32(height))

	fmt.Println("NK init")
	s.ctx = nk.NkPlatformInit(win, nk.PlatformInstallCallbacks)

	atlas := nk.NewFontAtlas()
	nk.NkFontStashBegin(&atlas)
	sansFont := nk.NkFontAtlasAddDefault(atlas, 16, nil)
	nk.NkFontStashEnd()
	if sansFont != nil {
		nk.NkStyleSetFont(s.ctx, sansFont.Handle())
	}

	return s
}

func (s *ScreenRenderer) startRendering() {
	exitC := make(chan struct{})
	doneC := make(chan struct{})
	closer.Bind(func() {
		close(exitC)
		<-doneC
	})

	fpsTicker := time.NewTicker(time.Second / 60)

	fmt.Println("starting to render...")

	for {
		select {
		case <-exitC:
			fmt.Println("exiting...")
			nk.NkPlatformShutdown()
			glfw.Terminate()
			fpsTicker.Stop()
			close(doneC)
			return
		case <-fpsTicker.C:
			if s.win.ShouldClose() {
				close(exitC)
				continue
			}
			glfw.PollEvents()
			s.displayFrame()
		}
	}
}

func (s *ScreenRenderer) displayFrame() {
	nk.NkPlatformNewFrame()

	// Layout
	bounds := nk.NkRect(0, 0, 160, 144)
	update := nk.NkBegin(s.ctx, "Demo", bounds, 0)

	if update > 0 {
		s.imageToDisplayMutex.RLock()

		frameImg := rgbaTex(&s.frameTex, s.imageToDisplay)
		nk.NkImage(s.ctx, frameImg)

		s.imageToDisplayMutex.RUnlock()
	}

	nk.NkEnd(s.ctx)

	// Render
	width, height := s.win.GetSize()
	gl.Viewport(0, 0, int32(width), int32(height))
	gl.Clear(gl.COLOR_BUFFER_BIT)
	nk.NkPlatformRender(nk.AntiAliasingOn, maxVertexBuffer, maxElementBuffer)
	s.win.SwapBuffers()
}

func rgbaTex(tex *uint32, rgba *image.RGBA) nk.Image {
	if *tex == 0 {
		gl.GenTextures(1, tex)
	}
	gl.BindTexture(gl.TEXTURE_2D, *tex)
	gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR_MIPMAP_NEAREST)
	gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR_MIPMAP_NEAREST)
	gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA8, int32(rgba.Bounds().Dx()), int32(rgba.Bounds().Dy()),
		0, gl.RGBA, gl.UNSIGNED_BYTE, unsafe.Pointer(&rgba.Pix[0]))
	gl.GenerateMipmap(gl.TEXTURE_2D)
	return nk.NkImageId(int32(*tex))
}
