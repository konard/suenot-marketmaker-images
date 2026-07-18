// Build a compact visual-review sheet for issue #70 images.
package main

import (
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"os"
)

const (
	thumbWidth  = 832
	thumbHeight = 468
)

var images = []string{
	"blog/onchain-liquidations-aave-compound.png",
	"blog/onchain-liquidations-aave-compound-health-factor.png",
	"blog/onchain-liquidations-aave-compound-oracle-trigger.png",
	"blog/onchain-liquidations-aave-compound-liquidation-bot.png",
	"blog/onchain-liquidations-aave-compound-bonus-competition.png",
}

func main() {
	sheet := image.NewRGBA(image.Rect(0, 0, thumbWidth*2, thumbHeight*3))
	for i, path := range images {
		input, err := os.Open(path)
		if err != nil {
			panic(err)
		}
		src, err := png.Decode(input)
		input.Close()
		if err != nil {
			panic(err)
		}
		x0 := (i % 2) * thumbWidth
		y0 := (i / 2) * thumbHeight
		for y := 0; y < thumbHeight; y++ {
			for x := 0; x < thumbWidth; x++ {
				sheet.Set(x0+x, y0+y, src.At(src.Bounds().Min.X+x*2, src.Bounds().Min.Y+y*2))
			}
		}
	}
	draw.Draw(sheet, image.Rect(thumbWidth, thumbHeight*2, thumbWidth*2, thumbHeight*3), image.Black, image.Point{}, draw.Src)
	output, err := os.Create("docs/screenshots/issue70-contact-sheet.png")
	if err != nil {
		panic(err)
	}
	defer output.Close()
	if err := png.Encode(output, sheet); err != nil {
		panic(err)
	}
	fmt.Println("docs/screenshots/issue70-contact-sheet.png")
}
