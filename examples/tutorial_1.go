package main

import (
	"encoding/json"
	"fmt"

	"github.com/GreenLightning/go-vjson"
)

type Post struct {
	Author        string
	Text          string
	NumberOfLikes int
}

func (p *Post) MarshalJSON() ([]byte, error) {
	return vjson.Marshal(p)
}

func (p *Post) UnmarshalJSON(data []byte) error {
	return vjson.Unmarshal(data, p)
}

func init() {
	vjson.Register(Post{}, PostV1{})
}

type PostV1 struct {
	Author        string
	Text          string
	NumberOfLikes int
}

func main() {
	input := []byte(`{ "Author": "Dolores", "Text": "Lorem ipsum dolor sit amet...", "NumberOfLikes": 99 }`)

	var post Post
	err := json.Unmarshal(input, &post)
	if err != nil {
		panic(err)
	}

	output, err := json.Marshal(&post)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Post: %+v\n", post)
	fmt.Printf("Output: %s\n", output)
}
