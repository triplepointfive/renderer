package main

import (
	"image"
	"image/color"
	"math"
	"os"

	"github.com/llgcode/draw2d/draw2dimg"
	"github.com/sheenobu/go-obj/obj"
)

const width = 400
const height = 400

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
	dest := image.NewRGBA(image.Rect(0, 0, width, height))
	gc := draw2dimg.NewGraphicContext(dest)

	gc.SetFillColor(color.RGBA{0x00, 0x00, 0x00, 0xff})
	gc.SetStrokeColor(color.RGBA{0xff, 0x00, 0x00, 0xff})
	gc.Clear()

	f, err := os.Open("tinyrenderer/obj/african_head/african_head.obj")
	if err != nil {
		panic(err)
	}

	head, err := obj.NewReader(f).Read()
	if err != nil {
		panic(err)
	}

	for _, face := range head.Faces {
		for i, point := range face.Points {
			v0, v1 := point.Vertex, face.Points[(i+1)%3].Vertex
			line(
				dest,
				(v0.X+1)*width/2,
				(v0.Y+1)*height/2,
				(v1.X+1)*width/2,
				(v1.Y+1)*height/2,
				white,
			)
		}
	}

	draw2dimg.SaveToPngFile("watcher/hello.png", dest)
}
