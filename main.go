package main

import (
	"image"
	"image/color"
	"image/draw"
	"log"
	"unsafe"

	"github.com/go-gl/gl/v3.2-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/golang-ui/nuklear/nk"
	"github.com/guigzzz/GoGB/backend"
)

const (
	winWidth  = 160
	winHeight = 144

	maxVertexBuffer  = 512 * 1024
	maxElementBuffer = 128 * 1024
)

func main() {

	cpu := backend.NewCPU()
	ppu := backend.NewPPU(cpu)
	ppu.Renderer()

	// runtime.LockOSThread()

	// // cpu := backend.NewCPU()
	// // ppu := backend.NewPPU(cpu)

	// if err := glfw.Init(); err != nil {
	// 	closer.Fatalln(err)
	// }
	// glfw.WindowHint(glfw.ContextVersionMajor, 3)
	// glfw.WindowHint(glfw.ContextVersionMinor, 2)
	// glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	// glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	// win, err := glfw.CreateWindow(winWidth, winHeight, "Nuklear Demo", nil, nil)
	// if err != nil {
	// 	closer.Fatalln(err)
	// }
	// win.MakeContextCurrent()

	// width, height := win.GetSize()
	// log.Printf("glfw: created window %dx%d", width, height)

	// if err := gl.Init(); err != nil {
	// 	closer.Fatalln("opengl: init failed:", err)
	// }
	// gl.Viewport(0, 0, int32(width), int32(height))

	// ctx := nk.NkPlatformInit(win, nk.PlatformInstallCallbacks)

	// atlas := nk.NewFontAtlas()
	// nk.NkFontStashBegin(&atlas)
	// sansFont := nk.NkFontAtlasAddDefault(atlas, 16, nil)
	// nk.NkFontStashEnd()
	// if sansFont != nil {
	// 	nk.NkStyleSetFont(ctx, sansFont.Handle())
	// }

	// exitC := make(chan struct{}, 1)
	// doneC := make(chan struct{}, 1)
	// closer.Bind(func() {
	// 	close(exitC)
	// 	<-doneC
	// })

	// state := &State{
	// 	bgColor: nk.NkRgba(28, 48, 62, 255),
	// }
	// fpsTicker := time.NewTicker(time.Second / 30)

	// frameTex := uint32(0)
	// for {
	// 	select {
	// 	case <-exitC:
	// 		nk.NkPlatformShutdown()
	// 		glfw.Terminate()
	// 		fpsTicker.Stop()
	// 		close(doneC)
	// 		return
	// 	case <-fpsTicker.C:
	// 		if win.ShouldClose() {
	// 			close(exitC)
	// 			continue
	// 		}
	// 		glfw.PollEvents()
	// 		gfxMain(win, ctx, state, &frameTex)
	// 	}
	// }
}

func gfxMain(win *glfw.Window, ctx *nk.Context, state *State, frameTex *uint32) {
	nk.NkPlatformNewFrame()

	// Layout
	bounds := nk.NkRect(0, 0, 160, 144)
	update := nk.NkBegin(ctx, "Demo", bounds, 0)

	rgba := image.NewRGBA(image.Rectangle{image.Point{0, 0}, image.Point{160, 144}})
	draw.Draw(rgba, rgba.Bounds(), &image.Uniform{color.RGBA{0, 255, 0, 255}}, image.ZP, draw.Src)

	if update > 0 {
		frameImg := rgbaTex(frameTex, rgba)
		nk.NkImage(ctx, frameImg)
		*frameTex++
	}

	nk.NkEnd(ctx)

	// Render
	width, height := win.GetSize()
	gl.Viewport(0, 0, int32(width), int32(height))
	gl.Clear(gl.COLOR_BUFFER_BIT)
	nk.NkPlatformRender(nk.AntiAliasingOn, maxVertexBuffer, maxElementBuffer)
	win.SwapBuffers()
}

func rgbaTex(tex *uint32, rgba *image.RGBA) nk.Image {
	if tex == nil {
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

type Option uint8

const (
	Easy Option = 0
	Hard Option = 1
)

type State struct {
	bgColor nk.Color
	prop    int32
	opt     Option
}

func onError(code int32, msg string) {
	log.Printf("[glfw ERR]: error %d: %s", code, msg)
}

func flag(v bool) int32 {
	if v {
		return 1
	}
	return 0
}
