package main

import (
	"fmt"
	"os"
	"image"
	"image/png"
	_"image/jpeg"
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
		fmt.Println("\nImage Processing Menu : ")
		fmt.Println("1. Negative")
		fmt.Println("2. Grayscale")
		fmt.Println("3. Histogram Equalization")
		fmt.Println("4. SOBEL edge detection")
		fmt.Println("5. CONVOLUTION (SHARPEN)")
		fmt.Println("6. Global Thresholding")
		fmt.Println("7. Dilation")
		fmt.Println("8. Denoise")
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
			outPath = "output/output_negative.png"
		case 2:
			result = rgb2gray(loaded)
			outPath = "output/output_gray.png"
		case 3:
			result = histogramEq(loaded)
			outPath = "output/output_histeq.png"
		case 4:
			result = sobel(rgb2gray(loaded))
			outPath = "output/output_sobel.png"
		case 5:
			rm, gm, bm, am := imageToRGBMatrices(loaded)
			result = convolution(rm,gm, bm, am, sharp)
			outPath = "output/output_convolution.png"
		// case 6:
		// 	result = globalThresh(loaded)
		// 	outPath = "output_thresholding.png"
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
