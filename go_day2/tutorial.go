package main

import "fmt"

// func main(){
// 	// var attempts int = 3
// 	// var port uint16 = 8080
// 	// var packetSize int64 = 4096
// 	// var score float64 = 0.95
// 	// fmt.Printf("Attempts: %d\n", attempts)
// 	// fmt.Printf("Port: %d\n", port)
// 	// fmt.Printf("Packet Size: %d\n", packetSize)
// 	// fmt.Printf("Score: %.2f\n", score)

// 	// rawPort := 700000
// 	// port := uint16(rawPort)
// 	// fmt.Println(port)    // port = 44640
// 	port, err := parsePort(70000)
// 	if err != nil {
// 		fmt.Println("Error:", err)
// 		return
// 	}
// 	fmt.Println(port)

// }

// func parsePort(rawPort int) (uint16, error){
// 	if rawPort < 1 || rawPort > 65535 {
// 		return 0, fmt.Errorf("port out of range: %d", rawPort)
// 	}
// 	return uint16(rawPort), nil
// }

func main(){
	command := "wget"
	switch command {
	case "ssh":
		fmt.Println("SSH command executed")
	case "wget":
		fmt.Println("Wget command executed")
	default:
		fmt.Println("Unknown command")
	}

	fmt.Println(command[0])
	fmt.Printf("First character: %c\n", command[0])

	var b byte = 'A'
	fmt.Printf("Byte value: %c\n", b)

	var r rune = 'ệ'
	fmt.Printf("Rune value: %c\n", r)
}