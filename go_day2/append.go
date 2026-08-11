package main

import "fmt"

func main() {
	values := make([]int, 0, 2)

	values = append(values, 10)
	fmt.Printf("values=%v len=%d cap=%d\n", values, len(values), cap(values))

	values = append(values, 20)
	fmt.Printf("values=%v len=%d cap=%d\n", values, len(values), cap(values))

	values = append(values, 30)
	fmt.Printf("values=%v len=%d cap=%d\n", values, len(values), cap(values))

	original := []int{1, 2, 3, 4}
	part := original[:2]
	part = append(part, 99)

	fmt.Println("original after shared append:", original)
	fmt.Println("part: ", part)

	independent := append([]int(nil), original[:2]...)
	independent = append(independent, 77)

	fmt.Println("original after cloned append:", original)
	fmt.Println("independent: ", independent)
}