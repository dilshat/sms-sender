package model

import "time"

const (
	//custom statuses
	NEW         string = "NEW"
	SUBMIT_OK          = "SM_OK"
	SUBMIT_FAIL        = "SM_FAIL"

	//delivery receipt statuses
	DELIVRD  = "DELIVRD"
	EXPIRED  = "EXPIRED"
	DELETED  = "DELETED"
	ACCEPTD  = "ACCEPTD"
	UNDELIV  = "UNDELIV"
	REJECTED = "REJECTED"
	UNKNOWN  = "UNKNOWN"
	ENROUTE  = "ENROUTE"
)

type Recipient struct {
	Id        uint32 `storm:"id,increment"`
	MessageId uint32 `storm:"index"`
	Phone     string `storm:"index"`
	Status    string
	DeliverId string    `storm:"index"`
	CreatedAt time.Time `storm:"index"`
}
