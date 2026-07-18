// Command create_issue82_contact_sheet creates visual review evidence for issue 82.
package main

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
)

var names = []string{
	"child-order-execution-tactics.png",
	"child-order-execution-tactics-two-layers.png",
	"child-order-execution-tactics-escalation-ladder.png",
	"child-order-execution-tactics-amend-vs-replace.png",
	"child-order-execution-tactics-iceberg.png",
}

func main() {
	const tileW, tileH, gap = 832, 468, 12
	sheet := image.NewRGBA(image.Rect(0, 0, tileW*2+gap, tileH*3+gap*2))
	for y := range sheet.Bounds().Dy() {
		for x := range sheet.Bounds().Dx() {
			sheet.Set(x, y, color.RGBA{2, 7, 18, 255})
		}
	}
	for i, name := range names {
		f, err := os.Open(filepath.Join("blog", name))
		if err != nil {
			panic(err)
		}
		src, _, err := image.Decode(f)
		f.Close()
		if err != nil {
			panic(err)
		}
		x0, y0 := (i%2)*(tileW+gap), (i/2)*(tileH+gap)
		// Every source is exactly twice the tile dimensions. Sample the center
		// pixel of each 2x2 source block for a dependency-free half-size preview.
		for y := range tileH {
			for x := range tileW {
				sheet.Set(x0+x, y0+y, src.At(x*2+1, y*2+1))
			}
		}
	}
	if err := os.MkdirAll("docs/screenshots", 0o755); err != nil {
		panic(err)
	}
	out, err := os.Create("docs/screenshots/issue82-contact-sheet.png")
	if err != nil {
		panic(err)
	}
	encoder := png.Encoder{CompressionLevel: png.BestCompression}
	if err := encoder.Encode(out, sheet); err != nil {
		out.Close()
		panic(err)
	}
	if err := out.Close(); err != nil {
		panic(err)
	}
}
