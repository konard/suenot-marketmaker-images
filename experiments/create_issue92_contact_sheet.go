// Build a compact visual-review sheet for issue #92.
package main

import (
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
)

const (
	thumbWidth  = 832
	thumbHeight = 468
	gap         = 16
)

var files = []string{
	"onchain-arbitrage-atomic-flash-loans.png",
	"onchain-arbitrage-atomic-flash-loans-atomic-loop.png",
	"onchain-arbitrage-atomic-flash-loans-cycle-detection.png",
	"onchain-arbitrage-atomic-flash-loans-backrun.png",
	"onchain-arbitrage-atomic-flash-loans-priority-auction.png",
}

func resizeHalf(src image.Image) image.Image {
	dst := image.NewRGBA(image.Rect(0, 0, thumbWidth, thumbHeight))
	bounds := src.Bounds()
	for y := 0; y < thumbHeight; y++ {
		for x := 0; x < thumbWidth; x++ {
			dst.Set(x, y, src.At(bounds.Min.X+x*2, bounds.Min.Y+y*2))
		}
	}
	return dst
}

func main() {
	root := filepath.Join("docs", "screenshots")
	if err := os.MkdirAll(root, 0o755); err != nil {
		panic(err)
	}
	canvas := image.NewRGBA(image.Rect(0, 0, thumbWidth*2+gap*3, thumbHeight*3+gap*4))
	draw.Draw(canvas, canvas.Bounds(), image.Black, image.Point{}, draw.Src)

	for index, name := range files {
		input, err := os.Open(filepath.Join("blog", name))
		if err != nil {
			panic(err)
		}
		img, err := png.Decode(input)
		input.Close()
		if err != nil {
			panic(err)
		}

		column := index % 2
		row := index / 2
		x := gap + column*(thumbWidth+gap)
		y := gap + row*(thumbHeight+gap)
		draw.Draw(canvas, image.Rect(x, y, x+thumbWidth, y+thumbHeight), resizeHalf(img), image.Point{}, draw.Src)
	}

	outputPath := filepath.Join(root, "issue92-contact-sheet.png")
	output, err := os.Create(outputPath)
	if err != nil {
		panic(err)
	}
	defer output.Close()
	if err := (&png.Encoder{CompressionLevel: png.BestCompression}).Encode(output, canvas); err != nil {
		panic(err)
	}
	fmt.Println(outputPath)
}
