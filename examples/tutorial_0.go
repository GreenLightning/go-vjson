package main

import (
	"encoding/json"
	"fmt"
)

type Post struct {
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

	fmt.Printf("Post: %+v\n", post)
}
