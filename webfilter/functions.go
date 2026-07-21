package main
import (
	"math"
	"image"
	"image/color"
)
func clampInt(v, min, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

func clampUint8(v float64) uint8 {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return uint8(v)
}

func imageToRGBMatrices(img image.Image) (matrix, matrix, matrix, matrix) {
	bounds:=img.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y
	rm:= make(matrix, height)
	gm:= make(matrix, height)
	bm:= make(matrix, height) 
	am:= make(matrix, height)

	for y:= range height {
		rm[y]= make([]float64, width)
		gm[y]= make([]float64, width)
		bm[y]= make([]float64, width)
		am[y]= make([]float64, width)
		
		for x:= range width {
			r, g, b, a := img.At(x,y).RGBA()
			rm[y][x]= float64(r>>8)
			gm[y][x]= float64(g>>8)
			bm[y][x]= float64(b>>8)
			am[y][x]= float64(a>>8)
		}
	}
	return rm, gm, bm, am
}

func RGBMatricesToImage(rm, gm, bm, am matrix) image.Image {
	height:=len(rm)
	width:=len(rm[0])
	out:= image.NewRGBA(image.Rect(0,0,width,height))
	for y:= range height {
		for x:= range width {
			out.Set(x, y, color.RGBA{
				R:clampUint8(rm[y][x]),
				G:clampUint8(gm[y][x]),
				B:clampUint8(bm[y][x]),
				A:clampUint8(am[y][x]),
			})
		}
	} 
	return out
}

func convolveMatrix(m matrix, k kernel) matrix {
	height := len(m)
	width := len(m[0])
	kSize := len(k)
	kRadius := kSize / 2
	out := make(matrix, height)

	for y := range height {
		out[y] = make([]float64, width)
		for x := range width {
			var pixval float64
			for j := range kSize {
				for i := range kSize {
					iy := clampInt(y+(j-kRadius), 0, height-1)
					ix := clampInt(x+(i-kRadius), 0, width-1)
					pixval += m[iy][ix] * k[j][i]
				}
			}
			out[y][x] = pixval
		}
	}
	return out
}

func convolution(rm, gm, bm, am matrix, k kernel) image.Image {
	crm := convolveMatrix(rm, k)
	cgm := convolveMatrix(gm, k)
	cbm := convolveMatrix(bm, k)
	
	return RGBMatricesToImage(crm, cgm, cbm, am)
}


func rgb2gray(img image.Image) *image.Gray {
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

func sobel(grey *image.Gray) image.Image {
	b := grey.Bounds()
	out := image.NewGray(b)

	for y := b.Min.Y + 1; y < b.Max.Y-1; y++ {
		for x := b.Min.X + 1; x < b.Max.X-1; x++ {
			pixelGx:= 0
			pixelGy:= 0

			for i := -1; i <= 1; i++ {
				for j := -1; j <= 1; j++ {
					val := int(grey.GrayAt(x+i, y+j).Y)
					pixelGx += val * Gx[i+1][j+1]
					pixelGy += val * Gy[i+1][j+1]
				}
			}
			mag := math.Sqrt(float64(pixelGx*pixelGx + pixelGy*pixelGy))
			if mag > 255 {
				mag = 255
			}
			out.SetGray(x, y, color.Gray{Y: uint8(mag)})
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

