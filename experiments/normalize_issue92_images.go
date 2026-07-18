// Normalize the generated issue #92 PNGs to the exact requested dimensions.
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

type imagePath struct {
	source      string
	destination string
}

var images = []imagePath{
	{"/home/box/.codex/generated_images/019f757e-2cf0-7af2-b69e-383ba915d401/exec-26a5ac0e-189d-4e07-8cc4-b195879a1003.png", "blog/onchain-arbitrage-atomic-flash-loans.png"},
	{"/home/box/.codex/generated_images/019f757e-2cf0-7af2-b69e-383ba915d401/exec-7d0cf5b9-2772-40ce-9343-48ce99955da7.png", "blog/onchain-arbitrage-atomic-flash-loans-atomic-loop.png"},
	{"/home/box/.codex/generated_images/019f757e-2cf0-7af2-b69e-383ba915d401/exec-5d64a0a1-c92f-4ad1-8240-1d3ad3698007.png", "blog/onchain-arbitrage-atomic-flash-loans-cycle-detection.png"},
	{"/home/box/.codex/generated_images/019f757e-2cf0-7af2-b69e-383ba915d401/exec-65e84e46-4832-4302-815a-b489bf1ead14.png", "blog/onchain-arbitrage-atomic-flash-loans-backrun.png"},
	{"/home/box/.codex/generated_images/019f757e-2cf0-7af2-b69e-383ba915d401/exec-fd3ddf65-db53-4ed4-b3c3-f0a51bd4f47b.png", "blog/onchain-arbitrage-atomic-flash-loans-priority-auction.png"},
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
				// detailed glow imagery in the requested file-size envelope.
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
	for _, image := range images {
		if err := normalize(image.source, image.destination); err != nil {
			panic(fmt.Sprintf("%s: %v", image.destination, err))
		}
		fmt.Println(image.destination)
	}
}
