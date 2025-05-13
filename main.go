package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"os"
	"strconv"
	"strings"
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

// 3d model parsed from wavefront object format
// https://en.wikipedia.org/wiki/Wavefront_.obj_file
type wfobj struct {
	Vertices []vertex
	Faces    []face
}

type vertex struct {
	x, y, z float64
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

func renderLines(i *image.RGBA) {
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

	// //line(i, 200, 200, 700, 100, cyan)
	// //line(i, 200, 200, 300, 700, magenta)
	// //line(i, 200, 200, 700, 700, yellow)
	// //line(i, 200, 200, 600, 500, green)
	// //line(i, 200, 200, 300, 0, red)
}

// render frontal 2D projection of model (no z-index)
func render2D(img *image.RGBA, model *wfobj) {
	imgWidth := img.Rect.Max.X
	imgHeight := img.Rect.Max.Y
	for _, face := range model.Faces {
		// map normalized coordinates to be relative to center of image
		v0 := model.Vertices[face.VRefs[0]]
		x0 := int(float64(imgWidth/2) + v0.x*float64(imgWidth/2))
		y0 := int(float64(imgHeight/2) - v0.y*float64(imgHeight/2))

		v1 := model.Vertices[face.VRefs[1]]
		x1 := int(float64(imgWidth/2) + v1.x*float64(imgWidth/2))
		y1 := int(float64(imgHeight/2) - v1.y*float64(imgHeight/2))

		v2 := model.Vertices[face.VRefs[2]]
		x2 := int(float64(imgWidth/2) + v2.x*float64(imgWidth/2))
		y2 := int(float64(imgHeight/2) - v2.y*float64(imgHeight/2))

		line(img, x0, y0, x1, y1, cyan)
		line(img, x0, y0, x2, y2, magenta)
		line(img, x1, y1, x2, y2, yellow)
	}
}

type tri [3]image.Point

// by y coordinate
func (t *tri) Sort() {
	if t[0].Y > t[1].Y {
		t[0], t[1] = t[1], t[0]
	}
	if t[0].Y > t[2].Y {
		fmt.Println("swap 02")
		t[0], t[2] = t[2], t[0]
	}
	if t[1].Y > t[2].Y {
		fmt.Println("swap 12")
		t[1], t[2] = t[2], t[1]
	}
}

func triangle(img *image.RGBA, t tri) {
	t.Sort()

	line(img, t[0].X, t[0].Y, t[1].X, t[1].Y, grey)
	line(img, t[0].X, t[0].Y, t[2].X, t[2].Y, grey)
	line(img, t[1].X, t[1].Y, t[2].X, t[2].Y, grey)

	// dx1 could be the left or right line but it doesn't really matter
	dx1 := t[1].X - t[0].X
	dx2 := t[2].X - t[0].X
	dy1 := t[1].Y - t[0].Y
	dy2 := t[2].Y - t[0].Y

	// top half
	for i := t[0].Y; i <= t[1].Y; i++ {
		sx1 := float64(dx1*(i-t[0].Y)) / float64(dy1)
		sx2 := float64(dx2*(i-t[0].Y)) / float64(dy2)
		x1 := t[0].X + int(sx1)
		x2 := t[0].X + int(sx2)
		swapgt(&x1, &x2)
		for ii := x1; ii <= x2; ii++ {
			img.Set(ii, i, cyan)
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

	renderLines(img)

	//start := time.Now()
	//render2D(img, model)
	//end := time.Now()
	//fmt.Printf("render in %v\n", end.Sub(start))

	t1 := tri{{450, 300}, {250, 200}, {400, 600}}
	triangle(img, t1)

	f, err := os.Create("out.png")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	if err := png.Encode(f, img); err != nil {
		panic(err)
	}
}
