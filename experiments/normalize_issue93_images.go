// Normalize issue #93 generated masters to exact 1664x936 PNG assets.
package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
)

type asset struct {
	source string
	target string
}

var assets = []asset{
	{"/home/box/.codex/generated_images/019f758d-6f7b-7ea2-98f9-ba36d61fc499/exec-a34ab149-ee4f-44ea-8911-aae42a36404b.png", "blog/mev-supply-chain-pbs-mevboost.png"},
	{"/home/box/.codex/generated_images/019f758d-6f7b-7ea2-98f9-ba36d61fc499/exec-5b709ce5-10c4-46e2-a06a-613f4b4af9b7.png", "blog/mev-supply-chain-pbs-mevboost-pbs-pipeline.png"},
	{"/home/box/.codex/generated_images/019f758d-6f7b-7ea2-98f9-ba36d61fc499/exec-aba44138-056a-4103-9ca3-e23a4f168469.png", "blog/mev-supply-chain-pbs-mevboost-bid-shading.png"},
	{"/home/box/.codex/generated_images/019f758d-6f7b-7ea2-98f9-ba36d61fc499/exec-1969eaf4-057b-4db3-acae-eff38a648598.png", "blog/mev-supply-chain-pbs-mevboost-order-flow.png"},
	{"/home/box/.codex/generated_images/019f758d-6f7b-7ea2-98f9-ba36d61fc499/exec-7ae47a39-f3a8-4380-b55b-48ec0946bd4f.png", "blog/mev-supply-chain-pbs-mevboost-solana-jito.png"},
}

func normalize(a asset) error {
	in, err := os.Open(a.source)
	if err != nil {
		return err
	}
	defer in.Close()
	src, err := png.Decode(in)
	if err != nil {
		return err
	}
	if src.Bounds().Dx() < 1664 || src.Bounds().Dy() < 936 {
		return fmt.Errorf("%s: source too small: %v", a.source, src.Bounds().Size())
	}
	dst := image.NewNRGBA(image.Rect(0, 0, 1664, 936))
	// Bilinear scaling is intentionally implemented here with the standard library
	// so this repository's image-normalization experiment stays dependency-free.
	sb := src.Bounds()
	for y := 0; y < 936; y++ {
		sy := (float64(y)+0.5)*float64(sb.Dy())/936.0 - 0.5
		y0 := int(sy)
		fy := sy - float64(y0)
		if y0 < 0 {
			y0, fy = 0, 0
		}
		if y0 >= sb.Dy()-1 {
			y0, fy = sb.Dy()-1, 0
		}
		y1 := y0 + 1
		if y1 >= sb.Dy() {
			y1 = y0
		}
		for x := 0; x < 1664; x++ {
			sx := (float64(x)+0.5)*float64(sb.Dx())/1664.0 - 0.5
			x0 := int(sx)
			fx := sx - float64(x0)
			if x0 < 0 {
				x0, fx = 0, 0
			}
			if x0 >= sb.Dx()-1 {
				x0, fx = sb.Dx()-1, 0
			}
			x1 := x0 + 1
			if x1 >= sb.Dx() {
				x1 = x0
			}
			var rgba [4]uint32
			c00 := src.At(sb.Min.X+x0, sb.Min.Y+y0)
			c10 := src.At(sb.Min.X+x1, sb.Min.Y+y0)
			c01 := src.At(sb.Min.X+x0, sb.Min.Y+y1)
			c11 := src.At(sb.Min.X+x1, sb.Min.Y+y1)
			p00 := [4]uint32{}
			p00[0], p00[1], p00[2], p00[3] = c00.RGBA()
			p10 := [4]uint32{}
			p10[0], p10[1], p10[2], p10[3] = c10.RGBA()
			p01 := [4]uint32{}
			p01[0], p01[1], p01[2], p01[3] = c01.RGBA()
			p11 := [4]uint32{}
			p11[0], p11[1], p11[2], p11[3] = c11.RGBA()
			for i := range rgba {
				top := float64(p00[i])*(1-fx) + float64(p10[i])*fx
				bottom := float64(p01[i])*(1-fx) + float64(p11[i])*fx
				rgba[i] = uint32(top*(1-fy) + bottom*fy + 0.5)
			}
			dst.SetNRGBA(x, y, color.NRGBA{uint8(rgba[0] >> 8), uint8(rgba[1] >> 8), uint8(rgba[2] >> 8), uint8(rgba[3] >> 8)})
		}
	}
	out, err := os.Create(a.target)
	if err != nil {
		return err
	}
	encoder := png.Encoder{CompressionLevel: png.BestCompression}
	if err := encoder.Encode(out, dst); err != nil {
		out.Close()
		return err
	}
	if err := out.Close(); err != nil {
		return err
	}
	fmt.Printf("wrote %s\n", a.target)
	return nil
}

func main() {
	for _, a := range assets {
		if err := normalize(a); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
}
