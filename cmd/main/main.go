package main

import (
	"log"

	"github.com/milanaleksic/igor/core"
)

// Version carries the program version (should be setup in compilation time to a proper value)
var Version = "undefined"

func main() {
	igor := core.New()
	userConfigs, err := igor.GetActiveUserConfigurations()
	if err != nil {
		log.Panicf("Could not fetch active user configurations, err=%v", err)
		return
	}

	// FIXME: parallelism
	for _, userConfig := range userConfigs {
		nonAnsweredMentions := userConfig.GetNonAnsweredMentions()
		for name, lastMentioned := range nonAnsweredMentions {
			log.Printf("Non-answered mention to: %v", name)
			igor.MarkAnswered(userConfig, name)
			if lastMentioned.Flow != "" {
				userConfig.RespondToFlow(lastMentioned.Flow, lastMentioned.ThreadID)
			} else {
				userConfig.RespondToPerson(lastMentioned.UserID)
			}
		}
	}
}
