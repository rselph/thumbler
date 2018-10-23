package main

import (
	"flag"
	"image"
	"image/color"
	"image/draw"
	_ "image/gif"
	"image/jpeg"
	"image/png"
	"log"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/nfnt/resize"

	_ "golang.org/x/image/tiff"
)

var (
	doBlack bool
	doWhite bool
	doPng   bool
	size    int
)

func main() {
	flag.BoolVar(&doBlack, "black", false, "use black background")
	flag.BoolVar(&doWhite, "white", false, "make PNG background white instead of clear")
	flag.BoolVar(&doPng, "png", false, "output png format instead of jpg")
	flag.IntVar(&size, "size", 128, "length of one side of the square thumbnail")
	flag.Parse()

	wg := &sync.WaitGroup{}
	jobChan := make(chan string)

	for i := 0; i < runtime.NumCPU(); i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for fname := range jobChan {
				makeThumb(fname)
			}
		}()
	}

	for _, glob := range flag.Args() {
		fnames, _ := filepath.Glob(glob)
		for _, fname := range fnames {
			jobChan <- fname
		}
	}

	close(jobChan)
	wg.Wait()
}

func makeThumb(fname string) {
	file, err := os.Open(fname)
	if err != nil {
		log.Println(err)
		return
	}
	defer file.Close()

	original, _, err := image.Decode(file)
	if err != nil {
		log.Println(err)
		return
	}

	var background color.Color
	switch {
	case doWhite:
		background = color.White
	case doBlack:
		background = color.Black
	case doPng:
		background = &color.RGBA{}
	default:
		background = color.White
	}

	small := resize.Thumbnail(uint(size), uint(size), original, resize.Lanczos3)
	thumb := image.NewNRGBA(image.Rect(0, 0, size, size))
	draw.Draw(thumb, thumb.Bounds(), &Solid{C: background}, image.Pt(0, 0), draw.Src)
	draw.Draw(thumb, thumb.Bounds(), small,
		image.Pt(-(size-small.Bounds().Dx())/2, -(size-small.Bounds().Dy())/2),
		draw.Src)

	outFileName := fname + ".thumb"
	if doPng {
		outFileName += ".png"
	} else {
		outFileName += ".jpg"
	}
	outFile, err := os.Create(outFileName)
	if err != nil {
		log.Println(err)
		return
	}
	defer outFile.Close()
	if doPng {
		png.Encode(outFile, thumb)
	} else {
		jpeg.Encode(outFile, thumb, &jpeg.Options{Quality: 80})
	}
}

type Solid struct {
	C color.Color
}

func (s *Solid) ColorModel() color.Model {
	return color.RGBAModel
}

func (s *Solid) Bounds() image.Rectangle {
	return image.Rect(math.MinInt32, math.MinInt32, math.MaxInt32, math.MaxInt32)
}

func (s *Solid) At(x, y int) color.Color {
	return s.C
}
