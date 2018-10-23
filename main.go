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
	doPng   bool
	size    int
)

func main() {
	flag.BoolVar(&doBlack, "black", false, "Use black backround")
	flag.BoolVar(&doPng, "png", false, "output png format instead of jpg")
	flag.IntVar(&size, "size", 128, "length of one side of the square thumbnail")
	flag.Parse()

	wg := &sync.WaitGroup{}
	jobChan := make(chan string)

	for i := 0; i < runtime.NumCPU(); i++ {
		wg.Add(1)
		go worker(wg, jobChan)
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

func worker(wg *sync.WaitGroup, jobs chan string) {
	defer wg.Done()

	for fname := range jobs {
		makeThumb(fname)
	}
}

func makeThumb(fname string) {
	file, err := os.Open(fname)
	if err != nil {
		log.Println(err)
		return
	}
	defer file.Close()

	i, _, err := image.Decode(file)
	if err != nil {
		log.Println(err)
		return
	}

	small := resize.Thumbnail(uint(size), uint(size), i, resize.Lanczos3)
	thumb := image.NewNRGBA(image.Rect(0, 0, size, size))
	draw.Draw(thumb, thumb.Bounds(), &Solid{C: color.White}, image.Pt(0, 0), draw.Src)
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
