package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
)

const (
	targetWidth  = 1664
	targetHeight = 936
	thumbWidth   = 416
	thumbHeight  = 234
)

var files = []string{
	"blog/gpu-precision-trap-fp32-backtest.png",
	"blog/gpu-precision-trap-fp32-backtest-no-fp64.png",
	"blog/gpu-precision-trap-fp32-backtest-cancellation.png",
	"blog/gpu-precision-trap-fp32-backtest-conv1d.png",
	"blog/gpu-precision-trap-fp32-backtest-fix.png",
}

func main() {
	thumbs := make([]image.Image, 0, len(files))
	for _, name := range files {
		img, err := load(name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", name, err)
			os.Exit(1)
		}
		b := img.Bounds()
		if b.Dx() != targetWidth || b.Dy() != targetHeight {
			fmt.Fprintf(os.Stderr, "%s: got %dx%d, expected %dx%d\n", name, b.Dx(), b.Dy(), targetWidth, targetHeight)
			os.Exit(1)
		}
		info, err := os.Stat(name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", name, err)
			os.Exit(1)
		}
		fmt.Printf("%s: %dx%d, %.2f MB\n", name, b.Dx(), b.Dy(), float64(info.Size())/1024/1024)
		thumbs = append(thumbs, resizeNearest(img, thumbWidth, thumbHeight))
	}

	if err := writeContactSheet(thumbs); err != nil {
		fmt.Fprintf(os.Stderr, "contact sheet: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("contact sheet: docs/screenshots/issue27-contact-sheet.png")
}

func load(name string) (image.Image, error) {
	f, err := os.Open(filepath.Clean(name))
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return png.Decode(f)
}

func resizeNearest(src image.Image, width, height int) image.Image {
	dst := image.NewRGBA(image.Rect(0, 0, width, height))
	b := src.Bounds()
	for y := 0; y < height; y++ {
		sy := b.Min.Y + y*b.Dy()/height
		for x := 0; x < width; x++ {
			sx := b.Min.X + x*b.Dx()/width
			dst.Set(x, y, src.At(sx, sy))
		}
	}
	return dst
}

func writeContactSheet(thumbs []image.Image) error {
	if err := os.MkdirAll("docs/screenshots", 0o755); err != nil {
		return err
	}

	const pad = 18
	sheetWidth := 2*thumbWidth + 3*pad
	sheetHeight := 3*thumbHeight + 4*pad
	sheet := image.NewRGBA(image.Rect(0, 0, sheetWidth, sheetHeight))
	draw.Draw(sheet, sheet.Bounds(), &image.Uniform{C: color.RGBA{8, 12, 22, 255}}, image.Point{}, draw.Src)

	for i, thumb := range thumbs {
		x := pad + (i%2)*(thumbWidth+pad)
		y := pad + (i/2)*(thumbHeight+pad)
		draw.Draw(sheet, image.Rect(x, y, x+thumbWidth, y+thumbHeight), thumb, image.Point{}, draw.Src)
	}

	out, err := os.Create("docs/screenshots/issue27-contact-sheet.png")
	if err != nil {
		return err
	}
	defer out.Close()
	return png.Encode(out, sheet)
}
