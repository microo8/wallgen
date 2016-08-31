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

	"github.com/golang/freetype/truetype"
	"github.com/microo8/wallgen/data"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

const unsplashURL string = "https://source.unsplash.com/random"

var (
	width    = flag.Int("w", 1920, "width of the image")
	height   = flag.Int("h", 1080, "height of the image")
	output   = flag.String("o", "wallpaper.png", "output file")
	text     = flag.String("t", "MEH", "printed text")
	fontFile = flag.String("font-file", "", "path to TrueType font")
	fontSize = flag.Int("font-size", 120, "Font size for the text")
	dpi      = flag.Int("dpi", 100, "DPI for the text")
)

//Flip returns a copy of input that has been flipped horizontally and vertically.
func Flip(input image.Image) image.Image {
	bounds := input.Bounds()
	newImg := image.NewRGBA(bounds)
	for x := 0; x < bounds.Max.X; x++ {
		for y := 0; y < bounds.Max.Y; y++ {
			newImg.Set(bounds.Max.X-x, bounds.Max.Y-y, input.At(x, y))
		}
	}
	return newImg
}

//Invert returns a copy of input that has its colors inverted.
func Invert(input image.Image) image.Image {
	bounds := input.Bounds()
	newImg := image.NewRGBA(bounds)

	for x := 0; x < bounds.Max.X; x++ {
		for y := 0; y < bounds.Max.Y; y++ {
			pixel := input.At(x, y).(color.RGBA)
			newImg.Set(x, y, color.RGBA{
				R: 0xff - pixel.R,
				G: 0xff - pixel.G,
				B: 0xff - pixel.B,
				A: pixel.A,
			})
		}
	}
	return newImg
}

//types enum
const (
	PNG int = iota
	JPG
)

func main() {
	flag.Parse()

	//getting output type before running everything
	var ext int
	outputLen := len(*output)
	switch {
	case outputLen < 4:
		fmt.Println("Output file must end with one of : .png/.jpg/.jpeg")
		return
	case strings.ToLower((*output)[outputLen-4:]) == ".png":
		ext = PNG
	case strings.ToLower((*output)[outputLen-4:]) == ".jpg" || strings.ToLower((*output)[outputLen-5:]) == ".jpeg":
		ext = JPG
	default:
		fmt.Println("Output file must end with one of : .png/.jpg/.jpeg")
		return
	}

	//download image
	chimg := make(chan draw.Image)
	go func() {
		resp, err := http.Get(fmt.Sprintf("%s/%dx%d", unsplashURL, *width, *height))
		defer resp.Body.Close()
		if err != nil {
			log.Fatal(err)
		}
		img, _, err := image.Decode(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		drawable := image.NewRGBA(img.Bounds())
		draw.Draw(drawable, drawable.Bounds(), img, img.Bounds().Min, draw.Src)
		chimg <- drawable
	}()

	//load font
	chmask := make(chan draw.Image)
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

		drawer := &font.Drawer{
			Dst: mask,
			Src: image.White,
			Face: truetype.NewFace(f, &truetype.Options{
				Size:    float64(*fontSize),
				DPI:     float64(*dpi),
				Hinting: font.HintingNone,
			}),
		}

		//split text rows and draw the lines
		texts := strings.Split(strings.Replace(*text, "\\n", "\n", -1), "\n")
		textHeight := int(*fontSize * *dpi / 72)
		Y := fixed.I(*height-len(texts)*textHeight) / 2

		for i, t := range texts {
			drawer.Dot = fixed.Point26_6{
				X: (fixed.I(*width) - drawer.MeasureString(t)) / 2,
				Y: Y + fixed.I(textHeight*(i+1)),
			}
			drawer.DrawBytes([]byte(t))
		}
		chmask <- mask
	}()

	img := <-chimg
	changedDst := Invert(Flip(img))

	mask := <-chmask
	draw.DrawMask(img, img.Bounds(), changedDst, image.ZP, mask, image.ZP, draw.Over)

	file, err := os.Create(*output)
	if err != nil {
		log.Fatal(err)
	}

	switch ext {
	case PNG:
		png.Encode(file, img)
	case JPG:
		jpeg.Encode(file, img, &jpeg.Options{Quality: 90})
	}
}
