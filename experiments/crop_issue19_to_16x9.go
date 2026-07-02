package main

import (
	"image"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
)

var files = []string{
	"ipc-tax-backtest-engine.png",
	"ipc-tax-backtest-engine-the-boundary.png",
	"ipc-tax-backtest-engine-serialization-tax.png",
	"ipc-tax-backtest-engine-chatty-vs-chunky.png",
	"ipc-tax-backtest-engine-spawn-cost.png",
	"ipc-tax-backtest-engine-break-even.png",
	"ipc-tax-backtest-engine-transport-floor.png",
	"ipc-tax-backtest-engine-the-verdict.png",
}

func main() {
	for _, name := range files {
		path := filepath.Join("blog", name)
		input, err := os.Open(path)
		if err != nil {
			panic(err)
		}
		src, err := png.Decode(input)
		input.Close()
		if err != nil {
			panic(err)
		}

		bounds := src.Bounds()
		width := bounds.Dx()
		height := bounds.Dy()
		targetWidth := (height * 16 / 9) / 16 * 16
		targetHeight := targetWidth * 9 / 16
		altHeight := (width * 9 / 16) / 9 * 9
		altWidth := altHeight * 16 / 9
		if altWidth <= width && altHeight <= height && altWidth*altHeight > targetWidth*targetHeight {
			targetWidth = altWidth
			targetHeight = altHeight
		}
		if width*9 != height*16 {
			xOffset := (width - targetWidth) / 2
			yOffset := (height - targetHeight) / 2
			cropped := image.NewRGBA(image.Rect(0, 0, targetWidth, targetHeight))
			draw.Draw(cropped, cropped.Bounds(), src, image.Point{X: bounds.Min.X + xOffset, Y: bounds.Min.Y + yOffset}, draw.Src)

			output, err := os.Create(path)
			if err != nil {
				panic(err)
			}
			if err := png.Encode(output, cropped); err != nil {
				output.Close()
				panic(err)
			}
			output.Close()
		}
	}
}
