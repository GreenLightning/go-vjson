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
	Num1, Num2, Num3, Num4, ExtraNum      int
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
		v3.Text5 = fmt.Sprintf("Extra: %s", v2.ExtraText)
		v3.Num1 = v2.Num1
		v3.Num2 = v2.Num2
		v3.Num3 = v2.Num3
		v3.Num4 = v2.Num4
		v3.Num5 = v2.ExtraNum
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
	Num1, Num2, Num3, Num4, ExtraNum      int
}

type DynamicV3 struct {
	Text1, Text2, Text3, Text4, Text5 string
	Num1, Num2, Num3, Num4            int
	Num5                              int `vjson:"ExtraNum"`
}

func (dynamic *DynamicV3) Upgrade(old *DynamicV2) {
	dynamic.Text5 = fmt.Sprintf("Extra: %s", old.ExtraText)
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

func (dynamic *Dynamic) UnmarshalJSON(data []byte) error {
	return Unmarshal(data, dynamic)
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
		resetRegistry()
		Register(DynamicByValue{}, DynamicV1{}, DynamicV2{}, DynamicV3{})

		test(t, DynamicByValue{
			Text1: "hello", Text2: "hello", Text3: "hello", Text4: "hello", Text5: "hello",
			Num1: 42, Num2: 42, Num3: 42, Num4: 42, Num5: 42,
		})
	})
	t.Run("DynamicOptimizedByValue", func(t *testing.T) {
		resetRegistry()
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
		resetRegistry()
		Register(DynamicByPointer{}, DynamicV1{}, DynamicV2{}, DynamicV3{})

		test(t, &DynamicByPointer{
			Text1: "hello", Text2: "hello", Text3: "hello", Text4: "hello", Text5: "hello",
			Num1: 42, Num2: 42, Num3: 42, Num4: 42, Num5: 42,
		})
	})
	t.Run("DynamicOptimizedByPointer", func(t *testing.T) {
		resetRegistry()
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
		resetRegistry()
		Register(DynamicByValue{}, DynamicV1{}, DynamicV2{}, DynamicV3{})
		b.ResetTimer()

		bench(b, DynamicByValue{
			Text1: "hello", Text2: "hello", Text3: "hello", Text4: "hello", Text5: "hello",
			Num1: 42, Num2: 42, Num3: 42, Num4: 42, Num5: 42,
		})
	})
	b.Run("DynamicOptimizedByValue", func(b *testing.B) {
		resetRegistry()
		Register(DynamicByValue{}, DynamicV1{}, DynamicV2{}, DynamicOptimizedV3{})
		b.ResetTimer()

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
		resetRegistry()
		Register(DynamicByPointer{}, DynamicV1{}, DynamicV2{}, DynamicV3{})
		b.ResetTimer()

		bench(b, &DynamicByPointer{
			Text1: "hello", Text2: "hello", Text3: "hello", Text4: "hello", Text5: "hello",
			Num1: 42, Num2: 42, Num3: 42, Num4: 42, Num5: 42,
		})
	})
	b.Run("DynamicOptimizedByPointer", func(b *testing.B) {
		resetRegistry()
		Register(DynamicByPointer{}, DynamicV1{}, DynamicV2{}, DynamicOptimizedV3{})
		b.ResetTimer()

		bench(b, &DynamicByPointer{
			Text1: "hello", Text2: "hello", Text3: "hello", Text4: "hello", Text5: "hello",
			Num1: 42, Num2: 42, Num3: 42, Num4: 42, Num5: 42,
		})
	})
}

func TestUnmarshal(t *testing.T) {
	data := []byte(`{"Version":2,"Text1":"hello","Text2":"hello","Text3":"hello","Text4":"hello","ExtraText":"extra","Num1":42,"Num2":42,"Num3":42,"Num4":42,"ExtraNum":42}`)

	unmarshal := func(t *testing.T, value interface{}) {
		err := json.Unmarshal(data, value)
		if err != nil {
			t.Fatal("unexpected err:", err)
		}
	}

	check := func(t *testing.T, text4, text5 string, num4, num5 int) {
		if text4 != "hello" {
			t.Error("wrong Text4:", text4)
		}
		if text5 != "Extra: extra" {
			t.Error("wrong Text5:", text5)
		}
		if num4 != 42 {
			t.Error("wrong Num4:", num4)
		}
		if num5 != 42 {
			t.Error("wrong Num5:", num5)
		}
	}

	t.Run("Hardcoded", func(t *testing.T) {
		var value Hardcoded
		unmarshal(t, &value)
		check(t, value.Text4, value.Text5, value.Num4, value.Num5)
	})
	t.Run("Dynamic", func(t *testing.T) {
		resetRegistry()
		Register(Dynamic{}, DynamicV1{}, DynamicV2{}, DynamicV3{})

		var value Dynamic
		unmarshal(t, &value)
		check(t, value.Text4, value.Text5, value.Num4, value.Num5)
	})
}

func BenchmarkUnmarshal(b *testing.B) {
	data := []byte(`{"Version":2,"Text1":"hello","Text2":"hello","Text3":"hello","Text4":"hello","ExtraText":"extra","Num1":42,"Num2":42,"Num3":42,"Num4":42,"ExtraNum":42}`)

	bench := func(b *testing.B, value interface{}) {
		for i := 0; i < b.N; i++ {
			err := json.Unmarshal(data, value)
			if err != nil {
				b.Fatal("unexpected err:", err)
			}
		}
	}

	b.Run("Hardcoded", func(b *testing.B) {
		var value Hardcoded
		bench(b, &value)
	})
	b.Run("Dynamic", func(b *testing.B) {
		resetRegistry()
		Register(Dynamic{}, DynamicV1{}, DynamicV2{}, DynamicV3{})
		b.ResetTimer()

		var value Dynamic
		bench(b, &value)
	})
}

type Ordered struct {
	A, B, C, D, E, F, G, H, I, J string
}

type OrderedV1 struct {
	A, B, C, D, E, F, G, H, I, J string
}

type OrderedV2 struct {
	A, B, C, D, E, F, G, H, I, J string
}

type OrderedV3 struct {
	A, B, C, D, E, F, G, H, I, J string
}

func (value *Ordered) UnmarshalJSON(data []byte) error {
	return Unmarshal(data, value)
}

type Unordered struct {
	A, B, C, D, E, F, G, H, I, J string
}

type UnorderedV1 struct {
	C, D, F, E, B, G, J, I, A, H string
}

type UnorderedV2 struct {
	E, G, D, B, I, J, A, F, H, C string
}

type UnorderedV3 struct {
	A, C, H, E, F, B, J, I, D, G string
}

func (value *Unordered) UnmarshalJSON(data []byte) error {
	return Unmarshal(data, value)
}

func BenchmarkUnmarshalSorting(b *testing.B) {
	data := []byte(`{"Version":1,"A":"aaaaa","B":"bbbbb","C":"ccccc","D":"ddddd","E":"eeeee","F":"fffff","G":"ggggg","H":"hhhhh","I":"iiiii","J":"jjjjj"}`)

	bench := func(b *testing.B, value interface{}) {
		for i := 0; i < b.N; i++ {
			err := json.Unmarshal(data, value)
			if err != nil {
				b.Fatal("unexpected err:", err)
			}
		}
	}

	b.Run("Ordered", func(b *testing.B) {
		resetRegistry()
		Register(Ordered{}, OrderedV1{}, OrderedV2{}, OrderedV3{})
		b.ResetTimer()

		var value Ordered
		bench(b, &value)
	})
	b.Run("Unordered", func(b *testing.B) {
		resetRegistry()
		Register(Unordered{}, UnorderedV1{}, UnorderedV2{}, UnorderedV3{})
		b.ResetTimer()

		var value Unordered
		bench(b, &value)
	})
}
