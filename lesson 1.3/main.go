package main

import (
	"fmt"
	"sort"
)

func sortNumbers(input []int) {
	sort.Slice(input, func(i, j int) bool {
		return input[i] > input[j]
	})

	fmt.Println("Отсортированные элементы:", input[0], input[1], input[2], input[3], input[4])
	fmt.Println("Самое большое число:", input[0])
	fmt.Println("Самое маленькое число:", input[4])

	sum := input[0] + input[1] + input[2] + input[3] + input[4]
	average := float64(sum) / 5.0
	fmt.Printf("Среднее-арифметическое: %.2f\n", average)
}

func main() {
	sortNumbers([]int{50, 2, 41, 3, 10})
}
