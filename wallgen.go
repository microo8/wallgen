package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/golang/freetype/truetype"
	"github.com/microo8/wallgen/data"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

const unsplashUrl string = "https://source.unsplash.com/random"

var (
	width    = flag.Int("w", 1920, "width of the image")
	height   = flag.Int("h", 1080, "height of the image")
	output   = flag.String("o", "wallpaper.png", "output file")
	text     = flag.String("t", "MEH", "printed text")
	fontFile = flag.String("font-file", "", "path to TrueType font")
	fontSize = flag.Int("font-size", 120, "Font size for the text")
	dpi      = flag.Int("dpi", 100, "DPI for the text")
)

type pixelColor struct {
	r, g, b, a uint32
}

func (c pixelColor) RGBA() (uint32, uint32, uint32, uint32) {
	return c.r, c.g, c.b, c.a
}

//Flip returns a copy of input that has been flipped horizontally and vertically.
func Flip(input image.Image) image.Image {

	var wg sync.WaitGroup
	//create new image
	bounds := input.Bounds()
	newImg := image.NewRGBA(bounds)
	for x := 0; x < bounds.Max.X; x++ {
		x := x
		wg.Add(1)
		go func() {
			defer wg.Done()
			for y := 0; y < bounds.Max.Y; y++ {
				newImg.Set(bounds.Max.X-x, bounds.Max.Y-y, input.At(x, y))
			}
		}()
	}
	wg.Wait()

	return newImg
}

//InvertColors returns a copy of input that has its colors inverted.
func InvertColors(input image.Image) image.Image {

	var wg sync.WaitGroup
	//create new image
	bounds := input.Bounds()
	newImg := image.NewRGBA(bounds)

	var currentPixelColor color.Color
	var r, g, b, a uint32
	for x := 0; x < bounds.Max.X; x++ {
		x := x
		wg.Add(1)
		go func() {
			defer wg.Done()
			for y := 0; y < bounds.Max.Y; y++ {
				r, g, b, a = input.At(x, y).RGBA()
				currentPixelColor = pixelColor{
					r: 0xffff - r,
					g: 0xffff - g,
					b: 0xffff - b,
					a: a,
				}
				newImg.Set(x, y, currentPixelColor)
			}
		}()
	}
	wg.Wait()

	return newImg
}

func main() {
	flag.Parse()

	chimg := make(chan image.Image)

	go func() {
		resp, err := http.Get(fmt.Sprintf("%s/%dx%d", unsplashUrl, *width, *height))
		if err != nil {
			log.Fatal(err)
		}
		img, err := jpeg.Decode(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		chimg <- img
	}()

	//load font
	chmask := make(chan image.Image)
	go func() {
		var fontBytes []byte
		var err error
		if *fontFile != "" {
			fontBytes, err = ioutil.ReadFile(*fontFile)
		} else {
			fontBytes, err = ubuntu.Asset("Ubuntu-B.ttf")
		}
		if err != nil {
			log.Fatal(err)
		}
		f, err := truetype.Parse(fontBytes)
		if err != nil {
			log.Fatal(err)
		}

		//generate text mast
		mask := image.NewRGBA(image.Rect(0, 0, *width, *height))

		d := &font.Drawer{
			Dst: mask,
			Src: image.White,
			Face: truetype.NewFace(f, &truetype.Options{
				Size:    float64(*fontSize),
				DPI:     float64(*dpi),
				Hinting: font.HintingNone,
			}),
		}

		texts := strings.Split(*text, "\\n")
		textHeight := int(*fontSize * *dpi / 72)
		Y := fixed.I(*height-len(texts)*textHeight) / 2

		for i, t := range texts {
			d.Dot = fixed.Point26_6{
				X: (fixed.I(*width) - d.MeasureString(t)) / 2,
				Y: Y + fixed.I(textHeight*(i+1)),
			}
			d.DrawBytes([]byte(t))
		}
		chmask <- mask
	}()

	mask := <-chmask
	img := <-chimg

	finalDst := image.NewRGBA(img.Bounds())
	changedDst := InvertColors(Flip(img))

	//Convert dst
	dstB := img.Bounds()
	draw.Draw(finalDst, finalDst.Bounds(), img, dstB.Min, draw.Src)
	draw.DrawMask(finalDst, finalDst.Bounds(), changedDst, image.ZP, mask, image.ZP, draw.Over)

	file, err := os.Create(*output)
	if err != nil {
		log.Fatal(err)
	}
	outputLen := len(*output)
	switch {
	case outputLen < 4:
		fmt.Println("Output file must end with one of : .png/.jpg/.jpeg")
	case strings.ToLower((*output)[outputLen-4:]) == ".png":
		png.Encode(file, finalDst)
	case strings.ToLower((*output)[outputLen-4:]) == ".jpg" || strings.ToLower((*output)[outputLen-4:]) == ".jpeg":
		jpeg.Encode(file, finalDst, &jpeg.Options{Quality: 90})
	}
}
