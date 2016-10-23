package main

import (
	"fmt"
	"os"

	"log"

	"github.com/milanaleksic/igor/core"
)

// Version carries the program version (should be setup in compilation time to a proper value)
var Version = "undefined"

func main() {
	fmt.Printf("Igor Flowdock Notifier version %v, arguments received: %+v\n", Version, os.Args[1:])

	igor := core.New()
	userConfigs, err := igor.GetActiveUserConfigurations()
	if err != nil {
		log.Panicf("Could not fetch active user configurations, err=%v", err)
		return
	}

	// FIXME: parallelism
	for _, userConfig := range userConfigs {
		for name, lastMentioned := range userConfig.GetNonAnsweredMentions() {
			igor.MarkAnswered(userConfig, name)
			if lastMentioned.Flow != "" {
				userConfig.RespondToFlow(lastMentioned.Flow, lastMentioned.ThreadID)
			} else {
				userConfig.RespondToPerson(lastMentioned.UserID)
			}
		}
	}
}
