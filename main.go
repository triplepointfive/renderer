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

var zBuffer = [][]float64{}

func triangle(r *image.RGBA, v0, v1, v2 *mgl.Vec3, color color.Color) {
	if v0.Y() < v1.Y() {
		v0, v1 = v1, v0
	}

	if v0.Y() < v2.Y() {
		v0, v2 = v2, v0
	}

	if v1.Y() < v2.Y() {
		v2, v1 = v1, v2
	}

	fill := func(t0, t1 *mgl.Vec3) {
		for y := t0.Y(); y <= t1.Y(); y++ {
			x1, x2 := onSide(y, t1, t0), onSide(y, v2, v0)
			for x := math.Min(x1, x2); x <= math.Max(x1, x2); x++ {
				z := centric(x, y, v0, v1, v2).Z()
				if zBuffer[int(x)][int(y)] < z {
					zBuffer[int(x)][int(y)] = z
					r.Set(int(x), int(y), color)
				}
			}
		}
	}

	fill(v1, v0)
	fill(v2, v1)
}

func squareDistance(v0, v1 *mgl.Vec3) float64 {
	return v0.X()*v1.X() + v0.Y()*v1.Y() + v0.Z()*v1.Z()
}

func centric(x, y float64, v0, v1, v2 *mgl.Vec3) (v *mgl.Vec3) {
	v = &mgl.Vec3{x, y, 0}

	d0 := squareDistance(v, v0)
	d1 := squareDistance(v, v1)
	d2 := squareDistance(v, v2)

	v[2] = (d0*v0.Z() + d1*v1.Z() + d2*v2.Z()) / (d0 + d1 + d2)

	return
}

func onSide(y float64, v1, v2 *mgl.Vec3) float64 {
	return ((v2.X()-v1.X())*y + (v1.X()*v2.Y() - v2.X()*v1.Y())) / (v2.Y() - v1.Y())
}

var red = color.RGBA{0xff, 0x00, 0x00, 0xff}
var green = color.RGBA{0x00, 0xff, 0x00, 0xff}
var white = color.RGBA{0xff, 0xff, 0xff, 0xff}

func screenSpace(v *mgl.Vec3) *mgl.Vec3 {
	return &mgl.Vec3{
		(v.X() + 1) * width / 2,
		(v.Y() + 1) * height / 2,
		(v.Z() + 1) / 2,
	}
}

func main() {
	for i := 0; i < width; i++ {
		zBuffer = append(zBuffer, make([]float64, height))
	}

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

		v1 := &mgl.Vec3{p1.X, p1.Y, p1.Z}
		v2 := &mgl.Vec3{p2.X, p2.Y, p2.Z}
		v3 := &mgl.Vec3{p3.X, p3.Y, p3.Z}

		intesity := light.Dot(v1.Sub(*v2).Cross(v2.Sub(*v3)).Normalize())

		if intesity < 0 {
			continue
		}

		colorComponent := uint8(intesity * 0xff)

		triangle(
			dest,
			screenSpace(v1),
			screenSpace(v2),
			screenSpace(v3),
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
