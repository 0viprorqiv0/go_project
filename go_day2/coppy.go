package main

import "fmt"

func clonePayload(payload []byte) []byte {
	if payload == nil{
		return nil
	}
	cloned := make([]byte, len(payload))
	copy(cloned, payload)
	return cloned
}

func main(){
	source := []byte("payload")
	fullCopy := make([]byte, len(source))
	fullCount := copy(fullCopy, source)

	shortCopy := make([]byte, 3)
	shortCount := copy(shortCopy, source)

	cloned := clonePayload(source)
	source[0] = 'P'

	fmt.Printf("full.copy: %q, copied=%d\n", fullCopy, fullCount)
	fmt.Printf("short.copy: %q, copied=%d\n", shortCopy, shortCount)
	fmt.Printf("source: %q\n", source)
	fmt.Printf("clone %q\n", cloned)
}