// // go:build ignore
// package main

// import "fmt"

// type kernel [][]int

// var sharp = kernel{
// 	{0, -1, 0},
// 	{-1, 5, -1},
// 	{0, -1, 0},
// }

// var image = kernel{
// 	{52, 55, 61, 59, 79, 61, 76, 61, 80, 93, 91, 90, 79, 75, 67},
// 	{63, 65, 66, 113, 144, 104, 63, 62, 59, 55, 90, 109, 85, 69, 71},
// 	{62, 59, 68, 113, 144, 104, 66, 63, 58, 71, 109, 85, 69, 71, 71},
// 	{63, 58, 71, 109, 104, 67, 65, 68, 62, 118, 92, 71, 71, 62, 65},
// 	{87, 79, 68, 66, 65, 67, 63, 68, 79, 91, 88, 71, 68, 62, 68},
// 	{68, 68, 68, 68, 68, 65, 62, 68, 88, 91, 79, 65, 63, 67, 68},
// 	{68, 63, 65, 68, 68, 68, 65, 68, 66, 71, 65, 62, 63, 62, 65},
// 	{62, 65, 68, 68, 63, 62, 62, 65, 65, 66, 66, 63, 62, 65, 62},
// 	{65, 65, 65, 68, 65, 63, 62, 63, 65, 68, 65, 65, 65, 65, 65},
// 	{63, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65},
// 	{65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65},
// 	{65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65},
// 	{65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65},
// 	{65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65},
// 	{65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65},
// }

func convolution(k kernel) kernel {
	var out kernel
	for x := 1; x < 14; x++ {
		for y := 1; y < 14; y++ {
			pixval := 0
			a := 0
			for i := x - 1; i <= x+1; i++ {
				b := 0
				for j := y - 1; j <= y+1; j++ {
					pixval += k[i][j] * sharp[a][b]
					b++
				}
				a++
			}
			k[x][y] = pixval
		}
	}
	return out
}

func convolutioN(img image.Image, k kernel) [][]float64 {
	var out [][]float64
	grayscale := rgb2gray(img)
	bounds := grayscale.bounds()
	for y:= bounds.Min.Y+1; y<bounds.Max.Y-1; y++ {
		for x:=bounds.Min.X+1; x<bounds.Max.Y-1; x++ {
			pixval:=0
			a:=a
			for i:=x - 1; i<= x+1; i++ {
				b:=0 
				for j:= y-1; j <= y+1; j++ {
					grayY := grayscale.GrayAt(i,j).Y
					pixval += grayY * k[a][b]
					b++
				}
				a++
			}
			out[x][y] = pixval
		}
	}
	return out
}

// func printMatrix(k kernel) {
// 	for i := 0; i < 15; i++ {
// 		for j := 0; j < 15; j++ {
// 			fmt.Printf("%4d ", k[i][j])
// 		}
// 		fmt.Println()
// 	}
// }

// func main() {
// 	result := convolution(image)

// 	fmt.Println("\nConvolved:")
// 	printMatrix(result)
// }
