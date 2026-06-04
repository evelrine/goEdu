package main

import (
	"fmt"
	"math/rand"
	"sort"
	"time"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	arr := [10]int{}
	for i := range arr {
		arr[i] = rand.Intn(100) + 1
	}

	slice := make([]int, len(arr))
	copy(slice, arr[:])

	sort.Ints(slice)

	fmt.Printf("Исходный массив: %v\n", arr)
	fmt.Printf("Отсортированный слайс: %v\n", slice)
}
