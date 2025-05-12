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
)

// instead of an arbitrary step, use dx to figure out how many pixels to draw
// creep along the y axis fractionally as we do so
//
// this does NOT work for lines where dy > dx because dx does not correspond
// to number of iterations in that case
func line(img *image.RGBA, x0, y0, x1, y1 int, c color.Color) {
	dy := y1 - y0
	dx := x1 - x0
	for i := 0; i <= dx; i++ {
		sy := float64(dy) * float64(i) / float64(dx)
		img.Set(x0+i, y0+int(sy), c)
	}
}

// stopgap before vertical lines actually work
func vline(img *image.RGBA, x int, c color.Color) {
	for i := 0; i < img.Rect.Max.Y; i++ {
		img.Set(x, i, c)
	}
}

func render(i *image.RGBA) {
	line(i, 0, 100, 800, 100, white)
	line(i, 0, 200, 800, 200, white)
	line(i, 0, 300, 800, 300, white)
	line(i, 0, 400, 800, 400, white)
	line(i, 0, 500, 800, 500, white)
	line(i, 0, 600, 800, 600, white)
	line(i, 0, 700, 800, 700, white)
	vline(i, 100, white)
	vline(i, 200, white)
	vline(i, 300, white)
	vline(i, 400, white)
	vline(i, 500, white)
	vline(i, 600, white)
	vline(i, 700, white)

	line(i, 200, 200, 700, 100, red)
	// not correct (dotted line) bc dy > dx
	line(i, 200, 200, 300, 700, cyan)
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
