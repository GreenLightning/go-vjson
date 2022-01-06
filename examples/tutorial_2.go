package main

import (
	"encoding/json"
	"fmt"

	"github.com/GreenLightning/go-vjson"
)

type Post struct {
	Author string
	Text   string
	Likes  int
}

func (p *Post) MarshalJSON() ([]byte, error) {
	return vjson.Marshal(p)
}

func (p *Post) UnmarshalJSON(data []byte) error {
	return vjson.Unmarshal(data, p)
}

func init() {
	vjson.Register(Post{}, PostV1{}, PostV2{})
}

type PostV1 struct {
	Author        string
	Text          string
	NumberOfLikes int
}

type PostV2 struct {
	Author string
	Text   string
	Likes  int `vjson:"NumberOfLikes"`
}

func main() {
	oldInput := []byte(`{ "Author": "Dolores", "Text": "Lorem ipsum dolor sit amet...", "NumberOfLikes": 99 }`)
	newInput := []byte(`{ "Version": 2, "Author": "Dolores", "Text": "Lorem ipsum dolor sit amet...", "Likes": 99 }`)

	var oldPost Post
	err := json.Unmarshal(oldInput, &oldPost)
	if err != nil {
		panic(err)
	}

	var newPost Post
	err = json.Unmarshal(newInput, &newPost)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Post: %+v\n", oldPost)
	fmt.Printf("Post: %+v\n", newPost)
}
