package main

import (
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	"image/png"
	"os"
)

func load(path string) (image.Image, string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, "", err
	}
	defer file.Close()

	img, format, err := image.Decode(file)
	if err != nil {
		return nil, "", err
	}
	return img, format, nil
}

func negative(img image.Image) image.Image {
	bounds := img.Bounds()
	out := image.NewRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			nr := uint8((65535 - r) >> 8)
			ng := uint8((65535 - g) >> 8)
			nb := uint8((65535 - b) >> 8)
			na := uint8(a >> 8)
			out.Set(x, y, color.RGBA{nr, ng, nb, na})
		}
	}
	return out
}

// histogram EQUALISATION

func rgb2gray(img image.Image) image.Image {
	b := img.Bounds()
	out := image.NewGray(b)
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			r, g, bl, _ := img.At(x, y).RGBA()
			rf := float64(r)
			gf := float64(g)
			bf := float64(bl)
			grayf := 0.299*rf + 0.587*gf + 0.114*bf
			gray := uint8(grayf / 256)
			out.Set(x, y, color.Gray{Y: gray})
		}
	}
	return out
}

func histogramEq(img image.Image) image.Image {
	bounds := img.Bounds()
	out := image.NewRGBA(bounds)

	var hist [256]int
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			nr := uint8(r >> 8)
			ng := uint8(g >> 8)
			nb := uint8(b >> 8)
			Y, _, _ := color.RGBToYCbCr(nr, ng, nb)
			hist[Y]++
		}
	}

	var cdf [256]int
	cdf[0] = hist[0]
	for i := 1; i < 256; i++ {
		cdf[i] = cdf[i-1] + hist[i]
	}
	cdfMin := 0
	for i := 0; i < 256; i++ {
		if cdf[i] != 0 {
			cdfMin = cdf[i]
			break
		}
	}

	size := bounds.Dx() * bounds.Dy()
	denom := float64(size - cdfMin)
	var lut [256]uint8
	for i := 0; i < 256; i++ {
		if denom <= 0 {
			lut[i] = uint8(i)
			continue
		}
		val := (float64(cdf[i]-cdfMin) / denom) * 255.0
		if val < 0 {
			val = 0
		}
		if val > 255 {
			val = 255
		}
		lut[i] = uint8(val)
	}
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			r8, g8, b8 := uint8(r>>8), uint8(g>>8), uint8(b>>8)
			Y, cb, cr := color.RGBToYCbCr(r8, g8, b8)
			newY := lut[Y]
			nr, ng, nb := color.YCbCrToRGB(newY, cb, cr)
			out.Set(x, y, color.RGBA{R: nr, G: ng, B: nb, A: uint8(a >> 8)})
		}
	}
	return out
}

func saveImage(path string, img image.Image) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

func main() {
	loaded, _, err := load("a.png")
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	for {
		fmt.Println("\n--- Image Processing Menu ---")
		fmt.Println("1. Negative")
		fmt.Println("2. Grayscale")
		fmt.Println("3. Histogram Equalization")
		fmt.Println("0. Exit")
		fmt.Print("Choose an option: ")

		var choice int
		_, err := fmt.Scanln(&choice)
		if err != nil {
			fmt.Println("invalid input:", err)
			continue
		}

		var result image.Image
		var outPath string

		switch choice {
		case 1:
			result = negative(loaded)
			outPath = "output_negative.png"
		case 2:
			result = rgb2gray(loaded)
			outPath = "output_gray.png"
		case 3:
			result = histogramEq(loaded)
			outPath = "output_histeq.png"
		case 0:
			fmt.Println("bye")
			return
		default:
			fmt.Println("unknown option")
			continue
		}

		err = saveImage(outPath, result)
		if err != nil {
			fmt.Println("save error:", err)
			continue
		}
		fmt.Println("saved to", outPath)
	}
}
