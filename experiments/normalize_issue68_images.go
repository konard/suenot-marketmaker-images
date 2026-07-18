// Normalize the generated issue #68 PNGs to the exact requested dimensions.
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
	{"/home/box/.codex/generated_images/019f7294-2421-76b2-8bf9-38e7942d0637/exec-1e4df18b-f85c-4600-aac4-6aebce98fe14.png", "blog/mev-sandwich-frontrunning-mempool.png"},
	{"/home/box/.codex/generated_images/019f7294-2421-76b2-8bf9-38e7942d0637/exec-459a0c2d-fdba-4862-8490-b67637653ac1.png", "blog/mev-sandwich-frontrunning-mempool-dark-forest.png"},
	{"/home/box/.codex/generated_images/019f7294-2421-76b2-8bf9-38e7942d0637/exec-a508553a-b422-4a76-a68a-b32550d5f7a3.png", "blog/mev-sandwich-frontrunning-mempool-sandwich-math.png"},
	{"/home/box/.codex/generated_images/019f7294-2421-76b2-8bf9-38e7942d0637/exec-bc8b7c27-f779-4100-aee3-ccb967c2d04f.png", "blog/mev-sandwich-frontrunning-mempool-taxonomy.png"},
	{"/home/box/.codex/generated_images/019f7294-2421-76b2-8bf9-38e7942d0637/exec-ee481055-f764-4b46-81d5-52b085533eaf.png", "blog/mev-sandwich-frontrunning-mempool-defenses.png"},
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
