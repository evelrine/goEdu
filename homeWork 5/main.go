package main

import "fmt"

type Book struct {
	Title  string
	Author string
	Year   int
}

func (b Book) GetInfo() string {
	return fmt.Sprintf("\"%s\" автор %s, %d", b.Title, b.Author, b.Year)
}

func main() {
	book1 := Book{
		Title:  "Война и мир",
		Author: "Лев Толстой",
		Year:   1869,
	}

	fmt.Println(book1.GetInfo())
}
