package main

import (
	"fmt"
	"sort"
)

func sortNumbers(input []int) []int {
	sort.Slice(input, func(i, j int) bool {
		return input[i] > input[j]
	})
	return input
}

func main() {
	numbers := make([]int, 5)

	fmt.Println("Введите 5 натуральных чисел:")

	for i := 0; i < 5; i++ {
		fmt.Printf("Число %d: ", i+1)
		fmt.Scan(&numbers[i])
	}

	sortedNumbers := sortNumbers(numbers)

	fmt.Println("Отсортированные элементы:", sortedNumbers[0], sortedNumbers[1], sortedNumbers[2], sortedNumbers[3], sortedNumbers[4])
	fmt.Println("Самое большое число:", sortedNumbers[0])
	fmt.Println("Самое маленькое число:", sortedNumbers[4])

	sum := sortedNumbers[0] + sortedNumbers[1] + sortedNumbers[2] + sortedNumbers[3] + sortedNumbers[4]
	average := float64(sum) / 5.0
	fmt.Printf("Среднее-арифметическое: %.2f\n", average)
}
