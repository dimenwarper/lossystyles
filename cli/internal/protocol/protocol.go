package protocol

import "encoding/json"

type MessageType string

const (
	MsgInit   MessageType = "init"
	MsgLog    MessageType = "log"
	MsgFinish MessageType = "finish"
)

type Message struct {
	Type    MessageType            `json:"type"`
	RunID   string                 `json:"run_id"`
	Project string                 `json:"project,omitempty"`
	Config  map[string]interface{} `json:"config,omitempty"`
	Theme   string                 `json:"theme,omitempty"`
	Step    int                    `json:"step,omitempty"`
	Metrics map[string]float64     `json:"metrics,omitempty"`
}

func Parse(data []byte) (Message, error) {
	var msg Message
	err := json.Unmarshal(data, &msg)
	return msg, err
}
