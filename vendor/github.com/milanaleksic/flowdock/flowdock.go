// Package flowdock contains helpful structs and
// methods for dealing with Flowdock's RESTful API's.
// Structs are based on the message types defined here: https://www.flowdock.com/api/message-types
package flowdock

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// flowdockGET is a convenience function for performing
// GET requests against the Flowdock API
func flowdockGET(apiKey, url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(apiKey, "BATMAN")
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

type flowMessage struct {
	Event    string `json:"event"`
	Content  string `json:"content"`
	ThreadID string `json:"thread_id,omitempty"`
	Flow     string `json:"flow"`
}

func pushMessage(flowAPIKey, message, threadID, flowTokenForPosting string) error {
	v := flowMessage{
		Event:    "message",
		Content:  message,
		ThreadID: threadID,
		Flow:     flowAPIKey,
	}
	client := http.Client{}
	pushURL := fmt.Sprintf("https://api.flowdock.com/messages?flow_token=%s", flowTokenForPosting)

	jsonStr, err := json.Marshal(v)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", pushURL, bytes.NewBuffer(jsonStr))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	//fmt.Printf("req = %s, resp = %+v, err = %v", jsonStr, resp, err)
	//data, err := ioutil.ReadAll(resp.Body)
	//if err != nil {
	//	return err
	//}
	//fmt.Printf("\nresp = %s", string(data))
	defer resp.Body.Close()
	if err != nil {
		return err
	}

	return nil
}

// PushMessageToFlowWithKey can uses the Flowdock "Push" API to start
// a new thread in a flow using any pseudonym the client wishes. Useful
// for e.g implementing bots.
func PushMessageToFlowWithKey(flowAPIKey, message string) error {
	return pushMessage(flowAPIKey, message, "", "")
}

// ReplyToThreadInFlowWithKey is similar to PushMessageToFlowWithKey
// except that it is used for replies rather than starting a new thread.
func ReplyToThreadInFlowWithKey(flowAPIKey, message, threadID, flowTokenForPosting string) error {
	return pushMessage(flowAPIKey, message, threadID, flowTokenForPosting)
}
