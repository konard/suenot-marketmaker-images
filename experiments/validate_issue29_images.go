package main

import (
	"fmt"
	"image/png"
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

	failed := false
	for _, path := range paths {
		f, err := os.Open(path)
		if err != nil {
			fmt.Printf("FAIL %s: open: %v\n", path, err)
			failed = true
			continue
		}
		img, err := png.Decode(f)
		closeErr := f.Close()
		if err != nil {
			fmt.Printf("FAIL %s: decode: %v\n", path, err)
			failed = true
			continue
		}
		if closeErr != nil {
			fmt.Printf("FAIL %s: close: %v\n", path, closeErr)
			failed = true
			continue
		}

		bounds := img.Bounds()
		width := bounds.Dx()
		height := bounds.Dy()
		stat, err := os.Stat(path)
		if err != nil {
			fmt.Printf("FAIL %s: stat: %v\n", path, err)
			failed = true
			continue
		}
		if width != 1664 || height != 936 {
			fmt.Printf("FAIL %s: got %dx%d, want 1664x936\n", path, width, height)
			failed = true
			continue
		}
		if stat.Size() < 400*1024 || stat.Size() > 4*1024*1024 {
			fmt.Printf("FAIL %s: suspicious size %d bytes\n", path, stat.Size())
			failed = true
			continue
		}
		fmt.Printf("OK %s: %dx%d, %.1f KiB\n", path, width, height, float64(stat.Size())/1024)
	}

	if failed {
		os.Exit(1)
	}
}
