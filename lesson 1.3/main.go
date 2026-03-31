package main

import (
	"fmt"
	"sort"
)

func sortNumbers(input []int) (output []int) {
	var sortNumbers []int = input
	sort.Slice(sortNumbers, func(i, j int) bool {
		return sortNumbers[i] > sortNumbers[j]
	})
	output = sortNumbers
	fmt.Println("Отсортированные элементы:", output)
	fmt.Println("Самое большое число:", output[0])
	fmt.Println("Самое маленькое число:", output[4])
	fmt.Println("Среднее-арифметическое:", (output[0]+output[1]+output[2]+output[3]+output[4])/5)
	return
}

func main() {
	sortNumbers([]int{500, 2, 40, 39, -1})
}
