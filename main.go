package main

import (
	"fmt"
	"os"

	"github.com/milanaleksic/flowdock-notifier/db"
)

// Version carries the program version (should be setup in compilation time to a proper value)
var Version = "undefined"

func main() {
	fmt.Printf("Flowdock Notifier version %v, arguments received: %+v\n", Version, os.Args[1:])
	database := db.New()
	if !database.IsActive() {
		fmt.Println("Application is not active!")
		return
	}
	fmt.Println("Seems that we are good to go!")
}
