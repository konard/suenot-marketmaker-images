package main

import (
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
)

const (
	targetWidth  = 1664
	targetHeight = 936
)

var files = []string{
	"blog/gpu-precision-trap-fp32-backtest.png",
	"blog/gpu-precision-trap-fp32-backtest-no-fp64.png",
	"blog/gpu-precision-trap-fp32-backtest-cancellation.png",
	"blog/gpu-precision-trap-fp32-backtest-conv1d.png",
	"blog/gpu-precision-trap-fp32-backtest-fix.png",
}

func main() {
	for _, name := range files {
		if err := normalize(name); err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", name, err)
			os.Exit(1)
		}
	}
}

func normalize(name string) error {
	src, err := load(name)
	if err != nil {
		return err
	}

	crop := centerCrop16x9(src.Bounds())
	dst := image.NewRGBA(image.Rect(0, 0, targetWidth, targetHeight))
	for y := 0; y < targetHeight; y++ {
		sy := crop.Min.Y + y*crop.Dy()/targetHeight
		for x := 0; x < targetWidth; x++ {
			sx := crop.Min.X + x*crop.Dx()/targetWidth
			dst.Set(x, y, src.At(sx, sy))
		}
	}

	out, err := os.Create(filepath.Clean(name))
	if err != nil {
		return err
	}
	defer out.Close()

	if err := png.Encode(out, dst); err != nil {
		return err
	}

	fmt.Printf("%s: normalized %dx%d -> %dx%d\n", name, src.Bounds().Dx(), src.Bounds().Dy(), targetWidth, targetHeight)
	return nil
}

func load(name string) (image.Image, error) {
	f, err := os.Open(filepath.Clean(name))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	img, err := png.Decode(f)
	if err != nil {
		return nil, err
	}

	b := img.Bounds()
	rgba := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(rgba, rgba.Bounds(), img, b.Min, draw.Src)
	return rgba, nil
}

func centerCrop16x9(b image.Rectangle) image.Rectangle {
	w, h := b.Dx(), b.Dy()
	cropW, cropH := w, h
	if w*targetHeight > h*targetWidth {
		cropW = h * targetWidth / targetHeight
	} else if w*targetHeight < h*targetWidth {
		cropH = w * targetHeight / targetWidth
	}

	x0 := b.Min.X + (w-cropW)/2
	y0 := b.Min.Y + (h-cropH)/2
	return image.Rect(x0, y0, x0+cropW, y0+cropH)
}
