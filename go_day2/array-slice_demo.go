package main

import "fmt"

func main() {
	// Array: phép gán sao chép toàn bộ ba phần tử.
	arrayOriginal := [3]int{10, 20, 30}
	arrayCopy := arrayOriginal
	arrayCopy[0] = 99
	fmt.Println("arrayOriginal:", arrayOriginal)
	fmt.Println("arrayCopy:", arrayCopy)

	// Slice: sliceView và sliceOriginal dùng chung backing array nên thay đổi sliceView sẽ ảnh hưởng đến sliceOriginal.
	sliceOriginal := []int{10, 20, 30}
	sliceView := sliceOriginal[0:2]
	sliceView[0] = 99
	fmt.Println("sliceOriginal:", sliceOriginal)
	fmt.Println("sliceView:", sliceView)
	//Slice: sliceCopy được tạo ra bằng cách sao chép các phần tử từ sliceOriginal, nên thay đổi sliceCopy sẽ không ảnh hưởng đến sliceOriginal.
}