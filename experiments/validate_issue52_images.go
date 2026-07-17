package main

import (
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"os"
	"path/filepath"
)

const (
	targetWidth  = 1664
	targetHeight = 936
	thumbWidth   = 416
	thumbHeight  = 234
	minBytes     = 500 * 1024
	maxBytes     = 1500 * 1024
)

var files = []string{
	"blog/dcc-garch-dynamic-correlation-crypto.png",
	"blog/dcc-garch-dynamic-correlation-crypto-correlation-breakdown.png",
	"blog/dcc-garch-dynamic-correlation-crypto-two-step.png",
	"blog/dcc-garch-dynamic-correlation-crypto-hedge-ratio.png",
	"blog/dcc-garch-dynamic-correlation-crypto-regime-signal.png",
}

func main() {
	thumbs := make([]image.Image, 0, len(files))
	for _, name := range files {
		img, err := load(name)
		if err != nil {
			fail(name, err)
		}
		if img.Bounds().Dx() != targetWidth || img.Bounds().Dy() != targetHeight {
			img = resizeBilinear(img, targetWidth, targetHeight)
			if err := save(name, quantize(img)); err != nil {
				fail(name, err)
			}
		}

		img, err = load(name) // Decode the final file, not only the generator output.
		if err != nil {
			fail(name, err)
		}
		b := img.Bounds()
		if b.Dx() != targetWidth || b.Dy() != targetHeight {
			fail(name, fmt.Errorf("got %dx%d, expected %dx%d", b.Dx(), b.Dy(), targetWidth, targetHeight))
		}
		if err := rejectTextChunks(name); err != nil {
			fail(name, err)
		}
		info, err := os.Stat(name)
		if err != nil {
			fail(name, err)
		}
		if info.Size() < minBytes || info.Size() > maxBytes {
			fail(name, fmt.Errorf("%.2f MB is outside the 0.5-1.5 MB quality band", float64(info.Size())/1024/1024))
		}
		fmt.Printf("OK %s: %dx%d, %.2f MB\n", name, b.Dx(), b.Dy(), float64(info.Size())/1024/1024)
		thumbs = append(thumbs, resizeBilinear(img, thumbWidth, thumbHeight))
	}

	if err := writeContactSheet(thumbs); err != nil {
		fail("contact sheet", err)
	}
	fmt.Println("OK docs/screenshots/issue52-contact-sheet.png")
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

func fail(name string, err error) {
	fmt.Fprintf(os.Stderr, "%s: %v\n", name, err)
	os.Exit(1)
}

func load(name string) (image.Image, error) {
	f, err := os.Open(filepath.Clean(name))
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return png.Decode(f)
}

func save(name string, img image.Image) error {
	f, err := os.Create(filepath.Clean(name))
	if err != nil {
		return err
	}
	defer f.Close()
	encoder := png.Encoder{CompressionLevel: png.BestCompression}
	return encoder.Encode(f, img)
}

func rejectTextChunks(name string) error {
	data, err := os.ReadFile(filepath.Clean(name))
	if err != nil {
		return err
	}
	for offset := 8; offset+12 <= len(data); {
		length := int(binary.BigEndian.Uint32(data[offset : offset+4]))
		if offset+12+length > len(data) {
			return fmt.Errorf("truncated PNG chunk")
		}
		kind := string(data[offset+4 : offset+8])
		if kind == "tEXt" || kind == "zTXt" || kind == "iTXt" {
			return fmt.Errorf("contains unexpected %s metadata", kind)
		}
		offset += 12 + length
	}
	return nil
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
	return save("docs/screenshots/issue52-contact-sheet.png", sheet)
}
