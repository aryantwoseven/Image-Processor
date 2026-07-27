package main

import "image"

type filterfunc func(image.Image) (image.Image, error)
type conv func(rm, gm, bm matrix, k kernel) (image.Image, error)

var filterRegistry = map[string]filterfunc{
	"negative": func(img image.Image) (image.Image, error) {
		return negative(img), nil
	},
	"grayscale": func(img image.Image) (image.Image, error) {
		return rgb2gray(img), nil
	},
	"histogrameq": func(img image.Image) (image.Image, error) {
		return histogramEq(img), nil
	},
	"sobel": func(img image.Image) (image.Image, error) {
		return sobel(rgb2gray(img)), nil
	},
	"sharp": func(img image.Image) (image.Image, error) {
		rm, gm, bm, am := imageToRGBMatrices(img)
		return convolution(rm, gm, bm, am, sharp), nil
	},
	"blur": func(img image.Image) (image.Image, error) {
		rm, gm, bm, am := imageToRGBMatrices(img)
		return convolution(rm, gm, bm, am, boxblur5), nil
	},
}
