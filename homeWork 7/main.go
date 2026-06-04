package main

import (
	"fmt"
	"sync"
)

func main() {
	var wg sync.WaitGroup

	for i := 1; i <= 3; i++ {
		wg.Add(1)
		go func(id int) {
			fmt.Printf("Горутина %d начала работу\n", id)
			fmt.Printf("Горутина %d завершила работу\n", id)
			wg.Done()
		}(i)
	}

	wg.Wait()
	fmt.Println("Все горутины завершили работу")
}
