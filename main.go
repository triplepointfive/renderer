package main

import (
	"fmt"
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

func triangle(r *image.RGBA, v0, v1, v2 *obj.Vertex, color color.Color) {
	fmt.Println(v0.Y)
	if v0.Y < v1.Y {
		v0, v1 = v1, v0
	}

	if v0.Y < v2.Y {
		v0, v2 = v2, v0
	}

	if v1.Y < v2.Y {
		v2, v1 = v1, v2
	}

	fill := func(t0, t1 *obj.Vertex) {
		for y := t0.Y; y <= t1.Y; y++ {
			x1, x2 := onSide(y, t1, t0), onSide(y, v2, v0)
			for x := math.Min(x1, x2); x <= math.Max(x1, x2); x++ {
				r.Set(int(x), int(y), color)
			}
		}
	}

	fill(v1, v0)
	fill(v2, v1)
}

func onSide(y float64, v1, v2 *obj.Vertex) float64 {
	return ((v2.X-v1.X)*y + (v1.X*v2.Y - v2.X*v1.Y)) / (v2.Y - v1.Y)
}

var red = color.RGBA{0xff, 0x00, 0x00, 0xff}
var green = color.RGBA{0x00, 0xff, 0x00, 0xff}
var white = color.RGBA{0xff, 0xff, 0xff, 0xff}

func screenSpace(v *obj.Vertex) *obj.Vertex {
	return &obj.Vertex{X: (v.X + 1) * width / 2, Y: (v.Y + 1) * height / 2}
}

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
		triangle(
			dest,
			screenSpace(face.Points[0].Vertex),
			screenSpace(face.Points[1].Vertex),
			screenSpace(face.Points[2].Vertex),
			white,
		)
	}

	draw2dimg.SaveToPngFile("watcher/hello.png", dest)
}
