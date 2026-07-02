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
	"blog/deflated-sharpe-multiple-testing.png",
	"blog/deflated-sharpe-multiple-testing-the-search-machine.png",
	"blog/deflated-sharpe-multiple-testing-the-toolkit.png",
	"blog/deflated-sharpe-multiple-testing-calibration.png",
	"blog/deflated-sharpe-multiple-testing-power-curve.png",
	"blog/deflated-sharpe-multiple-testing-correlated-grids.png",
	"blog/deflated-sharpe-multiple-testing-two-questions.png",
}

func main() {
	for _, name := range files {
		if err := validate(name); err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", name, err)
			os.Exit(1)
		}
	}
}

func validate(name string) error {
	f, err := os.Open(filepath.Clean(name))
	if err != nil {
		return err
	}
	defer f.Close()

	img, err := png.Decode(f)
	if err != nil {
		return err
	}

	b := img.Bounds()
	if b.Dx() != targetWidth || b.Dy() != targetHeight {
		return fmt.Errorf("got %dx%d, want %dx%d", b.Dx(), b.Dy(), targetWidth, targetHeight)
	}

	if !hasSignal(img) {
		return fmt.Errorf("image appears blank or nearly uniform")
	}

	info, err := f.Stat()
	if err != nil {
		return err
	}

	fmt.Printf("%s: ok %dx%d %.1f MB\n", name, b.Dx(), b.Dy(), float64(info.Size())/(1024*1024))
	return nil
}

func hasSignal(img image.Image) bool {
	b := img.Bounds()
	dst := image.NewRGBA(b)
	draw.Draw(dst, b, img, b.Min, draw.Src)

	var minR, minG, minB uint8 = 255, 255, 255
	var maxR, maxG, maxB uint8
	for y := b.Min.Y; y < b.Max.Y; y += 16 {
		for x := b.Min.X; x < b.Max.X; x += 16 {
			i := dst.PixOffset(x, y)
			r, g, bl := dst.Pix[i], dst.Pix[i+1], dst.Pix[i+2]
			if r < minR {
				minR = r
			}
			if g < minG {
				minG = g
			}
			if bl < minB {
				minB = bl
			}
			if r > maxR {
				maxR = r
			}
			if g > maxG {
				maxG = g
			}
			if bl > maxB {
				maxB = bl
			}
		}
	}

	return int(maxR-minR)+int(maxG-minG)+int(maxB-minB) > 80
}
