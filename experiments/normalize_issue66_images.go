// Command normalize_issue66_images converts generated source images into the
// exact 1664x936 PNG deliverables requested by issue 66.
package main

import (
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"os"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr, "usage: %s input.png output.png\n", os.Args[0])
		os.Exit(2)
	}
	in, err := os.Open(os.Args[1])
	if err != nil {
		panic(err)
	}
	src, _, err := image.Decode(in)
	in.Close()
	if err != nil {
		panic(err)
	}

	dst := image.NewRGBA(image.Rect(0, 0, 1664, 936))
	draw.Draw(dst, dst.Bounds(), src, src.Bounds().Min, draw.Src)
	out, err := os.Create(os.Args[2])
	if err != nil {
		panic(err)
	}
	encoder := png.Encoder{CompressionLevel: png.BestCompression}
	if err := encoder.Encode(out, dst); err != nil {
		out.Close()
		panic(err)
	}
	if err := out.Close(); err != nil {
		panic(err)
	}
}
