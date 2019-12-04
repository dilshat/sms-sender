package dto

type Id struct {
	Id uint32 `json:"id"`
}

type Message struct {
	Sender string   `json:"sender"`
	Text   string   `json:"text"`
	Phones []string `json:"phones"`
}

type MessageStatus struct {
	Id       uint32            `json:"id"`
	Sender   string            `json:"sender"`
	Text     string            `json:"text"`
	Statuses []RecipientStatus `json:"statuses"`
}

type RecipientStatus struct {
	Phone  string `json:"phone"`
	Status string `json:"status"`
}
