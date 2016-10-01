package main

import (
	"fmt"
	"os"
)

func main() {
	event := os.Args[1]
	fmt.Printf("Success, event received for this Lambda: %s", event)
}
