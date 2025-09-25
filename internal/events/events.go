package events

import "time"

type EventType string

const (
	EventJoined    EventType = "joined"
	EventLeft      EventType = "left"
	EventBroadcast EventType = "broadcast"
)

type Event struct {
	Type    EventType `json:"type,omitempty"`
	Sender  string    `json:"sender,omitempty"`
	Payload []byte    `json:"payload,omitempty"`
	At      time.Time `json:"at"`
}
