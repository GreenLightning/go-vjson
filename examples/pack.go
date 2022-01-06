package main

import (
	"encoding/json"
	"fmt"

	"github.com/GreenLightning/go-vjson"
)

type Example struct {
	Value int
}

func (e *Example) MarshalJSON() ([]byte, error) {
	return vjson.Marshal(e)
}

func (e *Example) UnmarshalJSON(data []byte) error {
	return vjson.Unmarshal(data, e)
}

func init() {
	vjson.Register(Example{}, ExampleV1{})
}

type ExampleV1 struct {
	Value string
}

func (latest *ExampleV1) Pack(example *Example) {
	latest.Value = fmt.Sprintf("%x", example.Value)
}

func (latest *ExampleV1) Unpack(example *Example) error {
	_, err := fmt.Sscanf(latest.Value, "%x", &example.Value)
	return err
}

func main() {
	input := []byte(`{ "Value": "42" }`)

	var example Example
	err := json.Unmarshal(input, &example)
	if err != nil {
		panic(err)
	}

	output, err := json.Marshal(&example)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Data: %+v\n", example)
	fmt.Printf("Output: %s\n", output)
}
