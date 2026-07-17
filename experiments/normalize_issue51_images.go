// Normalize the generated issue #51 PNGs to the exact requested dimensions.
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
	"/home/box/.codex/generated_images/019f6d9d-ea6a-7b62-b448-0b817f889c41/exec-c722715b-3e82-4b70-b0d1-998e3b4d0381.png": "blog/asymmetric-garch-crypto-leverage.png",
	"/home/box/.codex/generated_images/019f6d9d-ea6a-7b62-b448-0b817f889c41/exec-1240a165-e255-4fe6-b573-1df0ee275543.png": "blog/asymmetric-garch-crypto-leverage-news-impact.png",
	"/home/box/.codex/generated_images/019f6d9d-ea6a-7b62-b448-0b817f889c41/exec-0d3175a9-cd05-4c13-a218-b2041ded68b4.png": "blog/asymmetric-garch-crypto-leverage-fat-tails.png",
	"/home/box/.codex/generated_images/019f6d9d-ea6a-7b62-b448-0b817f889c41/exec-e69d52d8-a611-437f-b354-5a96e71120d5.png": "blog/asymmetric-garch-crypto-leverage-var-es.png",
	"/home/box/.codex/generated_images/019f6d9d-ea6a-7b62-b448-0b817f889c41/exec-601a4089-59e3-4eae-971d-8ab57afaeaf2.png": "blog/asymmetric-garch-crypto-leverage-model-selection.png",
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
				// Drop only the two least-significant channel bits. This subtle
				// quantization is visually imperceptible and compresses fine glow
				// gradients into the issue's requested file-size range.
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
