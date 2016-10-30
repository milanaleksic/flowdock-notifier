package main

import (
	"log"

	"github.com/milanaleksic/igor/core"
)

// SiteDeployment contains correct location of the website where the dpeloyment shall occur
var SiteDeployment = "undefined"

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
				if err := userConfig.RespondToFlow(lastMentioned.Flow, lastMentioned.ThreadID, SiteDeployment); err != nil {
					log.Panicf("Could not respond to flow, err=%v", err)
				}
			} else {
				if err := userConfig.RespondToPerson(lastMentioned.UserID, SiteDeployment); err != nil {
					log.Panicf("Could not respond to flow, err=%v", err)
				}
			}
		}
	}
}
