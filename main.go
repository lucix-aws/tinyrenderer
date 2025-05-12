package main

import (
	"image"
	"image/color"
	"image/png"
	"os"
)

const width = 800
const height = 800

var red = color.RGBA{
	0xff, 0, 0, 0xff,
}

var cyan = color.RGBA{
	0, 0xff, 0xff, 0xff,
}

// figure out the slope, and using an arbitrary step, traverse from
// p0 to p1 and fill in the color
func line(img *image.RGBA, x0, y0, x1, y1 int, c color.Color) {
	dy := float64(y1 - y0)
	dx := float64(x1 - x0)
	for i := float64(0); i <= 1; i += 0.1 {
		img.Set(x0+int(dx*i), y0+int(dy*i), c)
	}
}

func render(i *image.RGBA) {
	line(i, 0, 0, 400, 400, cyan)
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
