package main

import (
	"fmt"
	"math"
)

type Shape interface {
	Area() float64
}

type Circle struct {
	Radius float64
}

func (c Circle) Area() float64 {
	return math.Pi * c.Radius * c.Radius
}

type Rectangle struct {
	Length float64
	Width  float64
}

func (r Rectangle) Area() float64 {
	return r.Length * r.Width
}

func PrintAreas(shapes []Shape) {
	for i, shape := range shapes {
		fmt.Printf("Фигура %d: площадь = %.2f\n", i+1, shape.Area())
	}
}

func main() {
	circle := Circle{Radius: 5}
	rectangle := Rectangle{Length: 4, Width: 6}

	shapes := []Shape{circle, rectangle}

	circle2 := Circle{Radius: 3}
	rectangle2 := Rectangle{Length: 7, Width: 2}

	shapes = append(shapes, circle2, rectangle2)

	fmt.Println("Площади фигур:")
	PrintAreas(shapes)
}
