package main

import (
	"fmt"
	"log"
	"os"
	"time"

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
	database.SetLastCommunicationWith("Test", time.Now())
	if moment, err := database.GetLastCommunicationWith("Test"); err != nil {
		log.Fatal("Could not get last comm time for Test")
	} else {
		log.Println("Last moment for Test is: ", moment)
	}
	if moment, err := database.GetLastCommunicationWith("Test2"); err != nil {
		log.Fatal("Could not get last comm time for Test2")
	} else {
		log.Println("Last moment for Test2 is: ", moment)
	}
}
