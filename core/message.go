package core

type Message struct {
	Type string `json:"type"`

	Result string `json:"result,omitempty"`

	Progress uint64 `json:"progress,omitempty"`
}
