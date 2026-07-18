// Normalize the issue 78 source images to exact 1664x936 RGB PNGs.
package main

import (
	"image"
	"image/color"
	"image/png"
	"log"
	"os"
)

const (
	targetWidth  = 1664
	targetHeight = 936
)

func main() {
	if len(os.Args) != 3 {
		log.Fatalf("usage: %s input.png output.png", os.Args[0])
	}
	in, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	defer in.Close()

	src, _, err := image.Decode(in)
	if err != nil {
		log.Fatal(err)
	}
	dst := image.NewRGBA(image.Rect(0, 0, targetWidth, targetHeight))
	bounds := src.Bounds()
	for y := 0; y < targetHeight; y++ {
		sy := bounds.Min.Y + y*bounds.Dy()/targetHeight
		for x := 0; x < targetWidth; x++ {
			sx := bounds.Min.X + x*bounds.Dx()/targetWidth
			r, g, b, _ := src.At(sx, sy).RGBA()
			// Three-bit channel quantization preserves the neon composition while
			// keeping detailed generated imagery inside the requested size envelope.
			dst.SetRGBA(x, y, color.RGBA{uint8(r>>8) & 0xf8, uint8(g>>8) & 0xf8, uint8(b>>8) & 0xf8, 255})
		}
	}

	out, err := os.Create(os.Args[2])
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()
	encoder := png.Encoder{CompressionLevel: png.BestCompression}
	if err := encoder.Encode(out, dst); err != nil {
		log.Fatal(err)
	}
}
