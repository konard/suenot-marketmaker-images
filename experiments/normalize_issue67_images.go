// Normalize the generated issue #67 PNGs to the exact requested dimensions.
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
	"/home/box/.codex/generated_images/019f728a-bc10-7663-8ff7-f1e2981cd06b/exec-69cc8080-1be4-4b3e-a3a8-2b473b79337b.png": "blog/fill-simulation-partial-fills-backtest.png",
	"/home/box/.codex/generated_images/019f728a-bc10-7663-8ff7-f1e2981cd06b/exec-ad0d8bdd-28b7-4584-9bf6-b8f2522a696a.png": "blog/fill-simulation-partial-fills-backtest-fidelity-ladder.png",
	"/home/box/.codex/generated_images/019f728a-bc10-7663-8ff7-f1e2981cd06b/exec-be9f8cc5-63c1-4693-99dc-58a74766447c.png": "blog/fill-simulation-partial-fills-backtest-partial-state-machine.png",
	"/home/box/.codex/generated_images/019f728a-bc10-7663-8ff7-f1e2981cd06b/exec-3b9f4df1-ab22-4f8d-891f-5f3caabf34b9.png": "blog/fill-simulation-partial-fills-backtest-probability-bracket.png",
	"/home/box/.codex/generated_images/019f728a-bc10-7663-8ff7-f1e2981cd06b/exec-0aca8ecc-3b81-4f55-b163-47fb8af9f6a6.png": "blog/fill-simulation-partial-fills-backtest-calibration-loop.png",
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
