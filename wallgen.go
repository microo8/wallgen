/* This program is free software. It comes without any warranty, to
 * the extent permitted by applicable law. You can redistribute it
 * and/or modify it under the terms of the Do What The Fuck You Want
 * To Public License, Version 2, as published by Sam Hocevar. See
 * http://www.wtfpl.net/ for more details. */

package main

import (
	"encoding/hex"
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
	"strconv"
	"strings"

	"github.com/golang/freetype/truetype"
	"github.com/microo8/wallgen/data"
	"github.com/nfnt/resize"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

const unsplashURL string = "https://source.unsplash.com/random"

var (
	width     = flag.Int("w", 1920, "width of the image")
	height    = flag.Int("h", 1080, "height of the image")
	output    = flag.String("o", "wallpaper.png", "output file")
	text      = flag.String("t", "MEH", "printed text")
	query     = flag.String("q", "", "keyword to search in unsplash")
	fontColor = flag.String("c", "", "optional text color")
	fontFile  = flag.String("font-file", "", "path to TrueType font")
	fontSize  = flag.Int("font-size", 120, "Font size for the text")
	dpi       = flag.Int("dpi", 100, "DPI for the text")
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

//rgbTranslate interprets color settings for text RGB input
func rgbTranslate(input string) []int {
	rgbOut := []int{}
	split := strings.FieldsFunc(input, func(r rune) bool {
		switch r {
		case ',', ' ', '/', '-':
			return true
		}
		return false
	})
	for i := 0; i < len(split); i++ {
		value, _ := strconv.ParseInt(split[i], 10, 0)
		rgbOut = append(rgbOut, int(value))
	}
	return rgbOut
}

//hexTranslate interprets color settings for texxt hex input
func hexTranslate(input string) []int {
	rgbOut := []int{}
	for i := 0; i <= 4; i += 2 {
		value, _ := hex.DecodeString(strings.ToUpper(input[i : i+2]))
		rgbOut = append(rgbOut, int(value[0]))
	}
	return rgbOut
}

//Color changes text colors to the user-defined value.
func Color(input image.Image) image.Image {
	bounds := input.Bounds()
	newImg := image.NewRGBA(bounds)

	var rgb []int
	if len(*fontColor) == 6 {
		rgb = hexTranslate(*fontColor)
	} else {
		rgb = rgbTranslate(*fontColor)
	}

	for i := 0; i < len(rgb); i++ {
		if 255 < rgb[i] || rgb[i] < 0 {
			fmt.Println("Incorrect hex color input")
			os.Exit(1)
		}
	}

	for x := 0; x < bounds.Max.X; x++ {
		for y := 0; y < bounds.Max.Y; y++ {
			newImg.Set(x, y, color.RGBA{
				R: uint8(rgb[0]),
				G: uint8(rgb[1]),
				B: uint8(rgb[2]),
				A: uint8(255),
			})
		}
	}
	return newImg
}

//Evaluator determines text invert of user color change
func Evaluator(input image.Image) image.Image {
	if *fontColor != "" {
		return Color(input)
	}
	return Invert(Flip(input))
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
		resp, err := http.Get(fmt.Sprintf("%s/%dx%d/?%s", unsplashURL, *width, *height, *query))
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()
		img, _, err := image.Decode(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		if img.Bounds().Dx() != *width || img.Bounds().Dy() != *height {
			img = resize.Resize(uint(*width), uint(*height), img, resize.Bicubic)
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

	changedDst := Evaluator(img)

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
