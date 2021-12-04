package main

import (
	"image"
	"image/color"
	"image/draw"
	"os"
	"time"

	"github.com/ptrcnull/imagedrm"
)

func main() {
	img, err := imagedrm.NewImage()
	if err != nil {
		panic(err)
	}
	defer img.Close()

	sourceFile, err := os.Open("glenda.jpg")
	if err != nil {
		panic(err)
	}
	defer sourceFile.Close()

	source, _, err := image.Decode(sourceFile)
	if err != nil {
		panic(err)
	}

	draw.Draw(img, source.Bounds(), source, image.Point{}, draw.Src)

	for {
		img.Set(100, 100, color.RGBA{R: 255, G: 255, B: 255})
		time.Sleep(time.Second * 1)
		img.Set(100, 100, color.RGBA{})
		time.Sleep(time.Second * 1)
	}
}
