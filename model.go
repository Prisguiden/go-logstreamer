package main

type LogMessage struct {
	Level         string
	Message       Message
	Received      string
	Time          string
	Host          string
	Environment   string
	CorrelationID string
}

type Message struct {
	Text      string
	Exception string
	Custom1   string
	Custom2   string
	Custom3   string
	Custom4   string
}

type ELKAction struct {
	Index ELKActionDescription `json:"index"`
}

type ELKActionDescription struct {
	Index string `json:"_index"`
	Type  string `json:"_type"`
}
