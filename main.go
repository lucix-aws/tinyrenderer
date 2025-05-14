package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"math"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

const width = 800
const height = 800

var (
	red   = color.RGBA{0xff, 0, 0, 0xff}
	green = color.RGBA{0, 0xff, 0, 0xff}
	blue  = color.RGBA{0, 0, 0xff, 0xff}

	cyan    = color.RGBA{0, 0xff, 0xff, 0xff}
	magenta = color.RGBA{0xff, 0, 0xff, 0xff}
	yellow  = color.RGBA{0xff, 0xff, 0, 0xff}

	white = color.RGBA{0xff, 0xff, 0xff, 0xff}
	grey  = color.RGBA{0x80, 0x80, 0x80, 0xff}
)

var colors = []color.Color{red, green, blue, cyan, magenta, yellow}

func randcolor() color.Color {
	i := rand.Int() % len(colors)
	return colors[i]
}

// 3d model parsed from wavefront object format
// https://en.wikipedia.org/wiki/Wavefront_.obj_file
type wfobj struct {
	Vertices []vertex
	Faces    []face
}

// 3d point, also functions as a "vector" type
// if you did calc3 these operations will be familiar
type vertex struct {
	x, y, z float64
}

func (v *vertex) Length() float64 {
	return math.Sqrt(v.x*v.x + v.y*v.y + v.z*v.z)
}

func (v *vertex) DotProduct(o *vertex) float64 {
	return v.x*o.x + v.y*o.y + v.z*o.z
}

func (v *vertex) CrossProduct(o *vertex) *vertex {
	return &vertex{
		v.y*o.z - v.z*o.y,
		v.z*o.x - v.x*o.z,
		v.x*o.y - v.y*o.x,
	}
}

func (v *vertex) Sub(o *vertex) *vertex {
	return &vertex{
		v.x - o.x,
		v.y - o.y,
		v.z - o.z,
	}
}

func (v *vertex) Unit() *vertex {
	l := v.Length()
	return &vertex{
		v.x / l,
		v.y / l,
		v.z / l,
	}
}

// should be able to combine w/ vertex via generic
type Point3 struct {
	X, Y, Z int
}

func (i Point3) CrossProduct(j Point3) Point3 {
	return Point3{
		i.Y*j.Z - i.Z*j.Y,
		i.Z*j.X - i.X*j.Z,
		i.X*j.Y - i.Y*j.X,
	}
}

type face struct {
	VRefs []int
}

func parsewf(contents string) (*wfobj, error) {
	var obj wfobj

	lines := strings.Split(contents, "\n")
	for _, line := range lines {
		if line == "" { // blank
			continue
		} else if strings.HasPrefix(line, "#") { // comment
			continue
		} else if strings.HasPrefix(line, "vt") { // something texture
		} else if strings.HasPrefix(line, "vn") { // vertex normal?
		} else if strings.HasPrefix(line, "g") { // ???
		} else if strings.HasPrefix(line, "s") { // ???
		} else if strings.HasPrefix(line, "v") { // vertex
			vdecl := strings.Split(line, " ")
			if len(vdecl) != 4 {
				return nil, fmt.Errorf("invalid v decl: %q", line)
			}

			x, err := strconv.ParseFloat(vdecl[1], 64)
			if err != nil {
				return nil, fmt.Errorf("invalid v.x decl: %q", vdecl[1])
			}

			y, err := strconv.ParseFloat(vdecl[2], 64)
			if err != nil {
				return nil, fmt.Errorf("invalid v.y decl: %q", vdecl[2])
			}

			z, err := strconv.ParseFloat(vdecl[3], 64)
			if err != nil {
				return nil, fmt.Errorf("invalid v.z decl: %q", vdecl[3])
			}

			obj.Vertices = append(obj.Vertices, vertex{x, y, z})
		} else if strings.HasPrefix(line, "f") { // face
			fdecl := strings.Split(line, " ")
			if len(fdecl) != 4 {
				return nil, fmt.Errorf("invalid f decl: %q", line)
			}

			// only look at vertex part for now

			f1 := strings.Split(fdecl[1], "/")
			vref1, err := strconv.Atoi(f1[0])
			if err != nil {
				return nil, fmt.Errorf("invalid f1 decl: %q", f1)
			}

			f2 := strings.Split(fdecl[2], "/")
			vref2, err := strconv.Atoi(f2[0])
			if err != nil {
				return nil, fmt.Errorf("invalid f2 decl: %q", f2)
			}

			f3 := strings.Split(fdecl[3], "/")
			vref3, err := strconv.Atoi(f3[0])
			if err != nil {
				return nil, fmt.Errorf("invalid f3 decl: %q", f3)
			}

			obj.Faces = append(obj.Faces, face{
				VRefs: []int{vref1 - 1, vref2 - 1, vref3 - 1},
			})
		} else {
			return nil, fmt.Errorf("unrecognized line type: %q", line)
		}
	}

	return &obj, nil
}

type Matrix2f struct {
	A, B float64
	C, D float64
}

func (m Matrix2f) MultPoint(p image.Point) image.Point {
	return image.Point{
		int(m.A*float64(p.X) + m.B*float64(p.Y)),
		int(m.C*float64(p.X) + m.D*float64(p.Y)),
	}
}

// does not touch the z-coordinate
func (m Matrix2f) MultVertex(v vertex) vertex {
	return vertex{
		m.A*v.x + m.B*v.y,
		m.C*v.x + m.D*v.y,
		v.z,
	}
}

func line(img *image.RGBA, x0, y0, x1, y1 int, c color.Color) {
	// ensure draw is ltr
	if x1 < x0 {
		x0, x1 = x1, x0
		y0, y1 = y1, y0
	}

	dy := y1 - y0
	dx := x1 - x0
	if dy < 0 && -dy > dx {
		for i := 0; i <= -dy; i++ {
			sx := float64(dx*i) / float64(-dy)
			img.Set(x0+int(sx), y0-i, c)
		}
	} else if dx > dy {
		for i := 0; i <= dx; i++ {
			sy := float64(dy*i) / float64(dx)
			img.Set(x0+i, y0+int(sy), c)
		}
		// alternative: x0-based indexing
		// for i := x0; i <= x1; i++ {
		//     sy := float64(dy*(i-x0)) / float64(dx)
		//     img.Set(i, y0+int(sy), c)
		// }
	} else {
		for i := 0; i <= dy; i++ {
			sx := float64(dx*i) / float64(dy)
			img.Set(x0+int(sx), y0+i, c)
		}
	}
}

func renderGrid(i *image.RGBA) {
	// grid x
	line(i, 0, 100, 800, 100, grey)
	line(i, 0, 200, 800, 200, grey)
	line(i, 0, 300, 800, 300, grey)
	line(i, 0, 400, 800, 400, grey)
	line(i, 0, 500, 800, 500, grey)
	line(i, 0, 600, 800, 600, grey)
	line(i, 0, 700, 800, 700, grey)
	// grid y
	line(i, 100, 0, 100, 800, grey)
	line(i, 200, 0, 200, 800, grey)
	line(i, 300, 0, 300, 800, grey)
	line(i, 400, 0, 400, 800, grey)
	line(i, 500, 0, 500, 800, grey)
	line(i, 600, 0, 600, 800, grey)
	line(i, 700, 0, 700, 800, grey)
}

// render frontal 2D projection of model (no z-index)
func render2D(img *image.RGBA, model *wfobj) {
	lightSource := vertex{0, 0, -1}

	imgWidth := img.Rect.Max.X
	imgHeight := img.Rect.Max.Y
	zbuf := make([]float64, imgWidth*imgHeight)
	for i := range zbuf {
		zbuf[i] = math.Inf(-1)
	}

	for _, face := range model.Faces {
		v0 := model.Vertices[face.VRefs[0]]
		v1 := model.Vertices[face.VRefs[1]]
		v2 := model.Vertices[face.VRefs[2]]

		// test matrix on vertices
		identity2 := Matrix2f{
			1, 0,
			0.3, 1,
		}
		v0 = identity2.MultVertex(v0)
		v1 = identity2.MultVertex(v1)
		v2 = identity2.MultVertex(v2)

		// map vertices to be relative to center of image
		p0 := image.Point{
			int(float64(imgWidth/2) + v0.x*float64(imgWidth/2)),
			int(float64(imgHeight/2) - v0.y*float64(imgHeight/2)),
		}
		p1 := image.Point{
			int(float64(imgWidth/2) + v1.x*float64(imgWidth/2)),
			int(float64(imgHeight/2) - v1.y*float64(imgHeight/2)),
		}
		p2 := image.Point{
			int(float64(imgWidth/2) + v2.x*float64(imgWidth/2)),
			int(float64(imgHeight/2) - v2.y*float64(imgHeight/2)),
		}

		// calculate illumination on face. you can derive a brightness by
		// defining an arbitrary "light source" unit vector that describes
		// where the light is pointing. then if you take the dot product of
		// that and the unit vector _normal_ to two of the triangle's faces,
		// you get a brightness value. this value can be negative, which means
		// the polygon is facing away from you, so we can just skip drawing
		// that.
		face1 := v2.Sub(&v0)
		face2 := v1.Sub(&v0)
		faceNormal := face1.CrossProduct(face2).Unit()
		faceBrightness := lightSource.DotProduct(faceNormal)

		//line(img, x0, y0, x1, y1, cyan)
		//line(img, x0, y0, x2, y2, magenta)
		//line(img, x1, y1, x2, y2, yellow)
		if faceBrightness > 0 {
			rgb := byte(faceBrightness * 0xff)
			clr := color.RGBA{0, rgb, rgb, 0xff}
			triangle(img, tri{p0, p1, p2}, clr, v0.z, v1.z, v2.z, zbuf)
		}
	}
}

type tri [3]image.Point

// by y coordinate
func (t *tri) Sort() {
	if t[0].Y > t[1].Y {
		t[0], t[1] = t[1], t[0]
	}
	if t[0].Y > t[2].Y {
		t[0], t[2] = t[2], t[0]
	}
	if t[1].Y > t[2].Y {
		t[1], t[2] = t[2], t[1]
	}
}

// "primitive" triangleByHalves render: sort by y coordinates, divide into 2 sub-triangles,
// and fill one line at a time
func triangleByHalves(img *image.RGBA, t tri, c color.Color) {
	t.Sort()

	// dx1 could be the left or right line but it doesn't really matter
	dx1 := t[1].X - t[0].X
	dx2 := t[2].X - t[0].X
	dy1 := t[1].Y - t[0].Y
	dy2 := t[2].Y - t[0].Y

	// top half (if not flat top e.g Y1==Y2)
	// we need to remember our final x offsets here so we can pick up
	// from those points at the end
	x1 := t[0].X
	x2 := t[1].X
	if dy1 > 0 {
		for i := t[0].Y; i <= t[1].Y; i++ {
			sx1 := float64(dx1*(i-t[0].Y)) / float64(dy1)
			sx2 := float64(dx2*(i-t[0].Y)) / float64(dy2)
			x1 = t[0].X + int(sx1)
			x2 = t[0].X + int(sx2)
			swapgt(&x1, &x2)
			for ii := x1; ii <= x2; ii++ {
				img.Set(ii, i, c)
			}
		}
	}

	// bottom half (if not flat bottom e.g. Y2==Y3)
	// must recompute deltas
	// there's only ONE dy now because we're converging on the last point
	dy := t[2].Y - t[1].Y
	if dy > 0 {
		dx1 = t[2].X - x1
		dx2 = t[2].X - x2
		origX1 := x1
		origX2 := x2
		for i := t[1].Y; i <= t[2].Y; i++ {
			sx1 := float64(dx1*(i-t[1].Y)) / float64(dy)
			sx2 := float64(dx2*(i-t[1].Y)) / float64(dy)
			x1 := origX1 + int(sx1)
			x2 := origX2 + int(sx2)
			swapgt(&x1, &x2)
			for ii := x1; ii <= x2; ii++ {
				img.Set(ii, i, c)
			}
		}
	}
}

// there are several ways to compute barycentric coordinates given a
// triangle+Point, for this project I am not super concerned with how that is
// derived. What matters is that we understand _what_ barycentric coordinates
// are and how to interpret them to do things like z-buffering or
// shading/texturing a triangle.
func barycentric(t tri, p image.Point) vertex {
	v1 := Point3{t[2].X - t[0].X, t[1].X - t[0].X, t[0].X - p.X}
	v2 := Point3{t[2].Y - t[0].Y, t[1].Y - t[0].Y, t[0].Y - p.Y}
	cross := v1.CrossProduct(v2)
	if abs(cross.Z) < 1 {
		return vertex{-1, -1, -1}
	}
	return vertex{
		1 - float64(cross.X+cross.Y)/float64(cross.Z),
		float64(cross.Y) / float64(cross.Z),
		float64(cross.X) / float64(cross.Z),
	}
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

// draws a triangle using barycentric coordinates
func triangle(img *image.RGBA, t tri, c color.Color, z1, z2, z3 float64, zbuf []float64) {
	boxMin := image.Point{
		min(t[0].X, t[1].X, t[2].X),
		min(t[0].Y, t[1].Y, t[2].Y),
	}
	boxMax := image.Point{
		max(t[0].X, t[1].X, t[2].X),
		max(t[0].Y, t[1].Y, t[2].Y),
	}
	for x := boxMin.X; x < boxMax.X; x++ {
		for y := boxMin.Y; y < boxMax.Y; y++ {
			b := barycentric(t, image.Point{x, y})
			if b.x < 0 || b.y < 0 || b.z < 0 {
				continue
			}

			z := z1*b.x + z2*b.y + z3*b.z
			zi := x + y*img.Rect.Max.Y
			// coordinate transforms can make things go offscreen, should
			// probably handle that earlier
			if zi >= 0 && zi < len(zbuf) && zbuf[zi] < z {
				zbuf[zi] = z
				img.Set(x, y, c)
			}
		}
	}
}

func swapgt(i, j *int) {
	if *i > *j {
		*i, *j = *j, *i
	}
}

func main() {
	img := image.NewRGBA(image.Rectangle{
		Max: image.Point{width, height},
	})

	objfile, err := os.Open("in.obj")
	if err != nil {
		panic(err)
	}
	defer objfile.Close()

	obj, err := io.ReadAll(objfile)
	if err != nil {
		panic(err)
	}

	model, err := parsewf(string(obj))
	if err != nil {
		panic(err)
	}

	fmt.Printf("loaded model w/ %d vertices, %d faces\n", len(model.Vertices), len(model.Faces))

	renderGrid(img)

	start := time.Now()
	render2D(img, model)
	end := time.Now()
	fmt.Printf("render in %v\n", end.Sub(start))

	f, err := os.Create("out.png")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	if err := png.Encode(f, img); err != nil {
		panic(err)
	}
}
