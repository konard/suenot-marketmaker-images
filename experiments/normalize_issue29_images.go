package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
)

func main() {
	paths := []string{
		"blog/when-gpu-pays-off-sweep-roofline.png",
		"blog/when-gpu-pays-off-sweep-roofline-model.png",
		"blog/when-gpu-pays-off-sweep-roofline-single-tf-verdict.png",
		"blog/when-gpu-pays-off-sweep-roofline-batch-scaling.png",
		"blog/when-gpu-pays-off-sweep-roofline-decomposition.png",
		"blog/when-gpu-pays-off-sweep-roofline-decision-guide.png",
	}

	for _, path := range paths {
		if err := normalize(path); err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", path, err)
			os.Exit(1)
		}
		fmt.Printf("normalized %s\n", path)
	}
}

func normalize(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	src, err := png.Decode(f)
	if closeErr := f.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}

	b := src.Bounds()
	crop := image.Rect(b.Min.X, b.Min.Y, b.Max.X, b.Min.Y+940)
	if b.Dx() != 1672 || b.Dy() != 941 {
		crop = b
	}

	dst := image.NewRGBA(image.Rect(0, 0, 1664, 936))
	scaleBilinear(dst, src, crop)

	out, err := os.Create(path)
	if err != nil {
		return err
	}
	err = png.Encode(out, dst)
	if closeErr := out.Close(); err == nil {
		err = closeErr
	}
	return err
}

func scaleBilinear(dst *image.RGBA, src image.Image, crop image.Rectangle) {
	srcW := crop.Dx()
	srcH := crop.Dy()
	dstB := dst.Bounds()
	dstW := dstB.Dx()
	dstH := dstB.Dy()

	for y := 0; y < dstH; y++ {
		sy := float64(crop.Min.Y) + (float64(y)+0.5)*float64(srcH)/float64(dstH) - 0.5
		y0 := clampInt(int(math.Floor(sy)), crop.Min.Y, crop.Max.Y-1)
		y1 := clampInt(y0+1, crop.Min.Y, crop.Max.Y-1)
		wy := sy - math.Floor(sy)

		for x := 0; x < dstW; x++ {
			sx := float64(crop.Min.X) + (float64(x)+0.5)*float64(srcW)/float64(dstW) - 0.5
			x0 := clampInt(int(math.Floor(sx)), crop.Min.X, crop.Max.X-1)
			x1 := clampInt(x0+1, crop.Min.X, crop.Max.X-1)
			wx := sx - math.Floor(sx)

			dst.SetRGBA(x, y, blend(src.At(x0, y0), src.At(x1, y0), src.At(x0, y1), src.At(x1, y1), wx, wy))
		}
	}
}

func blend(c00, c10, c01, c11 color.Color, wx, wy float64) color.RGBA {
	r00, g00, b00, a00 := rgba(c00)
	r10, g10, b10, a10 := rgba(c10)
	r01, g01, b01, a01 := rgba(c01)
	r11, g11, b11, a11 := rgba(c11)

	return color.RGBA{
		R: uint8(weighted(r00, r10, r01, r11, wx, wy)),
		G: uint8(weighted(g00, g10, g01, g11, wx, wy)),
		B: uint8(weighted(b00, b10, b01, b11, wx, wy)),
		A: uint8(weighted(a00, a10, a01, a11, wx, wy)),
	}
}

func rgba(c color.Color) (float64, float64, float64, float64) {
	r, g, b, a := c.RGBA()
	return float64(r >> 8), float64(g >> 8), float64(b >> 8), float64(a >> 8)
}

func weighted(v00, v10, v01, v11, wx, wy float64) float64 {
	top := v00*(1-wx) + v10*wx
	bottom := v01*(1-wx) + v11*wx
	return math.Round(top*(1-wy) + bottom*wy)
}

func clampInt(v, min, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}
