package flowdock

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"io/ioutil"

	"github.com/njern/httpstream"
)

const (
	flowdockAPIURL = "https://api.flowdock.com"
)

// A Client is a Flowdock API client. It should be created
// using NewClient() and provided with a valid API key.
type Client struct {
	apiKey         string
	streamClient   *httpstream.Client
	organizations  []Organization
	availableFlows []Flow // TODO: Change to map[ID]Flow
	users          map[string]User
}

// NewClient creates a new Client and automatically fetches
// information about joined flows, known users, etc.
func NewClient(apiKey string) *Client {
	client := &Client{apiKey: apiKey}

	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		var err error
		client.users, err = getUsers(apiKey)
		if err != nil {
			log.Printf("Failed to get users: %v", err)
		}
		wg.Done()
	}()

	go func() {
		var err error
		client.availableFlows, err = getFlows(apiKey)
		if err != nil {
			log.Printf("Failed to get flows: %v", err)
		}
		wg.Done()
	}()

	go func() {
		var err error
		client.organizations, err = getOrganizations(apiKey)
		if err != nil {
			log.Printf("Failed to get organizations: %v", err)
		}
		wg.Done()
	}()

	wg.Wait()

	return client
}

// Connect connects the client to the streaming API. Flowdock events will
// be sent back to the caller over the events channel, as will errors.
// If flows is nil, the stream will contain events from all the
// flows the user has joined.
// TODO: pass online state (active / idle / offline)
// TODO: Flag for whether to receive priv messages or not (now defaulting to not)
func (c *Client) Connect(flows []Flow, events chan Event) error {
	stream := make(chan []byte, 1024) // Incoming (raw) data channel
	done := make(chan error)          // Error channel
	flowURL := flowStreamURL(c, flows)

	// Set up the stream. Note we need to set a random password string ("BATMAN") or things will break.
	c.streamClient = httpstream.NewBasicAuthClient(c.apiKey, "BATMAN", func(line []byte) {
		stream <- line
	})

	// Initialize the connection
	err := c.streamClient.Connect(flowURL, done)
	if err != nil {
		return err
	}

	// Fire up a goroutine that will listen to the stream
	// and pass events back to the client.
	go func() {
		for {
			select {
			case event := <-stream:
				parsedEvent, err := unmarshalFlowdockJSONEvent(event)
				if err != nil {
					events <- err
				} else {
					events <- parsedEvent
				}

			case err := <-done:
				if err != nil {
					// TODO: Actually handle errors instead of just closing the channel
					events <- err
					close(events)
				}
			}
		}
	}()

	return nil
}

// DetailsForUser returns a User object for the given user ID.
func (c *Client) DetailsForUser(id string) User {
	return c.users[id]
}

// DetailsForFlow returns a Flow object for the given Flow ID
// or nil if the client can't access details for that flow.
func (c *Client) DetailsForFlow(id string) *Flow {
	for _, flow := range c.availableFlows {
		if flow.ID == id {
			return &flow
		}
	}

	return nil
}

// SendMessage starts a new thread in the specified Flow
// TODO: Implement this
func SendMessage(flow Flow, message string) error {
	return nil
}

// SendReply replies to an existing thread
// TODO: Implement this
func SendReply(flow Flow, reply string, threadID int64) error {
	return nil
}

// flowStreamURL creates the complete URL used to connect
// to the streaming API endpoint, including the flows filter.
func flowStreamURL(c *Client, flows []Flow) string {
	if flows == nil {
		// Add all the flows!
		flows = c.availableFlows
	}

	flowURL := "https://stream.flowdock.com/flows?filter="
	for i, flow := range flows {
		if i == len(flows)-1 {
			// Special case; Last item - no comma.
			flowURL = flowURL + flow.Organization.APIName + "/" + flow.APIName
		} else {
			flowURL = flowURL + flow.Organization.APIName + "/" + flow.APIName + ","
		}
	}
	return flowURL
}

// GetMyMentions fetches all my mentions from Flowdock API using UNOFFICIAL api
func (c *Client) GetMyMentions(limit int) ([]MentionEvent, error) {
	body, err := flowdockGET(c.apiKey, fmt.Sprintf("https://www.flowdock.com/rest/notifications/mentions?limit=%d", limit))
	if err != nil {
		return nil, err
	}

	var mentions []MentionEvent
	err = json.Unmarshal(body, &mentions)
	if err != nil {
		return nil, err
	}
	return mentions, nil
}

// GetMyUnreadPrivateMessages fetches all private unread messages from Flowdock API using UNOFFICIAL api
func (c *Client) GetMyUnreadPrivateMessages() ([]PrivateMessageEvent, error) {
	body, err := flowdockGET(c.apiKey, "https://www.flowdock.com/rest/notifications/unreads")
	if err != nil {
		return nil, err
	}

	var mentions []PrivateMessageEvent
	err = json.Unmarshal(body, &mentions)
	if err != nil {
		return nil, err
	}
	return mentions, nil
}

// RespondToPerson allows to send a private message
func (c *Client) RespondToPerson(userID int64, msg string) error {
	v := flowMessage{
		Event:   "message",
		Content: msg,
	}
	client := http.Client{}
	pushURL := fmt.Sprintf("https://api.flowdock.com/private/%d/messages", userID)

	jsonStr, err := json.Marshal(v)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", pushURL, bytes.NewBuffer(jsonStr))
	if err != nil {
		return err
	}
	req.SetBasicAuth(c.apiKey, "BATMAN")
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)

	if resp.StatusCode > 299 && resp.StatusCode < 200 {
		fmt.Printf("req = %s, resp = %+v, err = %v", jsonStr, resp, err)
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		fmt.Printf("\nresp = %s", string(data))
	}

	defer resp.Body.Close()
	if err != nil {
		return err
	}

	return nil
}

// RespondToFlow allows to send a message to a certain flow/thread
func (c *Client) RespondToFlow(flow, thread, msg string) error {
	v := flowMessage{
		Event:    "message",
		Content:  msg,
		ThreadID: thread,
		Flow:     flow,
	}
	client := http.Client{}
	pushURL := "https://api.flowdock.com/messages"

	jsonStr, err := json.Marshal(v)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", pushURL, bytes.NewBuffer(jsonStr))
	if err != nil {
		return err
	}
	req.SetBasicAuth(c.apiKey, "BATMAN")
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)

	if resp.StatusCode > 299 && resp.StatusCode < 200 {
		fmt.Printf("req = %s, resp = %+v, err = %v", jsonStr, resp, err)
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		fmt.Printf("\nresp = %s", string(data))
	}

	defer resp.Body.Close()
	if err != nil {
		return err
	}

	return nil
}
