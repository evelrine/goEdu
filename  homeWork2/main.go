package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Print("Введите строку: ")

	if scanner.Scan() {
		input := scanner.Text()

		count := len([]rune(input))

		fmt.Printf("Количество символов в строке: %d\n", count)
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Ошибка при чтении ввода:", err)
	}
}
