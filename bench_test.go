package vjson

import (
	"encoding/json"
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

type HardcodedByValue Hardcoded

func (hardcoded HardcodedByValue) MarshalJSON() ([]byte, error) {
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

type HardcodedByPointer Hardcoded

func (hardcoded *HardcodedByPointer) MarshalJSON() ([]byte, error) {
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

type DynamicOptimizedV3 struct {
	Version                           int
	Text1, Text2, Text3, Text4, Text5 string
	Num1, Num2, Num3, Num4, Num5      int
}

type DynamicByValue Dynamic

func (dynamic DynamicByValue) MarshalJSON() ([]byte, error) {
	return Marshal(dynamic)
}

type DynamicByPointer Dynamic

func (dynamic *DynamicByPointer) MarshalJSON() ([]byte, error) {
	return Marshal(dynamic)
}

func TestMarshal(t *testing.T) {
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

	t.Run("HardcodedByValue", func(t *testing.T) {
		test(t, HardcodedByValue{
			Text1: "hello", Text2: "hello", Text3: "hello", Text4: "hello", Text5: "hello",
			Num1: 42, Num2: 42, Num3: 42, Num4: 42, Num5: 42,
		})
	})
	t.Run("DynamicByValue", func(t *testing.T) {
		ResetRegistry()
		Register(DynamicByValue{}, DynamicV1{}, DynamicV2{}, DynamicV3{})
		test(t, DynamicByValue{
			Text1: "hello", Text2: "hello", Text3: "hello", Text4: "hello", Text5: "hello",
			Num1: 42, Num2: 42, Num3: 42, Num4: 42, Num5: 42,
		})
	})
	t.Run("DynamicOptimizedByValue", func(t *testing.T) {
		ResetRegistry()
		Register(DynamicByValue{}, DynamicV1{}, DynamicV2{}, DynamicOptimizedV3{})
		test(t, DynamicByValue{
			Text1: "hello", Text2: "hello", Text3: "hello", Text4: "hello", Text5: "hello",
			Num1: 42, Num2: 42, Num3: 42, Num4: 42, Num5: 42,
		})
	})
	t.Run("HardcodedByPointer", func(t *testing.T) {
		test(t, &HardcodedByPointer{
			Text1: "hello", Text2: "hello", Text3: "hello", Text4: "hello", Text5: "hello",
			Num1: 42, Num2: 42, Num3: 42, Num4: 42, Num5: 42,
		})
	})
	t.Run("DynamicByPointer", func(t *testing.T) {
		ResetRegistry()
		Register(DynamicByPointer{}, DynamicV1{}, DynamicV2{}, DynamicV3{})
		test(t, &DynamicByPointer{
			Text1: "hello", Text2: "hello", Text3: "hello", Text4: "hello", Text5: "hello",
			Num1: 42, Num2: 42, Num3: 42, Num4: 42, Num5: 42,
		})
	})
	t.Run("DynamicOptimizedByPointer", func(t *testing.T) {
		ResetRegistry()
		Register(DynamicByPointer{}, DynamicV1{}, DynamicV2{}, DynamicOptimizedV3{})
		test(t, &DynamicByPointer{
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

	b.Run("HardcodedByValue", func(b *testing.B) {
		bench(b, HardcodedByValue{
			Text1: "hello", Text2: "hello", Text3: "hello", Text4: "hello", Text5: "hello",
			Num1: 42, Num2: 42, Num3: 42, Num4: 42, Num5: 42,
		})
	})
	b.Run("DynamicByValue", func(b *testing.B) {
		ResetRegistry()
		Register(DynamicByValue{}, DynamicV1{}, DynamicV2{}, DynamicV3{})
		bench(b, DynamicByValue{
			Text1: "hello", Text2: "hello", Text3: "hello", Text4: "hello", Text5: "hello",
			Num1: 42, Num2: 42, Num3: 42, Num4: 42, Num5: 42,
		})
	})
	b.Run("DynamicOptimizedByValue", func(b *testing.B) {
		ResetRegistry()
		Register(DynamicByValue{}, DynamicV1{}, DynamicV2{}, DynamicOptimizedV3{})
		bench(b, DynamicByValue{
			Text1: "hello", Text2: "hello", Text3: "hello", Text4: "hello", Text5: "hello",
			Num1: 42, Num2: 42, Num3: 42, Num4: 42, Num5: 42,
		})
	})
	b.Run("HardcodedByPointer", func(b *testing.B) {
		bench(b, &HardcodedByPointer{
			Text1: "hello", Text2: "hello", Text3: "hello", Text4: "hello", Text5: "hello",
			Num1: 42, Num2: 42, Num3: 42, Num4: 42, Num5: 42,
		})
	})
	b.Run("DynamicByPointer", func(b *testing.B) {
		ResetRegistry()
		Register(DynamicByPointer{}, DynamicV1{}, DynamicV2{}, DynamicV3{})
		bench(b, &DynamicByPointer{
			Text1: "hello", Text2: "hello", Text3: "hello", Text4: "hello", Text5: "hello",
			Num1: 42, Num2: 42, Num3: 42, Num4: 42, Num5: 42,
		})
	})
	b.Run("DynamicOptimizedByPointer", func(b *testing.B) {
		ResetRegistry()
		Register(DynamicByPointer{}, DynamicV1{}, DynamicV2{}, DynamicOptimizedV3{})
		bench(b, &DynamicByPointer{
			Text1: "hello", Text2: "hello", Text3: "hello", Text4: "hello", Text5: "hello",
			Num1: 42, Num2: 42, Num3: 42, Num4: 42, Num5: 42,
		})
	})
}
