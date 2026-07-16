package main

import (
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	"image/png"
	"math"
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
// DILATION
// EDGE DETECTION
// GLOBAL THRESHOLDING

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

	// 2. CDF
	var cdf [256]int
	cdf[0] = hist[0]
	for i := 1; i < 256; i++ {
		cdf[i] = cdf[i-1] + hist[i]
	}

	// 3. cdfMin
	cdfMin := 0
	for i := 0; i < 256; i++ {
		if cdf[i] != 0 {
			cdfMin = cdf[i]
			break
		}
	}

	// 4. LUT
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

	// 5. apply LUT
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
func histogramEqPerChannel(img image.Image) image.Image {
	bounds := img.Bounds()
	out := image.NewRGBA(bounds)

	// separate histograms for R, G, B
	var histR, histG, histB [256]int
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			histR[uint8(r>>8)]++
			histG[uint8(g>>8)]++
			histB[uint8(b>>8)]++
		}
	}

	size := bounds.Dx() * bounds.Dy()

	buildLUT := func(hist [256]int) [256]uint8 {
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
		return lut
	}

	lutR := buildLUT(histR)
	lutG := buildLUT(histG)
	lutB := buildLUT(histB)

	// apply each channel's own LUT independently
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			nr := lutR[uint8(r>>8)]
			ng := lutG[uint8(g>>8)]
			nb := lutB[uint8(b>>8)]
			out.Set(x, y, color.RGBA{R: nr, G: ng, B: nb, A: uint8(a >> 8)})
		}
	}
	return out
}

// RGB (0-255 each) -> HSV (H: 0-360, S: 0-1, V: 0-1)
func rgbToHSV(r, g, b uint8) (h, s, v float64) {
	rf, gf, bf := float64(r)/255, float64(g)/255, float64(b)/255
	max := math.Max(rf, math.Max(gf, bf))
	min := math.Min(rf, math.Min(gf, bf))
	delta := max - min

	v = max

	if max == 0 {
		s = 0
	} else {
		s = delta / max
	}

	if delta == 0 {
		h = 0
	} else if max == rf {
		h = 60 * math.Mod((gf-bf)/delta, 6)
	} else if max == gf {
		h = 60 * ((bf-rf)/delta + 2)
	} else {
		h = 60 * ((rf-gf)/delta + 4)
	}
	if h < 0 {
		h += 360
	}
	return h, s, v
}

// HSV -> RGB (0-255 each)
func hsvToRGB(h, s, v float64) (r, g, b uint8) {
	c := v * s
	x := c * (1 - math.Abs(math.Mod(h/60, 2)-1))
	m := v - c

	var rf, gf, bf float64
	switch {
	case h < 60:
		rf, gf, bf = c, x, 0
	case h < 120:
		rf, gf, bf = x, c, 0
	case h < 180:
		rf, gf, bf = 0, c, x
	case h < 240:
		rf, gf, bf = 0, x, c
	case h < 300:
		rf, gf, bf = x, 0, c
	default:
		rf, gf, bf = c, 0, x
	}

	r = uint8(math.Round((rf + m) * 255))
	g = uint8(math.Round((gf + m) * 255))
	b = uint8(math.Round((bf + m) * 255))
	return r, g, b
}

func histogramEqHSV(img image.Image) image.Image {
	bounds := img.Bounds()
	out := image.NewRGBA(bounds)

	// 1. histogram of V, scaled to 0-255 bins
	var hist [256]int
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			r8, g8, b8 := uint8(r>>8), uint8(g>>8), uint8(b>>8)
			_, _, v := rgbToHSV(r8, g8, b8)
			vBin := uint8(math.Round(v * 255))
			hist[vBin]++
		}
	}

	// 2. CDF
	var cdf [256]int
	cdf[0] = hist[0]
	for i := 1; i < 256; i++ {
		cdf[i] = cdf[i-1] + hist[i]
	}

	// 3. cdfMin
	cdfMin := 0
	for i := 0; i < 256; i++ {
		if cdf[i] != 0 {
			cdfMin = cdf[i]
			break
		}
	}

	// 4. LUT: old V-bin -> new V-bin
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

	// 5. apply LUT to V only, keep H and S
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			r8, g8, b8 := uint8(r>>8), uint8(g>>8), uint8(b>>8)
			h, s, v := rgbToHSV(r8, g8, b8)

			vBin := uint8(math.Round(v * 255))
			newV := float64(lut[vBin]) / 255.0

			nr, ng, nb := hsvToRGB(h, s, newV)
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
			result = histogramEqHSV(loaded)
			outPath = "output_histeq_hsv.png"
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
