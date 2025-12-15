package messaging

import (
	"time"
)

// MessageStatus represents the status of a message
type MessageStatus string

const (
	MessagePending MessageStatus = "pending"
	MessageSkipped MessageStatus = "skipped"
)

// MessageRecord tracks message state
type MessageRecord struct {
	ProfileURL string
	Template   string
	Status     MessageStatus
	CreatedAt  time.Time
}

// NewMessageRecord creates a new message record
func NewMessageRecord(profileURL, template string, status MessageStatus) MessageRecord {
	return MessageRecord{
		ProfileURL: profileURL,
		Template:   template,
		Status:     status,
		CreatedAt:  time.Now(),
	}
}
