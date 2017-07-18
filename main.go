package main

import (
	// "fmt"
	"image"
	"image/color"
	"math"
	"os"

	mgl "github.com/go-gl/mathgl/mgl64"
	"github.com/llgcode/draw2d/draw2dimg"
	"github.com/sheenobu/go-obj/obj"
)

const width = 400
const height = 400

func triangle(r *image.RGBA, v0, v1, v2 *obj.Vertex, color color.Color) {
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

	go fill(v1, v0)
	go fill(v2, v1)
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

	light := mgl.Vec3{0, 0, 1}

	for _, face := range head.Faces {
		p1 := face.Points[0].Vertex
		p2 := face.Points[1].Vertex
		p3 := face.Points[2].Vertex

		v1 := mgl.Vec3{p1.X, p1.Y, p1.Z}
		v2 := mgl.Vec3{p2.X, p2.Y, p2.Z}
		v3 := mgl.Vec3{p3.X, p3.Y, p3.Z}

		intesity := light.Dot(v1.Sub(v2).Cross(v2.Sub(v3)).Normalize())

		if intesity < 0 {
			continue
		}

		colorComponent = uint8(intesity * 0xff)

		triangle(
			dest,
			screenSpace(p1),
			screenSpace(p2),
			screenSpace(p3),
			color.RGBA{
				colorComponent,
				colorComponent,
				colorComponent,
				0xff,
			},
		)
	}

	draw2dimg.SaveToPngFile("watcher/hello.png", dest)
}
