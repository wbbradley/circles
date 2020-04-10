package main

import (
	"image"
	"image/color"

	"github.com/llgcode/draw2d/draw2dimg"
	"github.com/llgcode/draw2d/draw2dkit"
)

type Circle struct {
	x, y, radius float64
}

func main() {
	var thickness float64 = 4.0
	width := float64(800.0)
	height := float64(800.0)
	rootCircle := &Circle{
		x:      width / 2.0,
		y:      height / 2.0,
		radius: width/40.0 - thickness*2.0,
	}

	// Initialize the graphic context on an RGBA image
	dest := image.NewRGBA(image.Rect(0, 0, int(width), int(height)))
	gc := draw2dimg.NewGraphicContext(dest)

	// Set some properties
	gc.SetFillColor(color.RGBA{0xff, 0xff, 0xff, 0xff})
	gc.BeginPath()
	gc.MoveTo(0, 0)
	gc.LineTo(width, 0)
	gc.LineTo(width, height)
	gc.LineTo(0, height)
	gc.Close()
	gc.Fill()

	gc.BeginPath()
	gc.SetLineWidth(thickness)
	gc.SetFillColor(color.RGBA{0xff, 0xcc, 0xdd, 0xff})
	gc.SetStrokeColor(color.RGBA{0x44, 0x44, 0x44, 0xff})
	draw2dkit.Circle(gc, rootCircle.x, rootCircle.y, rootCircle.radius)
	gc.Fill()

	// Save to file
	draw2dimg.SaveToPngFile("hello.png", dest)

}
