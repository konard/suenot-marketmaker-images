// Command create_issue91_contact_sheet creates visual review evidence for issue 91.
package main

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
)

var names = []string{
	"impermanent-loss-lvr-lp-profitability.png",
	"impermanent-loss-lvr-lp-profitability-il-curve.png",
	"impermanent-loss-lvr-lp-profitability-concentrated-leverage.png",
	"impermanent-loss-lvr-lp-profitability-lvr-decomposition.png",
	"impermanent-loss-lvr-lp-profitability-markout.png",
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
		for y := range tileH {
			for x := range tileW {
				sheet.Set(x0+x, y0+y, src.At(x*2+1, y*2+1))
			}
		}
	}
	if err := os.MkdirAll("docs/screenshots", 0o755); err != nil {
		panic(err)
	}
	out, err := os.Create("docs/screenshots/issue91-contact-sheet.png")
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
