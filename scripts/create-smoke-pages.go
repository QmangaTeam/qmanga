package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
)

func main() {
	if len(os.Args) != 2 {
		panic("usage: go run scripts/create-smoke-pages.go OUTPUT_DIRECTORY")
	}
	if err := os.MkdirAll(os.Args[1], 0700); err != nil {
		panic(err)
	}
	for page := 1; page <= 3; page++ {
		canvas := image.NewRGBA(image.Rect(0, 0, 800, 1200))
		draw.Draw(canvas, canvas.Bounds(), image.NewUniform(color.RGBA{245, 240, 225, 255}), image.Point{}, draw.Src)
		for panel := 0; panel < page+2; panel++ {
			shade := color.RGBA{uint8(35 + page*20), uint8(70 + panel*15), 85, 255}
			draw.Draw(canvas, image.Rect(40, 40+panel*190, 760, 200+panel*190), image.NewUniform(shade), image.Point{}, draw.Src)
		}
		file, err := os.Create(filepath.Join(os.Args[1], fmt.Sprintf("page-%02d.png", page)))
		if err != nil {
			panic(err)
		}
		if err := png.Encode(file, canvas); err != nil {
			file.Close()
			panic(err)
		}
		if err := file.Close(); err != nil {
			panic(err)
		}
	}
}
