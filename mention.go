package igor

import "time"

// MentionContext gives context when in which way last mention was made
type MentionContext struct {
	Message  string
	Moment   time.Time
	Flow     string
	ThreadID string
	User     string
	UserID   int64
}
