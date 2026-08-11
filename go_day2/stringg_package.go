package main

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

func main(){
	rawPayload := []byte( "WGET http://example.com/a.sh\r\n")
	trimmedPayload := bytes.TrimSpace(rawPayload)

	if !bytes.Contains(trimmedPayload, []byte("http://")){
		fmt.Println("payload does not contain an HTTP URL")
		return
	}

	fields := strings.Fields(string(trimmedPayload))
	if(len(fields) == 0){
		fmt.Println("empty command")
		return
	}

	command := strings.ToLower(fields[0])
	args := append([]string(nil), fields[1:]...)

	port, err := strconv.Atoi("8080")
	if err != nil {
		fmt.Println("invalid port:", err)
		return
	}
	if port < 1 || port > 65535 {
		fmt.Printf("port out of range: %d\n", port)
		return
	}
	enabled, err := strconv.ParseBool("true")
	if err != nil {
		fmt.Println("invalid enable flag:", err)
		return
	}
	maxPayload, err := strconv.ParseInt("4096", 10, 64)
	if err != nil {
		fmt.Println("invalid payload size:", err)
		return
	}

	fmt.Printf("command=%s args=%v\n", command, args)
	fmt.Printf("port=%s enabled=%t maxPayload=%d\n", strconv.Itoa(port), enabled, maxPayload)
}