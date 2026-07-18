package main

import (
	"bytes"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"strings"
)

const (
	targetWidth  = 1664
	targetHeight = 936
	thumbWidth   = 416
	thumbHeight  = 234
)

var files = []string{
	"blog/uniswap-v3-concentrated-liquidity-quants.png",
	"blog/uniswap-v3-concentrated-liquidity-quants-virtual-reserves.png",
	"blog/uniswap-v3-concentrated-liquidity-quants-tick-grid.png",
	"blog/uniswap-v3-concentrated-liquidity-quants-fee-accounting.png",
	"blog/uniswap-v3-concentrated-liquidity-quants-short-vol-payoff.png",
}

var targetFiles = flag.String("files", "", "optional comma-separated files to normalize")

func main() {
	flag.Parse()
	selected := files
	if *targetFiles != "" {
		selected = strings.Split(*targetFiles, ",")
	}
	thumbs := make([]image.Image, 0, len(selected))
	for _, name := range selected {
		img, err := load(name)
		if err != nil {
			fail(name, err)
		}
		finalImage := quantize(resizeBilinear(img, targetWidth, targetHeight))
		if err := save(name, finalImage); err != nil {
			fail(name, err)
		}
		final, err := load(name)
		if err != nil {
			fail(name, err)
		}
		fmt.Printf("OK %s: %dx%d\n", name, final.Bounds().Dx(), final.Bounds().Dy())
		thumbs = append(thumbs, resizeBilinear(final, thumbWidth, thumbHeight))
	}
	if len(selected) == len(files) {
		if err := writeContactSheet(thumbs); err != nil {
			fail("contact sheet", err)
		}
	}
}

func load(name string) (image.Image, error) {
	data, err := os.ReadFile(filepath.Clean(name))
	if err != nil {
		return nil, err
	}
	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	decoded := image.NewRGBA(img.Bounds())
	draw.Draw(decoded, decoded.Bounds(), img, img.Bounds().Min, draw.Src)
	return decoded, nil
}

func save(name string, img image.Image) error {
	cleanName := filepath.Clean(name)
	temp, err := os.CreateTemp(filepath.Dir(cleanName), ".normalize-*.png")
	if err != nil {
		return err
	}
	tempName := temp.Name()
	defer os.Remove(tempName)
	if err := (&png.Encoder{CompressionLevel: png.BestCompression}).Encode(temp, img); err != nil {
		temp.Close()
		return err
	}
	if err := temp.Close(); err != nil {
		return err
	}
	return os.Rename(tempName, cleanName)
}

func quantize(src image.Image) image.Image {
	dst := image.NewRGBA(src.Bounds())
	draw.Draw(dst, dst.Bounds(), src, src.Bounds().Min, draw.Src)
	for i := 0; i < len(dst.Pix); i += 4 {
		dst.Pix[i] &= 0xfc
		dst.Pix[i+1] &= 0xfc
		dst.Pix[i+2] &= 0xfc
	}
	return dst
}

func resizeBilinear(src image.Image, width, height int) image.Image {
	dst := image.NewRGBA(image.Rect(0, 0, width, height))
	b := src.Bounds()
	for y := 0; y < height; y++ {
		sy := (float64(y)+0.5)*float64(b.Dy())/float64(height) - 0.5
		y0 := int(math.Floor(sy))
		fy := sy - float64(y0)
		if y0 < 0 {
			y0, fy = 0, 0
		}
		y1 := min(y0+1, b.Dy()-1)
		for x := 0; x < width; x++ {
			sx := (float64(x)+0.5)*float64(b.Dx())/float64(width) - 0.5
			x0 := int(math.Floor(sx))
			fx := sx - float64(x0)
			if x0 < 0 {
				x0, fx = 0, 0
			}
			x1 := min(x0+1, b.Dx()-1)
			c00 := color.RGBAModel.Convert(src.At(b.Min.X+x0, b.Min.Y+y0)).(color.RGBA)
			c10 := color.RGBAModel.Convert(src.At(b.Min.X+x1, b.Min.Y+y0)).(color.RGBA)
			c01 := color.RGBAModel.Convert(src.At(b.Min.X+x0, b.Min.Y+y1)).(color.RGBA)
			c11 := color.RGBAModel.Convert(src.At(b.Min.X+x1, b.Min.Y+y1)).(color.RGBA)
			dst.SetRGBA(x, y, color.RGBA{
				R: blend(c00.R, c10.R, c01.R, c11.R, fx, fy),
				G: blend(c00.G, c10.G, c01.G, c11.G, fx, fy),
				B: blend(c00.B, c10.B, c01.B, c11.B, fx, fy),
				A: blend(c00.A, c10.A, c01.A, c11.A, fx, fy),
			})
		}
	}
	return dst
}

func blend(a, b, c, d uint8, fx, fy float64) uint8 {
	top := float64(a)*(1-fx) + float64(b)*fx
	bottom := float64(c)*(1-fx) + float64(d)*fx
	return uint8(math.Round(top*(1-fy) + bottom*fy))
}

func writeContactSheet(thumbs []image.Image) error {
	if err := os.MkdirAll("docs/screenshots", 0o755); err != nil {
		return err
	}
	const pad = 18
	sheet := image.NewRGBA(image.Rect(0, 0, 2*thumbWidth+3*pad, 3*thumbHeight+4*pad))
	draw.Draw(sheet, sheet.Bounds(), &image.Uniform{C: color.RGBA{2, 8, 23, 255}}, image.Point{}, draw.Src)
	for i, thumb := range thumbs {
		x := pad + (i%2)*(thumbWidth+pad)
		y := pad + (i/2)*(thumbHeight+pad)
		draw.Draw(sheet, image.Rect(x, y, x+thumbWidth, y+thumbHeight), thumb, image.Point{}, draw.Src)
	}
	return save("docs/screenshots/issue69-contact-sheet.png", sheet)
}

func fail(name string, err error) {
	fmt.Fprintf(os.Stderr, "%s: %v\n", name, err)
	os.Exit(1)
}
