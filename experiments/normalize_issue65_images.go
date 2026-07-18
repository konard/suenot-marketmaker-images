package main

import (
	"image"
	"image/draw"
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
		log.Fatalf("usage: %s INPUT OUTPUT", os.Args[0])
	}

	in, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	src, err := png.Decode(in)
	if closeErr := in.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		log.Fatal(err)
	}

	b := src.Bounds()
	if b.Dx() < targetWidth || b.Dy() < targetHeight {
		log.Fatalf("source is smaller than target: %dx%d", b.Dx(), b.Dy())
	}

	// Image generation can return a few extra edge pixels. Center-crop to the
	// requested aspect ratio before resampling so the composition is preserved.
	crop := b
	if b.Dx()*targetHeight > b.Dy()*targetWidth {
		cropWidth := b.Dy() * targetWidth / targetHeight
		crop.Min.X += (b.Dx() - cropWidth) / 2
		crop.Max.X = crop.Min.X + cropWidth
	} else if b.Dx()*targetHeight < b.Dy()*targetWidth {
		cropHeight := b.Dx() * targetHeight / targetWidth
		crop.Min.Y += (b.Dy() - cropHeight) / 2
		crop.Max.Y = crop.Min.Y + cropHeight
	}

	dst := image.NewRGBA(image.Rect(0, 0, targetWidth, targetHeight))
	for y := 0; y < targetHeight; y++ {
		sy := crop.Min.Y + y*crop.Dy()/targetHeight
		for x := 0; x < targetWidth; x++ {
			sx := crop.Min.X + x*crop.Dx()/targetWidth
			dst.Set(x, y, src.At(sx, sy))
		}
	}
	draw.Draw(dst, dst.Bounds(), dst, image.Point{}, draw.Src)

	out, err := os.Create(os.Args[2])
	if err != nil {
		log.Fatal(err)
	}
	encoder := png.Encoder{CompressionLevel: png.BestCompression}
	err = encoder.Encode(out, dst)
	if closeErr := out.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		log.Fatal(err)
	}
}
