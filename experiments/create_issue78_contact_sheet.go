// Create a compact visual-review sheet for the five issue 78 images.
package main

import (
	"image"
	"image/draw"
	"image/png"
	"log"
	"os"
	"path/filepath"
)

var names = []string{
	"implementation-shortfall-tca-execution.png",
	"implementation-shortfall-tca-execution-perold-gap.png",
	"implementation-shortfall-tca-execution-cost-decomposition.png",
	"implementation-shortfall-tca-execution-markout-curves.png",
	"implementation-shortfall-tca-execution-close-the-loop.png",
}

func main() {
	const tileW, tileH = 832, 468
	sheet := image.NewRGBA(image.Rect(0, 0, tileW*2, tileH*3))
	for i, name := range names {
		f, err := os.Open(filepath.Join("blog", name))
		if err != nil {
			log.Fatal(err)
		}
		src, _, err := image.Decode(f)
		f.Close()
		if err != nil {
			log.Fatal(err)
		}
		dx, dy := (i%2)*tileW, (i/2)*tileH
		for y := 0; y < tileH; y++ {
			for x := 0; x < tileW; x++ {
				sheet.Set(dx+x, dy+y, src.At(src.Bounds().Min.X+x*src.Bounds().Dx()/tileW, src.Bounds().Min.Y+y*src.Bounds().Dy()/tileH))
			}
		}
	}
	// Duplicate the hero into the unused final tile so every row is visually balanced.
	draw.Draw(sheet, image.Rect(tileW, tileH*2, tileW*2, tileH*3), sheet, image.Point{}, draw.Src)
	if err := os.MkdirAll("docs/screenshots", 0o755); err != nil {
		log.Fatal(err)
	}
	out, err := os.Create("docs/screenshots/issue78-contact-sheet.png")
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()
	if err := png.Encode(out, sheet); err != nil {
		log.Fatal(err)
	}
}
