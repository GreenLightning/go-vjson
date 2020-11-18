package vjson

import (
	"encoding/json"
	"testing"
)

type Simple struct {
	Text   string
	Number int
}

func (s Simple) MarshalJSON() ([]byte, error) {
	return Marshal(s)
}

type SimpleV1 struct {
	Text   string
	Number int
}

func TestNotRegistered(t *testing.T) {
	ResetRegistry()

	value := Simple{Text: "hello", Number: 42}

	_, err := json.Marshal(value)
	if err == nil {
		t.Fatal("missing error")
	}
}
