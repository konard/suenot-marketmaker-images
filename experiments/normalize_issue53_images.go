// Normalize the generated issue #53 PNGs to the exact requested dimensions.
package main

import (
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"os"
)

const (
	targetWidth  = 1664
	targetHeight = 936
)

var images = map[string]string{
	"/home/box/.codex/generated_images/019f6dae-10e2-7c13-a8cc-95ca0cba87ba/exec-2079e5a0-87bf-42d3-8936-d1481b0b0847.png": "blog/volatility-targeting-garch-strategy.png",
	"/home/box/.codex/generated_images/019f6dae-10e2-7c13-a8cc-95ca0cba87ba/exec-2268d50f-d743-4c14-908d-a02a5428834f.png": "blog/volatility-targeting-garch-strategy-vol-targeting.png",
	"/home/box/.codex/generated_images/019f6dae-10e2-7c13-a8cc-95ca0cba87ba/exec-67569ae8-fe9e-4fff-be6e-92079d93f13e.png": "blog/volatility-targeting-garch-strategy-forecast-contest.png",
	"/home/box/.codex/generated_images/019f6dae-10e2-7c13-a8cc-95ca0cba87ba/exec-062b17c7-a9db-40cf-9127-19116beb6b34.png": "blog/volatility-targeting-garch-strategy-forecast-eval.png",
	"/home/box/.codex/generated_images/019f6dae-10e2-7c13-a8cc-95ca0cba87ba/exec-626135a7-70b4-4b13-8a6b-979f2f597dde.png": "blog/volatility-targeting-garch-strategy-walk-forward.png",
}

func normalize(source, destination string) error {
	input, err := os.Open(source)
	if err != nil {
		return err
	}
	defer input.Close()

	src, err := png.Decode(input)
	if err != nil {
		return err
	}
	bounds := src.Bounds()
	if bounds.Dx() < targetWidth || bounds.Dy() < targetHeight {
		return fmt.Errorf("source too small: %dx%d", bounds.Dx(), bounds.Dy())
	}

	left := bounds.Min.X + (bounds.Dx()-targetWidth)/2
	top := bounds.Min.Y + (bounds.Dy()-targetHeight)/2
	rgba := image.NewRGBA(image.Rect(0, 0, targetWidth, targetHeight))
	draw.Draw(rgba, rgba.Bounds(), src, image.Pt(left, top), draw.Src)
	dst := image.NewRGBA(rgba.Bounds())
	for y := 0; y < targetHeight; y++ {
		for x := 0; x < targetWidth; x++ {
			offset := rgba.PixOffset(x, y)
			for channel := 0; channel < 3; channel++ {
				// Two-bit channel quantization is visually imperceptible and keeps
				// the detailed glow imagery in the requested file-size envelope.
				dst.Pix[offset+channel] = rgba.Pix[offset+channel] & 0xfc
			}
			dst.Pix[offset+3] = 0xff
		}
	}

	output, err := os.Create(destination)
	if err != nil {
		return err
	}
	defer output.Close()

	encoder := png.Encoder{CompressionLevel: png.BestCompression}
	return encoder.Encode(output, dst)
}

func main() {
	for source, destination := range images {
		if err := normalize(source, destination); err != nil {
			panic(fmt.Sprintf("%s: %v", destination, err))
		}
		fmt.Println(destination)
	}
}
