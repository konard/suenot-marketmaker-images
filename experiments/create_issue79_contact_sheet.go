package main

import (
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"os"
)

var names = []string{
	"blog/maker-taker-fees-rebates-execution.png",
	"blog/maker-taker-fees-rebates-execution-break-even.png",
	"blog/maker-taker-fees-rebates-execution-adverse-selection.png",
	"blog/maker-taker-fees-rebates-execution-fee-tiers.png",
	"blog/maker-taker-fees-rebates-execution-queue-value.png",
}

func main() {
	const width, height, gap = 832, 468, 24
	sheet := image.NewRGBA(image.Rect(0, 0, width*2+gap*3, height*3+gap*4))
	draw.Draw(sheet, sheet.Bounds(), &image.Uniform{color.RGBA{3, 7, 18, 255}}, image.Point{}, draw.Src)

	for i, name := range names {
		file, err := os.Open(name)
		if err != nil {
			log.Fatal(err)
		}
		img, err := png.Decode(file)
		file.Close()
		if err != nil {
			log.Fatal(err)
		}
		x, y := gap+(i%2)*(width+gap), gap+(i/2)*(height+gap)
		for dy := 0; dy < height; dy++ {
			for dx := 0; dx < width; dx++ {
				sheet.Set(x+dx, y+dy, img.At(img.Bounds().Min.X+dx*img.Bounds().Dx()/width, img.Bounds().Min.Y+dy*img.Bounds().Dy()/height))
			}
		}
	}

	if err := os.MkdirAll("docs/screenshots", 0o755); err != nil {
		log.Fatal(err)
	}
	out, err := os.Create("docs/screenshots/issue79-contact-sheet.png")
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()
	if err := png.Encode(out, sheet); err != nil {
		log.Fatal(err)
	}
}
