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

func main() {
	img := image.NewRGBA(image.Rectangle{
		Max: image.Point{width, height},
	})

	img.Set(400, 400, red)

	f, err := os.Create("out.png")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	if err := png.Encode(f, img); err != nil {
		panic(err)
	}
}
