package main

import (
	"fmt"
	"time"
)

func main() {

	for {
		fmt.Println("Hello from streaming-service")
		time.Sleep(5 * time.Second)
	}
}
