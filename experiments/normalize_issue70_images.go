// Normalize the generated issue #70 PNGs to the exact requested dimensions.
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
	"/home/box/.codex/generated_images/019f729a-3c55-7ae0-9832-9ce0e1e79d08/exec-77718875-9e9a-412f-aa1a-b5aa9887fe55.png": "blog/onchain-liquidations-aave-compound.png",
	"/home/box/.codex/generated_images/019f729a-3c55-7ae0-9832-9ce0e1e79d08/exec-f602794b-e070-4de7-b186-236b0687bbee.png": "blog/onchain-liquidations-aave-compound-health-factor.png",
	"/home/box/.codex/generated_images/019f729a-3c55-7ae0-9832-9ce0e1e79d08/exec-d5ee08d5-3e75-4906-b432-2c4f4869ab9a.png": "blog/onchain-liquidations-aave-compound-oracle-trigger.png",
	"/home/box/.codex/generated_images/019f729a-3c55-7ae0-9832-9ce0e1e79d08/exec-a32e5f88-8b14-40c9-8f41-4b15b5d49f2c.png": "blog/onchain-liquidations-aave-compound-liquidation-bot.png",
	"/home/box/.codex/generated_images/019f729a-3c55-7ae0-9832-9ce0e1e79d08/exec-024aed98-82e5-40c6-a161-b1550b84331b.png": "blog/onchain-liquidations-aave-compound-bonus-competition.png",
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
				// the detailed glow imagery near the requested file-size envelope.
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
