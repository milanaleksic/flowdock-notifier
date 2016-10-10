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
	if mentions, err := i.client.GetMyMentions(10); err != nil {
		log.Fatalf("Could not fetch flowdock mentions because of: %+v", err)
	} else {
		for _, mention := range mentions {
			i.addMessageToResult(mention.Message, result)
		}
	}
	if privateMessages, err := i.client.GetMyUnreadPrivateMessages(); err != nil {
		log.Fatalf("Could not fetch flowdock private message because of: %+v", err)
	} else {
		for _, privateMessage := range privateMessages {
			i.addMessageToResult(privateMessage.Message, result)
		}
	}
	return
}

func (i *Igor) addMessageToResult(message flowdock.MessageEvent, result map[string]MentionContext) {
	if message.UserID == "0" {
		// HAL or some other app
		return
	}
	if len(i.nameRegex.FindStringIndex(message.Content)) == 0 {
		// ignoring if no explicit mention
		return
	}
	mentionMoment := time.Unix(message.Timestamp/1000, 0)
	user := i.client.DetailsForUser(message.UserID)
	if mentionMoment.Before(i.database.GetActivationTimeStart()) {
		return
	}
	lastComm, err := i.database.GetLastCommunicationWith(user.Nick)
	if err != nil {
		log.Fatalf("Could not get last comm time for %s, err=%+v", user.Nick, err)
	}
	if lastComm != nil && lastComm.After(mentionMoment) {
		log.Printf("Ignoring since %v is after %v", lastComm, mentionMoment)
		return
	}
	if _, ok := result[user.Nick]; !ok {
		result[user.Nick] = MentionContext{
			Message:  message.Content,
			Moment:   time.Unix(message.Timestamp/1000, 0),
			Flow:     message.Flow,
			ThreadID: message.ThreadID,
			User:     user.Nick,
		}
	}
}

// Answer will send a message in the adequate Flow/Thread
func (i *Igor) Answer(name string, lastComm MentionContext) {
	err := i.database.SetLastCommunicationWith(name, time.Now())
	if err != nil {
		log.Fatalf("Could not write to DB, err=%+v", err)
	}
	//FIXME: remove protection
	if name != "Milan" {
		log.Printf("Not answering to %s", name)
		return
	}
	if _, err := i.database.GetResponseMessage(); err != nil {
		log.Fatalf("Could not answer to %s because of %+v", name, err)
	} else if lastComm.Flow != "" {
		log.Printf(`Answering to %s`, name)
		i.client.RespondToFlow(lastComm.Flow, lastComm.ThreadID, "Test")
	} else {
		log.Printf(`Answering to %s via private message`, name)
		//FIXME: still does not know how to send a private message
	}
}
