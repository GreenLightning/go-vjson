package vjson

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

type Hardcoded struct {
	Text1, Text2, Text3, Text4, Text5 string
	Num1, Num2, Num3, Num4, Num5      int
}

type HardcodedV1 struct {
	Version                    int
	Text1, Text2, Text3, Text4 string
	Num1, Num2, Num3, Num4     int
}

type HardcodedV2 struct {
	Version                               int
	Text1, Text2, Text3, Text4, ExtraText string
	Num1, Num2, Num3, Num4, Num5          int
}

type HardcodedV3 struct {
	Version                           int
	Text1, Text2, Text3, Text4, Text5 string
	Num1, Num2, Num3, Num4, Num5      int
}

func (hardcoded Hardcoded) MarshalJSON() ([]byte, error) {
	latest := HardcodedV3{
		Version: 3,
		Text1:   hardcoded.Text1,
		Text2:   hardcoded.Text2,
		Text3:   hardcoded.Text3,
		Text4:   hardcoded.Text4,
		Text5:   hardcoded.Text5,
		Num1:    hardcoded.Num1,
		Num2:    hardcoded.Num2,
		Num3:    hardcoded.Num3,
		Num4:    hardcoded.Num4,
		Num5:    hardcoded.Num5,
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
		v2.Num1 = v1.Num1
		v2.Num2 = v1.Num2
		v2.Num3 = v1.Num3
		v2.Num4 = v1.Num4
		fallthrough
	case 2:
		v3.Text1 = v2.Text1
		v3.Text2 = v2.Text2
		v3.Text3 = v2.Text3
		v3.Text4 = v2.Text4
		v3.Text5 = v2.ExtraText
		v3.Num1 = v2.Num1
		v3.Num2 = v2.Num2
		v3.Num3 = v2.Num3
		v3.Num4 = v2.Num4
		v3.Num5 = v2.Num5
		fallthrough
	case 3:
	}

	latest := &v3

	hardcoded.Text1 = latest.Text1
	hardcoded.Text2 = latest.Text2
	hardcoded.Text3 = latest.Text3
	hardcoded.Text4 = latest.Text4
	hardcoded.Text5 = latest.Text5
	hardcoded.Num1 = latest.Num1
	hardcoded.Num2 = latest.Num2
	hardcoded.Num3 = latest.Num3
	hardcoded.Num4 = latest.Num4
	hardcoded.Num5 = latest.Num5

	return nil
}

type Dynamic struct {
	Text1, Text2, Text3, Text4, Text5 string
	Num1, Num2, Num3, Num4, Num5      int
}

type DynamicV1 struct {
	Text1, Text2, Text3, Text4 string
	Num1, Num2, Num3, Num4     int
}

type DynamicV2 struct {
	Text1, Text2, Text3, Text4, ExtraText string
	Num1, Num2, Num3, Num4, Num5          int
}

type DynamicV3 struct {
	Text1, Text2, Text3, Text4, Text5 string
	Num1, Num2, Num3, Num4, Num5      int
}

func (dynamic Dynamic) MarshalJSON() ([]byte, error) {
	return Marshal(dynamic)
}

type DynamicOptimized struct {
	Text1, Text2, Text3, Text4, Text5 string
	Num1, Num2, Num3, Num4, Num5      int
}

type DynamicOptimizedV1 struct {
	Version                    int
	Text1, Text2, Text3, Text4 string
	Num1, Num2, Num3, Num4     int
}

type DynamicOptimizedV2 struct {
	Version                               int
	Text1, Text2, Text3, Text4, ExtraText string
	Num1, Num2, Num3, Num4, Num5          int
}

type DynamicOptimizedV3 struct {
	Version                           int
	Text1, Text2, Text3, Text4, Text5 string
	Num1, Num2, Num3, Num4, Num5      int
}

func (dynamic DynamicOptimized) MarshalJSON() ([]byte, error) {
	return Marshal(dynamic)
}

func TestBenchMarshal(t *testing.T) {
	test := func(t *testing.T, value interface{}) {
		data, err := json.Marshal(value)
		if err != nil {
			t.Fatal("unexpected err:", err)
		}

		str := string(data)
		if strings.Count(str, `"Version"`) != 1 {
			t.Fatal("wrong data:", str)
		}
		if !strings.Contains(str, `"Version":3`) {
			t.Fatal("wrong data:", str)
		}
		if !strings.Contains(str, `"Text1":"hello"`) {
			t.Fatal("wrong data:", str)
		}
		if !strings.Contains(str, `"Num1":42`) {
			t.Fatal("wrong data:", str)
		}
	}

	t.Run("Hardcoded", func(t *testing.T) {
		test(t, Hardcoded{
			Text1: "hello", Text2: "hello", Text3: "hello", Text4: "hello", Text5: "hello",
			Num1: 42, Num2: 42, Num3: 42, Num4: 42, Num5: 42,
		})
	})
	t.Run("Dynamic", func(t *testing.T) {
		ResetRegistry()
		Register(Dynamic{}, DynamicV1{}, DynamicV2{}, DynamicV3{})
		test(t, Dynamic{
			Text1: "hello", Text2: "hello", Text3: "hello", Text4: "hello", Text5: "hello",
			Num1: 42, Num2: 42, Num3: 42, Num4: 42, Num5: 42,
		})
	})
	t.Run("DynamicOptimized", func(t *testing.T) {
		ResetRegistry()
		Register(DynamicOptimized{}, DynamicOptimizedV1{}, DynamicOptimizedV2{}, DynamicOptimizedV3{})
		test(t, DynamicOptimized{
			Text1: "hello", Text2: "hello", Text3: "hello", Text4: "hello", Text5: "hello",
			Num1: 42, Num2: 42, Num3: 42, Num4: 42, Num5: 42,
		})
	})
}

func BenchmarkMarshal(b *testing.B) {
	bench := func(b *testing.B, value interface{}) {
		for i := 0; i < b.N; i++ {
			_, err := json.Marshal(value)
			if err != nil {
				b.Fatal("unexpected err:", err)
			}
		}
	}

	b.Run("Hardcoded", func(b *testing.B) {
		bench(b, Hardcoded{
			Text1: "hello", Text2: "hello", Text3: "hello", Text4: "hello", Text5: "hello",
			Num1: 42, Num2: 42, Num3: 42, Num4: 42, Num5: 42,
		})
	})
	b.Run("Dynamic", func(b *testing.B) {
		ResetRegistry()
		Register(Dynamic{}, DynamicV1{}, DynamicV2{}, DynamicV3{})
		bench(b, Dynamic{
			Text1: "hello", Text2: "hello", Text3: "hello", Text4: "hello", Text5: "hello",
			Num1: 42, Num2: 42, Num3: 42, Num4: 42, Num5: 42,
		})
	})
	b.Run("DynamicOptimized", func(b *testing.B) {
		ResetRegistry()
		Register(DynamicOptimized{}, DynamicOptimizedV1{}, DynamicOptimizedV2{}, DynamicOptimizedV3{})
		bench(b, DynamicOptimized{
			Text1: "hello", Text2: "hello", Text3: "hello", Text4: "hello", Text5: "hello",
			Num1: 42, Num2: 42, Num3: 42, Num4: 42, Num5: 42,
		})
	})
}
