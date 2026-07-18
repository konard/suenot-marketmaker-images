package main

import (
	"image"
	"image/draw"
	"image/png"
	"log"
	"os"
)

var names = []string{
	"almgren-chriss-optimal-execution.png",
	"almgren-chriss-optimal-execution-impact-vs-risk.png",
	"almgren-chriss-optimal-execution-sinh-trajectory.png",
	"almgren-chriss-optimal-execution-efficient-frontier.png",
	"almgren-chriss-optimal-execution-calibration.png",
}

func main() {
	const thumbW, thumbH, gap = 832, 468, 16
	sheet := image.NewRGBA(image.Rect(0, 0, thumbW*2+gap, thumbH*3+gap*2))

	for i, name := range names {
		f, err := os.Open("blog/" + name)
		if err != nil {
			log.Fatal(err)
		}
		src, err := png.Decode(f)
		f.Close()
		if err != nil {
			log.Fatal(err)
		}
		x0 := (i % 2) * (thumbW + gap)
		y0 := (i / 2) * (thumbH + gap)
		for y := 0; y < thumbH; y++ {
			for x := 0; x < thumbW; x++ {
				sheet.Set(x0+x, y0+y, src.At(src.Bounds().Min.X+x*2, src.Bounds().Min.Y+y*2))
			}
		}
	}

	out, err := os.Create("docs/screenshots/issue65-contact-sheet.png")
	if err != nil {
		log.Fatal(err)
	}
	err = png.Encode(out, sheet)
	if closeErr := out.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		log.Fatal(err)
	}
	_ = draw.Src
}
