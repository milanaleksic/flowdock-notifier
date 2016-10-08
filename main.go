package main

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"time"

	"github.com/milanaleksic/flowdock"
	"github.com/milanaleksic/flowdock-notifier/db"
)

// Version carries the program version (should be setup in compilation time to a proper value)
var Version = "undefined"

func main() {
	fmt.Printf("Flowdock Notifier version %v, arguments received: %+v\n", Version, os.Args[1:])

	conf := readConfig()

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

	for name, lastMentioned := range userAndLastMention(conf) {
		log.Printf("A Mention by: %s %1.f hours ago ", name, time.Since(lastMentioned).Hours())
	}
}

func userAndLastMention(conf config) (result map[string]time.Time) {
	nameRegex := regexp.MustCompile(fmt.Sprintf("(?i)@%s", conf.MyUsername))
	result = make(map[string]time.Time)
	client := flowdock.NewClient(conf.FlowdockToken)
	if mentions, err := client.GetMyMentions(50); err != nil {
		log.Fatalf("Could not fetch flowdock mentions because of: %+v", err)
	} else {
		for _, mention := range mentions {
			if mention.Message.UserID == "0" {
				// HAL or some other app
				continue
			}
			if len(nameRegex.FindStringIndex(mention.Message.Content)) == 0 {
				// ignoring if no explicit mention
				continue
			}
			user := client.DetailsForUser(mention.Message.UserID)
			if _, ok := result[user.Nick]; !ok {
				result[user.Nick] = time.Unix(mention.Message.Timestamp/1000, 0)
			}
		}
	}
	return
}
