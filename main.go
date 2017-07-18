package main

import (
	"image"
	"image/color"
	"math"

	"github.com/llgcode/draw2d/draw2dimg"
)

func line(r *image.RGBA, x0, y0, x1, y1 float64, color color.Color) {
	steep := false
	if math.Abs(x0-x1) < math.Abs(y0-y1) {
		x0, y0 = y0, x0
		x1, y1 = y1, x1
		steep = true
	}

	if x0 > x1 {
		x0, x1 = x1, x0
		y0, y1 = y1, y0
	}

	dx := int(x1 - x0)
	dy := int(y1 - y0)
	derror2 := int(math.Abs(float64(dy)) * 2)
	error2 := 0
	y := int(y0)

	for x := x0; x <= x1; x++ {
		if steep {
			r.Set(y, int(x), color)
		} else {
			r.Set(int(x), y, color)
		}

		error2 += derror2
		if error2 > dx {
			if y1 > y0 {
				y++
			} else {
				y--
			}
			error2 -= dx * 2
		}
	}
}

var red = color.RGBA{0xff, 0x00, 0x00, 0xff}
var white = color.RGBA{0xff, 0xff, 0xff, 0xff}

func main() {
	dest := image.NewRGBA(image.Rect(0, 0, 100, 100))
	gc := draw2dimg.NewGraphicContext(dest)

	gc.SetFillColor(color.RGBA{0x00, 0x00, 0x00, 0xff})
	gc.SetStrokeColor(color.RGBA{0xff, 0x00, 0x00, 0xff})
	gc.Clear()

	line(dest, 13, 20, 80, 40, white)
	line(dest, 20, 13, 40, 80, red)
	line(dest, 80, 40, 13, 20, red)
	draw2dimg.SaveToPngFile("watcher/hello.png", dest)
}
