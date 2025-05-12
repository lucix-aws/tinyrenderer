package main

import (
	"image"
	"image/color"
	"image/png"
	"os"
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

// instead of an arbitrary step, use dx to figure out how many pixels to draw
// creep along the y axis fractionally as we do so
//
// this does NOT work for lines where dy > dx because dx does not correspond
// to number of iterations in that case
func line(img *image.RGBA, x0, y0, x1, y1 int, c color.Color) {
	dy := y1 - y0
	dx := x1 - x0
	if dx > dy {
		for i := 0; i <= dx; i++ {
			sy := float64(dy*i) / float64(dx)
			img.Set(x0+i, y0+int(sy), c)
		}
	} else {
		for i := 0; i <= dy; i++ {
			sx := float64(dx*i) / float64(dy)
			img.Set(x0+int(sx), y0+i, c)
		}
	}
}

func render(i *image.RGBA) {
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

	line(i, 200, 200, 700, 100, cyan)
	line(i, 200, 200, 300, 700, magenta)
	line(i, 200, 200, 700, 700, yellow)
}

func main() {
	img := image.NewRGBA(image.Rectangle{
		Max: image.Point{width, height},
	})

	render(img)

	f, err := os.Create("out.png")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	if err := png.Encode(f, img); err != nil {
		panic(err)
	}
}
