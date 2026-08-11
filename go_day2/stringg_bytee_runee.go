package main

import (
	"fmt"
	"unicode/utf8"
)

func main(){
	text := "Việt"
	raw := []byte(text)
	runes := []rune(text)

	fmt.Printf("text=%q\n", text)
	fmt.Printf("bytes=%v\n", raw)
	fmt.Printf("runes=%U\n", runes)
	fmt.Printf("len(string)=%d bytes\n", len(text))
	fmt.Printf("rune count=%d\n", utf8.RuneCountInString(text))

	for byteIndex, r := range text {
		fmt.Printf(
			"byteIndex=%d rune=%q codePoint=%u\n", byteIndex, r, r,
		)
	}

	invalidPayload := []byte{0xff, 0xfe, 'A'}
	fmt.Printf("valid UTF-8 = %t raw=%v\n", utf8.Valid(invalidPayload), invalidPayload)
}