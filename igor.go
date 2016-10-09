package igor

import (
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/milanaleksic/flowdock"
	"github.com/milanaleksic/igor/db"
)

// MentionContext gives context when in which way last mention was made
type MentionContext struct {
	Message  string
	Moment   time.Time
	Flow     string
	ThreadID string
	User     string
}

// Igor is the main entrypoint to work with the library
type Igor struct {
	client    *flowdock.Client
	nameRegex *regexp.Regexp
	database  *db.DB
}

// New creates new Igor based on Flowdock username and API token
func New(username, flowdockToken string) *Igor {
	return &Igor{
		nameRegex: regexp.MustCompile(fmt.Sprintf("(?i)@%s", username)),
		client:    flowdock.NewClient(flowdockToken),
		database:  db.New(),
	}
}

// IsActive will return true if in the current moment in time configuration says
// we should react on the notifications/mentions
func (i *Igor) IsActive() bool {
	return i.database.IsActive()
}

// GetUserAndLastMention returns when was direction mention last received, per user that executed the mention
func (i *Igor) GetUserAndLastMention() (result map[string]MentionContext) {
	result = make(map[string]MentionContext)
	if mentions, err := i.client.GetMyMentions(50); err != nil {
		log.Fatalf("Could not fetch flowdock mentions because of: %+v", err)
	} else {
		for _, mention := range mentions {
			if mention.Message.UserID == "0" {
				// HAL or some other app
				continue
			}
			if len(i.nameRegex.FindStringIndex(mention.Message.Content)) == 0 {
				// ignoring if no explicit mention
				continue
			}
			moment := time.Unix(mention.Message.Timestamp/1000, 0)
			user := i.client.DetailsForUser(mention.Message.UserID)
			if moment.Before(i.database.GetActivationTimeStart()) {
				continue
			}
			lastComm, err := i.database.GetLastCommunicationWith(user.Nick)
			if err != nil {
				log.Fatalf("Could not get last comm time for %s, err=%+v", user.Nick, err)
			}
			if lastComm != nil && lastComm.After(moment) {
				continue
			}
			if _, ok := result[user.Nick]; !ok {
				result[user.Nick] = MentionContext{
					Message:  mention.Message.Content,
					Moment:   time.Unix(mention.Message.Timestamp/1000, 0),
					Flow:     mention.Message.Flow,
					ThreadID: mention.Message.ThreadID,
					User:     user.Nick,
				}
			}
		}
	}
	return
}

// Answer will send a message in the adequate Flow/Thread
func (i *Igor) Answer(name string, lastComm MentionContext) {
	err := i.database.SetLastCommunicationWith(name, time.Now())
	if err != nil {
		log.Fatalf("Could not write to DB, err=%+v", err)
	}
	if msg, err := i.database.GetResponseMessage(); err != nil {
		log.Fatalf("Could not answer to %s because of %+v", name, err)
	} else {
		log.Printf(`Answering to %s "%s"`, name, msg)
		//TODO: send the answer!
	}
}
