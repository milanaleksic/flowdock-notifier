package igor

import (
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/milanaleksic/flowdock"
)

// GetUserAndLastMention returns when was direction mention last received, per user that executed the mention
func GetUserAndLastMention(username, flowdockToken string) (result map[string]time.Time) {
	nameRegex := regexp.MustCompile(fmt.Sprintf("(?i)@%s", username))
	result = make(map[string]time.Time)
	client := flowdock.NewClient(flowdockToken)
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
