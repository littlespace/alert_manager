package models

type EventType int

const (
	EventType_ACTIVE     EventType = 1
	EventType_EXPIRED    EventType = 2
	EventType_SUPPRESSED EventType = 3
	EventType_CLEARED    EventType = 4
	EventType_ACKD       EventType = 5
	EventType_ESCALATED  EventType = 6
)

var EventMap = map[string]EventType{
	"ACTIVE":     EventType_ACTIVE,
	"EXPIRED":    EventType_EXPIRED,
	"SUPPRESSED": EventType_SUPPRESSED,
	"CLEARED":    EventType_CLEARED,
	"ACKD":       EventType_ACKD,
	"ESCALATED":  EventType_ESCALATED,
}

func (e EventType) String() string {
	for str, ev := range EventMap {
		if e == ev {
			return str
		}
	}
	return "UNKNOWN"
}

// AlertEvent signifies a type of action on an alert
type AlertEvent struct {
	Alert *Alert
	Type  EventType
}
