package model

import "time"

type Message struct {
	Id        uint32 `storm:"id,increment"`
	Text      string
	Sender    string
	CreatedAt time.Time `storm:"index"`
}
