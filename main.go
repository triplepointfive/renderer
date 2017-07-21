package main

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"os"

	"github.com/ftrvxmtrx/tga"
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
		UV     *mgl.Vec2
	}

	// Face -
	Face [3]*Vertex

	// Program -
	Program struct {
		Screen      *image.RGBA
		FaceTexture image.Image
	}

	// VertexOut -
	VertexOut struct {
		Position *mgl.Vec3
		Normal   *mgl.Vec3
		UV       *mgl.Vec2
	}
)

// Run -
func (program *Program) Run(faces []*Face) {
	for _, face := range faces {
		m0 := vertexShader(face[0])
		m1 := vertexShader(face[1])
		m2 := vertexShader(face[2])

		yMax := 0.0
		yMin := float64(height)
		xMax := 0.0
		xMin := float64(width)
		for _, m := range []*VertexOut{m0, m1, m2} {
			yMax = math.Max(yMax, m.Position.Y())
			yMin = math.Min(yMin, m.Position.Y())
			xMax = math.Max(xMax, m.Position.X())
			xMin = math.Min(xMin, m.Position.X())
		}

		for y := math.Floor(yMin); y <= math.Ceil(yMax); y++ {
			for x := math.Floor(xMin); x <= math.Ceil(xMax); x++ {
				m := centric(x, y, m0, m1, m2)
				if m == nil {
					continue
				}
				z := m.Position.Z()

				if zBuffer[int(x)][int(y)] < z {
					if fragColor := program.fragmentShader(m); fragColor != nil {
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
}

func vertexShader(v *Vertex) (out *VertexOut) {
	out = &VertexOut{}
	out.Position = &mgl.Vec3{
		(v.Pos.X() + 1) * width / 2,
		(v.Pos.Y() + 1) * height / 2,
		(v.Pos.Z() + 1) / 2,
	}

	out.Normal = v.Normal
	out.UV = v.UV
	return
}

func (program *Program) fragmentShader(in *VertexOut) *mgl.Vec4 {
	if in.Normal.Z() < 0 {
		return nil
	}
	lightIntensity := light.Dot(*in.Normal)

	if lightIntensity < 0 {
		return nil
	}

	pos := in.UV
	size := program.FaceTexture.Bounds().Size()
	x := pos.X() * float64(size.X)
	y := float64(size.Y) - pos.Y()*float64(size.Y)

	r, g, b, _ := program.FaceTexture.At(int(x), int(y)).RGBA()

	return &mgl.Vec4{
		float64(r) / 0xff,
		float64(g) / 0xff,
		float64(b) / 0xff,
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

func centric(x, y float64, v0, v1, v2 *VertexOut) *VertexOut {
	d0, d1, d2 := barycentric(x, y, v0.Position, v1.Position, v2.Position)
	if d0 < 0 || d1 < 0 || d2 < 0 {
		return nil
	}

	average := func(f func(*VertexOut) float64) float64 {
		return d0*f(v0) + d1*f(v1) + d2*f(v2)
	}

	return &VertexOut{
		Position: &mgl.Vec3{
			x,
			y,
			average(func(v *VertexOut) float64 { return v.Position.Z() }),
		},
		Normal: &mgl.Vec3{
			average(func(v *VertexOut) float64 { return v.Normal.X() }),
			average(func(v *VertexOut) float64 { return v.Normal.Y() }),
			average(func(v *VertexOut) float64 { return v.Normal.Z() }),
		},
		UV: &mgl.Vec2{
			average(func(v *VertexOut) float64 { return v.UV.X() }),
			average(func(v *VertexOut) float64 { return v.UV.Y() }),
		},
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
					UV:     &mgl.Vec2{p1.Texture.U, p1.Texture.V},
				},
				&Vertex{
					Pos:    &mgl.Vec3{p2.Vertex.X, p2.Vertex.Y, p2.Vertex.Z},
					Normal: &mgl.Vec3{p2.Normal.X, p2.Normal.Y, p2.Normal.Z},
					UV:     &mgl.Vec2{p2.Texture.U, p2.Texture.V},
				},
				&Vertex{
					Pos:    &mgl.Vec3{p3.Vertex.X, p3.Vertex.Y, p3.Vertex.Z},
					Normal: &mgl.Vec3{p3.Normal.X, p3.Normal.Y, p3.Normal.Z},
					UV:     &mgl.Vec2{p3.Texture.U, p3.Texture.V},
				},
			},
		)
	}

	program := &Program{
		Screen:      dest,
		FaceTexture: loadTexture(),
	}
	program.Run(faces)

	draw2dimg.SaveToPngFile("watcher/hello.png", dest)
	fmt.Println("Done")
}

func loadTexture() image.Image {
	f, err := os.Open("tinyrenderer/obj/african_head/african_head_diffuse.tga")
	if err != nil {
		panic(err)
	}

	img, err := tga.Decode(f)
	if err != nil {
		panic(err)
	}

	return img
}
