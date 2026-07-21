package main

type matrix [][]float64
type kernel [][]float64

var sharp = kernel{
	{0, -1, 0},
	{-1, 5, -1},
	{0, -1, 0},
}
var Gx = [3][3]int{
	{-1, 0, 1},
	{-2, 0, 2},
	{-1, 0, 1},
}

var Gy = [3][3]int{
	{-1, -2, -1},
	{ 0,  0,  0},
	{ 1,  2,  1},
}
var gx = kernel{
	{-1,  0,  1},
	{-2,  0,  2},
	{-1,  0,  1},
}
var gy = kernel{
	{-1, -2, -1},
	{ 0,  0,  0},
	{ 1,  2,  1},
}
