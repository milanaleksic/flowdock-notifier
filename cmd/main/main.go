package main

import (
	"fmt"
	"log"
	"os"

	"github.com/milanaleksic/igor"
)

// Version carries the program version (should be setup in compilation time to a proper value)
var Version = "undefined"

func main() {
	fmt.Printf("Igor Flowdock Notifier version %v, arguments received: %+v\n", Version, os.Args[1:])

	conf := readConfig()

	igor := igor.New(conf.MyUsername, conf.FlowdockToken)

	if !igor.IsActive() {
		fmt.Println("Turned off / outside of the working scope!")
		return
	}

	for name, lastMentioned := range igor.GetUserAndLastMention() {
		//FIXME: remove protection
		if name == "Milan" {
			igor.Answer(name, lastMentioned)
		} else {
			log.Printf("Not answering to %s", name)
		}
	}
}
