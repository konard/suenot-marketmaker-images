package main

import (
	"image"
	"image/color"
	"image/png"
	"log"
	"os"
	"path/filepath"
)

const (
	targetWidth  = 1664
	targetHeight = 936
)

func main() {
	if len(os.Args) != 3 {
		log.Fatalf("usage: %s INPUT.png OUTPUT.png", os.Args[0])
	}

	in, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	src, err := png.Decode(in)
	in.Close()
	if err != nil {
		log.Fatal(err)
	}

	dst := image.NewRGBA(image.Rect(0, 0, targetWidth, targetHeight))
	srcBounds := centeredCrop(src.Bounds())
	for y := 0; y < targetHeight; y++ {
		for x := 0; x < targetWidth; x++ {
			sx := srcBounds.Min.X + x*srcBounds.Dx()/targetWidth
			sy := srcBounds.Min.Y + y*srcBounds.Dy()/targetHeight
			dst.SetRGBA(x, y, color.RGBAModel.Convert(src.At(sx, sy)).(color.RGBA))
		}
	}

	if err := os.MkdirAll(filepath.Dir(os.Args[2]), 0o755); err != nil {
		log.Fatal(err)
	}
	out, err := os.Create(os.Args[2])
	if err != nil {
		log.Fatal(err)
	}
	encoder := png.Encoder{CompressionLevel: png.BestCompression}
	if err := encoder.Encode(out, dst); err != nil {
		out.Close()
		log.Fatal(err)
	}
	if err := out.Close(); err != nil {
		log.Fatal(err)
	}
}

func centeredCrop(bounds image.Rectangle) image.Rectangle {
	w, h := bounds.Dx(), bounds.Dy()
	if w*targetHeight > h*targetWidth {
		cropWidth := h * targetWidth / targetHeight
		x := bounds.Min.X + (w-cropWidth)/2
		return image.Rect(x, bounds.Min.Y, x+cropWidth, bounds.Max.Y)
	}
	cropHeight := w * targetHeight / targetWidth
	y := bounds.Min.Y + (h-cropHeight)/2
	return image.Rect(bounds.Min.X, y, bounds.Max.X, y+cropHeight)
}
