package users

import (
	"encoding/json"
)

type Status int

const (
	UndefinedStatus Status = iota
	PendingStatus
	InvitedStatus
	ActiveStatus
	InactiveStatus
	DeletedStatus
	endStatus
)

func (s Status) String() string {
	switch s {
	case PendingStatus:
		return "pending"
	case InvitedStatus:
		return "invited"
	case ActiveStatus:
		return "active"
	case InactiveStatus:
		return "inactive"
	case DeletedStatus:
		return "deleted"
	default:
		return ""
	}
}

func (s Status) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}
