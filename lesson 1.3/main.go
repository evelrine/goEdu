package main

import (
	"fmt"
	"sort"
)

func sortNumbers(input []int) (output []int) {
	var sortNumbers []int = input
	sort.Slice(sortNumbers, func(i, j int) bool {
		return sortNumbers[i] > sortNumbers[j] // от большего к меньшему
	})
	output = sortNumbers // присваиваем результат в output
	fmt.Println(output)
	return
}

func main() {
	sortNumbers([]int{5, 2, 4, 3, 1}) // передаём срез чисел
}
