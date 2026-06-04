package main

import "fmt"

func sumAll(numbers ...int) int {
	sum := 0
	for _, num := range numbers {
		sum += num
	}
	return sum
}

func main() {
	fmt.Println(sumAll(1, 2, 3))        // 6
	fmt.Println(sumAll(10, -2, 4, 7))   // 19
	fmt.Println(sumAll())               // 0
	fmt.Println(sumAll(5))              // 5
	fmt.Println(sumAll(-1, -2, -3, -4)) // -10
}
