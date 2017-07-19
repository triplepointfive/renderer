package main

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"os"

	mgl "github.com/go-gl/mathgl/mgl64"
	"github.com/llgcode/draw2d/draw2dimg"
	"github.com/sheenobu/go-obj/obj"
)

const (
	width  = 400
	height = 400
)

var zBuffer = [][]float64{}
var light = mgl.Vec3{0, 0, 1}

var (
	red   = color.RGBA{0xff, 0x00, 0x00, 0xff}
	green = color.RGBA{0x00, 0xff, 0x00, 0xff}
	white = color.RGBA{0xff, 0xff, 0xff, 0xff}
)

type (
	// Vertex -
	Vertex struct {
		Pos    *mgl.Vec3
		Normal *mgl.Vec3
	}

	// Face -
	Face [3]*Vertex

	// Program -
	Program struct {
		Screen *image.RGBA
	}

	// VertexOut -
	VertexOut struct {
		Position       *mgl.Vec3
		LightIntensity float64
	}
)

// Run -
func (program *Program) Run(faces []*Face) {
	for _, face := range faces {
		m0 := vertexShader(face[0])
		m1 := vertexShader(face[1])
		m2 := vertexShader(face[2])

		if m0.Position.Y() < m1.Position.Y() {
			m0, m1 = m1, m0
		}

		if m0.Position.Y() < m2.Position.Y() {
			m0, m2 = m2, m0
		}

		if m1.Position.Y() < m2.Position.Y() {
			m2, m1 = m1, m2
		}

		fill := func(t0, t1 *VertexOut) {
			for y := t0.Position.Y(); y <= t1.Position.Y(); y++ {
				x1, x2 := onSide(y, t1.Position, t0.Position), onSide(y, m2.Position, m0.Position)
				for x := math.Min(x1, x2); x <= math.Max(x1, x2); x++ {
					m := centric(x, y, m0, m1, m2)
					z := m.Position.Z()

					if zBuffer[int(x)][int(y)] < z {
						if fragColor := fragmentShader(m); fragColor != nil {
							zBuffer[int(x)][int(y)] = z
							program.Screen.Set(
								int(x),
								int(y),
								color.RGBA{
									uint8(fragColor.X() * 0xff),
									uint8(fragColor.Y() * 0xff),
									uint8(fragColor.Z() * 0xff),
									uint8(fragColor.W() * 0xff),
								},
							)
						}
					}
				}
			}
		}

		fill(m1, m0)
		fill(m2, m1)
	}
}

func vertexShader(v *Vertex) (out *VertexOut) {
	out = &VertexOut{}
	out.Position = &mgl.Vec3{
		(v.Pos.X() + 1) * width / 2,
		(v.Pos.Y() + 1) * height / 2,
		(v.Pos.Z() + 1) / 2,
	}

	out.LightIntensity = light.Dot(*v.Normal)
	return
}

func fragmentShader(in *VertexOut) *mgl.Vec4 {
	if in.LightIntensity < 0 {
		return nil
	}

	return &mgl.Vec4{
		in.LightIntensity,
		in.LightIntensity,
		in.LightIntensity,
		1,
	}
}

func barycentric(x, y float64, v0, v1, v2 *mgl.Vec3) (a1, a2, a3 float64) {
	u := mgl.Vec3{v2.X() - v0.X(), v1.X() - v0.X(), v0.X() - x}.Cross(mgl.Vec3{v2.Y() - v0.Y(), v1.Y() - v0.Y(), v0.Y() - y})

	a1 = 1.0 - (u.X()+u.Y())/u.Z()
	a2 = u.Y() / u.Z()
	a3 = u.X() / u.Z()
	return
}

func centric(x, y float64, v0, v1, v2 *VertexOut) (v *VertexOut) {
	v = &VertexOut{Position: &mgl.Vec3{x, y, 0}}

	d0, d1, d2 := barycentric(x, y, v0.Position, v1.Position, v2.Position)

	average := func(f func(*VertexOut) float64) float64 {
		return d0*f(v0) + d1*f(v1) + d2*f(v2)
	}

	v.Position[2] = average(func(v *VertexOut) float64 { return v.Position.Z() })
	v.LightIntensity = average(func(v *VertexOut) float64 { return v.LightIntensity })
	return
}

func onSide(y float64, v1, v2 *mgl.Vec3) float64 {
	return ((v2.X()-v1.X())*y + (v1.X()*v2.Y() - v2.X()*v1.Y())) / (v2.Y() - v1.Y())
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

	var faces []*Face

	for _, face := range head.Faces {
		p1 := face.Points[0]
		p2 := face.Points[1]
		p3 := face.Points[2]

		faces = append(
			faces,
			&Face{
				&Vertex{
					Pos:    &mgl.Vec3{p1.Vertex.X, p1.Vertex.Y, p1.Vertex.Z},
					Normal: &mgl.Vec3{p1.Normal.X, p1.Normal.Y, p1.Normal.Z},
				},
				&Vertex{
					Pos:    &mgl.Vec3{p2.Vertex.X, p2.Vertex.Y, p2.Vertex.Z},
					Normal: &mgl.Vec3{p2.Normal.X, p2.Normal.Y, p2.Normal.Z},
				},
				&Vertex{
					Pos:    &mgl.Vec3{p3.Vertex.X, p3.Vertex.Y, p3.Vertex.Z},
					Normal: &mgl.Vec3{p3.Normal.X, p3.Normal.Y, p3.Normal.Z},
				},
			},
		)
	}

	(&Program{Screen: dest}).Run(faces)

	draw2dimg.SaveToPngFile("watcher/hello.png", dest)
	fmt.Println("Done")
}
