// Create a visual-review contact sheet for the issue #93 image set.
package main

import (
	"image"
	"image/draw"
	"image/png"
	"os"
)

var names = []string{
	"mev-supply-chain-pbs-mevboost.png",
	"mev-supply-chain-pbs-mevboost-pbs-pipeline.png",
	"mev-supply-chain-pbs-mevboost-bid-shading.png",
	"mev-supply-chain-pbs-mevboost-order-flow.png",
	"mev-supply-chain-pbs-mevboost-solana-jito.png",
}

func main() {
	const tileW, tileH = 832, 468
	sheet := image.NewNRGBA(image.Rect(0, 0, tileW*2, tileH*3))
	for i, name := range names {
		f, err := os.Open("blog/" + name)
		if err != nil {
			panic(err)
		}
		src, err := png.Decode(f)
		f.Close()
		if err != nil {
			panic(err)
		}
		tile := image.NewNRGBA(image.Rect(0, 0, tileW, tileH))
		for y := 0; y < tileH; y++ {
			for x := 0; x < tileW; x++ {
				tile.Set(x, y, src.At(src.Bounds().Min.X+x*2, src.Bounds().Min.Y+y*2))
			}
		}
		draw.Draw(sheet, image.Rect((i%2)*tileW, (i/2)*tileH, (i%2+1)*tileW, (i/2+1)*tileH), tile, image.Point{}, draw.Src)
	}
	out, err := os.Create("docs/screenshots/issue93-contact-sheet.png")
	if err != nil {
		panic(err)
	}
	defer out.Close()
	if err := png.Encode(out, sheet); err != nil {
		panic(err)
	}
}
