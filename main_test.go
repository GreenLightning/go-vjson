package vjson

import (
	"encoding/json"
	"strings"
	"testing"
)

type Regular struct {
	Text   string
	Number int
}

func TestRegular(t *testing.T) {
	value := Regular{Text: "hello", Number: 42}

	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal("unexpected err:", err)
	}

	if string(data) != `{"Text":"hello","Number":42}` {
		t.Error("wrong data:", string(data))
	}
}

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

func TestSimple(t *testing.T) {
	ResetRegistry()

	Register(Simple{}, SimpleV1{})

	value := Simple{Text: "hello", Number: 42}

	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal("unexpected err:", err)
	}

	str := string(data)
	if strings.Count(str, `"Version"`) != 1 {
		t.Error("wrong data:", str)
	}
	if !strings.Contains(str, `"Version":1`) {
		t.Error("wrong data:", str)
	}
	if !strings.Contains(str, `"Text":"hello"`) {
		t.Error("wrong data:", str)
	}
	if !strings.Contains(str, `"Number":42`) {
		t.Error("wrong data:", str)
	}
}
