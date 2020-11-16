package vjson

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

type Hardcoded struct {
	Text1   string
	Text2   string
	Text3   string
	Text4   string
	Text5   string
	Number0 int
	Number1 int
	Number2 int
	Number3 int
	Number4 int
	Number5 int
}

type HardcodedV1 struct {
	Version int
	Text1   string
	Text2   string
	Text3   string
	Text4   string
	Number0 int
	Number1 int
	Number2 int
	Number3 int
	Number4 int
}

type HardcodedV2 struct {
	Version   int
	Text1     string
	Text2     string
	Text3     string
	Text4     string
	ExtraText string
	Number0   int
	Number1   int
	Number2   int
	Number3   int
	Number4   int
	Number5   int
}

type HardcodedV3 struct {
	Version int
	Text1   string
	Text2   string
	Text3   string
	Text4   string
	Text5   string
	Number0 int
	Number1 int
	Number2 int
	Number3 int
	Number4 int
	Number5 int
}

func (hardcoded Hardcoded) MarshalJSON() ([]byte, error) {
	latest := HardcodedV3{
		Version: 3,
		Text1:   hardcoded.Text1,
		Text2:   hardcoded.Text2,
		Text3:   hardcoded.Text3,
		Text4:   hardcoded.Text4,
		Text5:   hardcoded.Text5,
		Number0: hardcoded.Number0,
		Number1: hardcoded.Number1,
		Number2: hardcoded.Number2,
		Number3: hardcoded.Number3,
		Number4: hardcoded.Number4,
		Number5: hardcoded.Number5,
	}
	return json.Marshal(latest)
}

func (hardcoded *Hardcoded) UnmarshalJSON(bytes []byte) error {
	version, err := unmarshalVersion(bytes)
	if err != nil {
		return err
	}

	var v1 HardcodedV1
	var v2 HardcodedV2
	var v3 HardcodedV3

	switch version {
	case 1:
		err = json.Unmarshal(bytes, &v1)
	case 2:
		err = json.Unmarshal(bytes, &v2)
	case 3:
		err = json.Unmarshal(bytes, &v3)
	default:
		err = fmt.Errorf("unsupported version: %d", version)
	}

	if err != nil {
		return err
	}

	switch version {
	case 1:
		v2.Text1 = v1.Text1
		v2.Text2 = v1.Text2
		v2.Text3 = v1.Text3
		v2.Text4 = v1.Text4
		v2.Number0 = v1.Number0
		v2.Number1 = v1.Number1
		v2.Number2 = v1.Number2
		v2.Number3 = v1.Number3
		v2.Number4 = v1.Number4
		fallthrough
	case 2:
		v3.Text1 = v2.Text1
		v3.Text2 = v2.Text2
		v3.Text3 = v2.Text3
		v3.Text4 = v2.Text4
		v3.Text5 = v2.ExtraText
		v3.Number0 = v2.Number0
		v3.Number1 = v2.Number1
		v3.Number2 = v2.Number2
		v3.Number3 = v2.Number3
		v3.Number4 = v2.Number4
		v3.Number5 = v2.Number5
		fallthrough
	case 3:
	}

	latest := &v3

	hardcoded.Text1 = latest.Text1
	hardcoded.Text2 = latest.Text2
	hardcoded.Text3 = latest.Text3
	hardcoded.Text4 = latest.Text4
	hardcoded.Text5 = latest.Text5
	hardcoded.Number0 = latest.Number0
	hardcoded.Number1 = latest.Number1
	hardcoded.Number2 = latest.Number2
	hardcoded.Number3 = latest.Number3
	hardcoded.Number4 = latest.Number4
	hardcoded.Number5 = latest.Number5

	return nil
}

type Dynamic struct {
	Text1   string
	Text2   string
	Text3   string
	Text4   string
	Text5   string
	Number0 int
	Number1 int
	Number2 int
	Number3 int
	Number4 int
	Number5 int
}

type DynamicV1 struct {
	Version int
	Text1   string
	Text2   string
	Text3   string
	Text4   string
	Number0 int
	Number1 int
	Number2 int
	Number3 int
	Number4 int
}

type DynamicV2 struct {
	Version   int
	Text1     string
	Text2     string
	Text3     string
	Text4     string
	ExtraText string
	Number0   int
	Number1   int
	Number2   int
	Number3   int
	Number4   int
	Number5   int
}

type DynamicV3 struct {
	Version int
	Text1   string
	Text2   string
	Text3   string
	Text4   string
	Text5   string
	Number0 int
	Number1 int
	Number2 int
	Number3 int
	Number4 int
	Number5 int
}

func (dynamic Dynamic) MarshalJSON() ([]byte, error) {
	return Marshal(dynamic)
}

func TestHardcodedMarshal(t *testing.T) {
	value := Hardcoded{
		Text1:   "hello",
		Text2:   "hello",
		Text3:   "hello",
		Text4:   "hello",
		Text5:   "hello",
		Number0: 42,
		Number1: 42,
		Number2: 42,
		Number3: 42,
		Number4: 42,
		Number5: 42,
	}

	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal("unexpected err:", err)
	}

	str := string(data)
	if !strings.Contains(str, `"Version":3`) {
		t.Error("wrong data:", str)
	}
	if strings.Count(str, `"Version"`) != 1 {
		t.Error("wrong data:", str)
	}
}

func TestDynamicMarshal(t *testing.T) {
	ResetRegistry()
	Register(Dynamic{}, DynamicV1{}, DynamicV2{}, DynamicV3{})

	value := Dynamic{
		Text1:   "hello",
		Text2:   "hello",
		Text3:   "hello",
		Text4:   "hello",
		Text5:   "hello",
		Number0: 42,
		Number1: 42,
		Number2: 42,
		Number3: 42,
		Number4: 42,
		Number5: 42,
	}

	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal("unexpected err:", err)
	}

	str := string(data)
	if !strings.Contains(str, `"Version":3`) {
		t.Error("wrong data:", str)
	}
	if strings.Count(str, `"Version"`) != 1 {
		t.Error("wrong data:", str)
	}
}

func BenchmarkHardcodedMarshal(b *testing.B) {
	for i := 0; i < b.N; i++ {
		value := Hardcoded{
			Text1:   "hello",
			Text2:   "hello",
			Text3:   "hello",
			Text4:   "hello",
			Text5:   "hello",
			Number0: i,
			Number1: i,
			Number2: i,
			Number3: i,
			Number4: i,
			Number5: i,
		}
		_, err := json.Marshal(value)
		if err != nil {
			b.Fatal("unexpected err:", err)
		}
	}
}

func BenchmarkDynamicMarshal(b *testing.B) {
	ResetRegistry()
	Register(Dynamic{}, DynamicV1{}, DynamicV2{}, DynamicV3{})
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		value := Dynamic{
			Text1:   "hello",
			Text2:   "hello",
			Text3:   "hello",
			Text4:   "hello",
			Text5:   "hello",
			Number0: i,
			Number1: i,
			Number2: i,
			Number3: i,
			Number4: i,
			Number5: i,
		}
		_, err := json.Marshal(value)
		if err != nil {
			b.Fatal("unexpected err:", err)
		}
	}
}
