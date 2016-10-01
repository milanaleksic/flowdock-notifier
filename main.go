package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("Hello World from Go!")
	fmt.Printf("Success, arguments received: %+v", os.Args)
}
