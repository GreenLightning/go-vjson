package vjson

import (
	"encoding/json"
	"strings"
	"testing"
)

type Simple struct {
	Text   string
	Number int
}

func (s Simple) MarshalJSON() ([]byte, error) {
	return Marshal(s)
}

func (s *Simple) UnmarshalJSON(data []byte) error {
	return Unmarshal(s, data)
}

type SimpleV1 struct {
	Text   string
	Number int
}

func TestMarshalNotRegistered(t *testing.T) {
	ResetRegistry()

	value := Simple{Text: "hello", Number: 42}

	_, err := json.Marshal(value)
	if err == nil {
		t.Fatal("missing error")
	}
	if !strings.Contains(err.Error(), "registered") {
		t.Error("unexpected err:", err)
	}
}

func TestUnmarshalNotRegistered(t *testing.T) {
	ResetRegistry()

	data := []byte(`{"Version":1,"Text":"hello","Number":42}`)

	var value Simple
	err := json.Unmarshal(data, &value)
	if err == nil {
		t.Fatal("missing error")
	}
	if !strings.Contains(err.Error(), "registered") {
		t.Error("unexpected err:", err)
	}
}

func TestNegativeVersion(t *testing.T) {
	ResetRegistry()
	Register(Simple{}, SimpleV1{})

	data := []byte(`{"Version":-1,"Text":"hello","Number":42}`)

	var value Simple
	err := json.Unmarshal(data, &value)
	if err == nil {
		t.Fatal("missing error")
	}
	if !strings.Contains(err.Error(), "negative") {
		t.Error("unexpected err:", err)
	}
}

func TestUnsupportedVersion(t *testing.T) {
	ResetRegistry()
	Register(Simple{}, SimpleV1{})

	data := []byte(`{"Version":100,"Text":"hello","Number":42}`)

	var value Simple
	err := json.Unmarshal(data, &value)
	if err == nil {
		t.Fatal("missing error")
	}
	if !strings.Contains(err.Error(), "unsupported") {
		t.Error("unexpected err:", err)
	}
}

func TestUnmarshalSimple(t *testing.T) {
	ResetRegistry()
	Register(Simple{}, SimpleV1{})

	data := []byte(`{"Version":1,"Text":"hello","Number":42}`)

	var value Simple
	err := json.Unmarshal(data, &value)
	if err != nil {
		t.Fatal("unexpected err:", err)
	}
	if value.Text != "hello" {
		t.Errorf("wrong text: %+v", value)
	}
	if value.Number != 42 {
		t.Errorf("wrong number: %+v", value)
	}
}

type Multiple struct {
	B string
	C string
	D string
}

func (value *Multiple) UnmarshalJSON(data []byte) error {
	return Unmarshal(value, data)
}

type MultipleV1 struct {
	A string
	B string
}

type MultipleV2 struct {
	A string
	B string
	C string
}

type MultipleV3 struct {
	B string
	C string
	D string
}

func TestUnmarshalMultiple(t *testing.T) {
	ResetRegistry()
	Register(Multiple{}, MultipleV1{}, MultipleV2{}, MultipleV3{})

	data := []byte(`{"Version":2,"A":"a","B":"b","C":"c"}`)

	var value Multiple
	err := json.Unmarshal(data, &value)
	if err != nil {
		t.Fatal("unexpected err:", err)
	}
	if value.B != "b" || value.C != "c" {
		t.Errorf("wrong value: %+v", value)
	}
}

type Renaming struct {
	X string
	Y string
}

func (value *Renaming) UnmarshalJSON(data []byte) error {
	return Unmarshal(value, data)
}

type RenamingV1 struct {
	A string
	B string
}

type RenamingV2 struct {
	X string `vjson:"A"`
	Y string
}

func TestUnmarshalRenaming(t *testing.T) {
	ResetRegistry()
	Register(Renaming{}, RenamingV1{}, RenamingV2{})

	data := []byte(`{"Version":1,"A":"x","B":"b"}`)

	var value Renaming
	err := json.Unmarshal(data, &value)
	if err != nil {
		t.Fatal("unexpected err:", err)
	}
	if value.X != "x" || value.Y != "" {
		t.Errorf("wrong value: %+v", value)
	}
}
