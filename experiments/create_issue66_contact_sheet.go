// Command create_issue66_contact_sheet creates review evidence for issue 66.
package main

import (
	"image"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
)

var names = []string{
	"twap-vwap-pov-execution-algorithms.png",
	"twap-vwap-pov-execution-algorithms-three-schedulers.png",
	"twap-vwap-pov-execution-algorithms-volume-curve.png",
	"twap-vwap-pov-execution-algorithms-pov-feedback.png",
	"twap-vwap-pov-execution-algorithms-benchmark-race.png",
}

func main() {
	const tileW, tileH = 832, 468
	sheet := image.NewRGBA(image.Rect(0, 0, tileW*2, tileH*3))
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
		x0, y0 := (i%2)*tileW, (i/2)*tileH
		// Every source is exactly twice the tile dimensions, so a 2x2 box
		// average gives a deterministic half-size preview without dependencies.
		for y := 0; y < tileH; y++ {
			for x := 0; x < tileW; x++ {
				draw.Draw(sheet, image.Rect(x0+x, y0+y, x0+x+1, y0+y+1), src, image.Pt(x*2, y*2), draw.Src)
			}
		}
	}
	if err := os.MkdirAll("docs/screenshots", 0o755); err != nil {
		panic(err)
	}
	out, err := os.Create("docs/screenshots/issue66-contact-sheet.png")
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
