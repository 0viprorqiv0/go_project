package main

import "fmt"

func firstN(data []byte, n int) []byte {
	if n <= 0 || len(data) == 0 {
		return nil
	}
	if n > len(data) {
		n = len(data)
	}
	return data[:n]
}

func payload(packet []byte, declaredLength int) ([]byte, error) {
	if declaredLength < 0 || declaredLength > len(packet) {
		return nil, fmt.Errorf("invalid payload length %d for packet of length %d", declaredLength, len(packet))
	}
	return packet[:declaredLength], nil
}

func main() {
	data := make([]byte, 4, 8)
	copy(data, []byte("ABCD"))

	fmt.Printf("data=%q len=%d cap %d\n", data, len(data), cap(data))
	fmt.Printf("firstN(data, 2)=%q\n", firstN(data, 2))
	fmt.Printf("firstN(data, 99)=%q\n", firstN(data, 99))
	fmt.Printf("firstN(nil, 2)=%q\n", firstN(nil, 2))

	body, err := payload(data, 3)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Printf("body=%q\n", body)

	_, err = payload(data, 100)
	if err != nil {
		fmt.Println("rejected:", err)
	}
}
