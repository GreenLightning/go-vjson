# Introduction

This package adds versioning on top of the standard library's json package for
building backward-compatible formats.

Each struct is versioned independently using an integer version number, which is
added to the serialized JSON under the `"Version"` key. In the code, the data
format of each version is explicitly defined in a separate struct declaration.
Advanced features include automatic renaming using tags and running custom
functions during version upgrades.

For example (`examples/introduction.go`):

```go
// The struct used by the rest of the application:
type User struct {
	ID          string // in hex
	UserName    string // for @mentions
	DisplayName string // might contain spaces, etc.
}

// For best compatibility with encoding/json, we recommend defining these two methods:
func (u *User) MarshalJSON() ([]byte, error) {
	return vjson.Marshal(u)
}

func (u *User) UnmarshalJSON(data []byte) error {
	return vjson.Unmarshal(data, u)
}

// The individual versions have to be registered:
func init() {
	vjson.Register(User{}, UserV1{}, UserV2{}, UserV3{})
}

type UserV1 struct {
	ID   int
	Name string
}

// This version distinguishes between UserName and DisplayName.
// Both are initialized with the value of Name from V1.
type UserV2 struct {
	ID          int
	UserName    string `vjson:"Name"`
	DisplayName string `vjson:"Name"`
}

// This version switches to storing the user ID as a string.
type UserV3 struct {
	// Specifying an empty tag prevents the value from being copied from the previous version,
	// which would not work in this case, because the type has changed.
	ID          string `vjson:""`
	UserName    string
	DisplayName string
}

// This function is run during the upgrade and converts the ID.
func (v3 *UserV3) Upgrade(v2 *UserV2) {
	v3.ID = fmt.Sprintf("%04x", v2.ID)
}

func main() {
	// A missing version key implies version 1.
	input := []byte(`{ "ID": 42, "Name": "dale_cooper" }`)

	var user User
	err := json.Unmarshal(input, &user)
	if err != nil {
		panic(err)
	}

	output, err := json.Marshal(&user)
	if err != nil {
		panic(err)
	}

	fmt.Printf("User: %+v\n", user)
	fmt.Printf("Output: %s\n", output)
	// User: {ID:002a UserName:dale_cooper DisplayName:dale_cooper}
	// Output: {"Version":3,"ID":"002a","UserName":"dale_cooper","DisplayName":"dale_cooper"}
}
```

# Tutorial

`vjson` can be added to a working program when required. For example, given a
 social media application that stores posts in JSON:

```go
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
```

Assuming you want to rename `NumberOfLikes` to just `Likes`, first copy `Post`
to `PostV1` to preserve the data format of the first version (you can also
choose a different name) and add an `init` function to register this version.

```go
type Post struct {
	Author        string
	Text          string
	NumberOfLikes int
}

func init() {
	vjson.Register(Post{}, PostV1{})
}

type PostV1 struct {
	Author        string
	Text          string
	NumberOfLikes int
}
```

To be compatible with `encoding/json` we also add some boilerplate functions
that forward to `vjson`:

```go
func (p *Post) MarshalJSON() ([]byte, error) {
	return vjson.Marshal(p)
}

func (p *Post) UnmarshalJSON(data []byte) error {
	return vjson.Unmarshal(data, p)
}
```

This iteration of the program is now using `vjson` but otherwise almost
identical to the previous iteration. The only slight difference is that if you
call `json.Marshal(&post)`, the output will contain a version number:
`{"Version":1,"Author":"Dolores","Text":"Lorem ipsum dolor sit
amet...","NumberOfLikes":99}`.

Now, to actually change the name of the field, create a new version struct with
the new name (don't forget to register it). The tag on the `Likes` field tells
`vjson` that this field was renamed, so that the value of the old field is
copied over:

```go
type Post struct {
	Author string
	Text   string
	Likes  int
}

func init() {
	vjson.Register(Post{}, PostV1{}, PostV2{}) // added PostV2{}
}

type PostV1 struct { ... } // unchanged

type PostV2 struct {
	Author string
	Text   string
	Likes  int `vjson:"NumberOfLikes"`
}
```

That's it. This iteration of the program now accepts both of these JSON objects:

```
{ "Author": "Dolores", "Text": "Lorem ipsum dolor sit amet...", "NumberOfLikes": 99 }
{ "Version": 2, "Author": "Dolores", "Text": "Lorem ipsum dolor sit amet...", "Likes": 99 }
```

If the `Version` key is missing, `vjson` assumes version 1 for backward
compatibility, so output from the original program, which did not use `vjson`,
is accepted as well.

# Features

Marshaling always produces the latest version. During unmarshaling the version
is read from the data and the appropriate version struct is used. If the data is
not at the latest version, the version struct is upgraded to the next version
until it is the latest version.

Upgrading involves copying over fields with the same name from the older
version. Tags can be used to specify the name of different field whose value
should be copied into the tagged field. This is useful for renaming fields or
using the value of a different field as the default value for a field (see
examples in the tutorial and introduction). For the copying to work, the types
of the fields must match. `Register` will panic if this is not the case. Use an
empty tag to disable copying even if a field with the same name exists. Tags for
the `encoding/json` package can be used on the version structs. They are ignored
by the vjson package.

Additionally, an optional `Upgrade` method can be defined on a version struct
taking as an argument a pointer to the previous version (again, see introduction
for an example). This function is called for upgrading after the fields have
been copied and can contain custom upgrade logic.

The latest version struct can define optional `Pack` and `Unpack` methods to
convert between the general-use struct and the version struct. If these methods
are not defined, conversion is performed by copying fields of the same name
(ignoring tags). Unlike upgrading, if one of these methods is defined, no
copying is performed for the corresponding conversion. Example (see `examples\pack.go`):

```go
type Example struct {
	Value int
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
```

The `Upgrade`, `Pack` and `Unpack` methods may optionally have a return value of
type `error`.

To slightly improve serialization speed the latest version struct should have a
`Version int` field, which is automatically used by the library to add the
version number to the generated JSON. If not present, the generated JSON has to
be copied to add the version number.

# Limitations

The model of this package is that each type is versioned independently. This
obviously results in a much looser coupling between types than say a single
version number for an entire file, which might contain structs of many different
types. In general, this has worked quite well for me, but it is something to be
aware of. In particular, updating multiple structs together, where the update
logic of one struct depends on the data of another struct might get a little
more complicated.

The standard library's `encoding/json` package does not provide any mechanism to
store context for a particular operation (e.g. a `context.Context` as part of
each `json.Encoder` and passed to `MarshalJSON`). Because of this, the registry
of versioned types has to be stored in a global variable and we cannot provide
individual encoders with different registries. A context for `json.Unmarshal`
would also be useful for global version numbers, as the detected version could
be stored in the context.

The library depends on the `MarshalJSON`/`UnmarshalJSON` methods for interfacing
with `encoding/json`. However, there are some tricky edge cases where these
methods are not called because they have a pointer receiver instead of a value
receiver:

```go
package main

import (
	"encoding/json"
	"fmt"
)

type A struct{}

// Change this to a value receiver (func (a A) ...) to make it work
// in all four test cases below.
func (a *A) MarshalJSON() ([]byte, error) {
	fmt.Print("A.MarshalJSON")
	return []byte("null"), nil
}

type ParentA struct{
	A A
}

func main() {
	fmt.Print("A by value: ")
	json.Marshal(A{}) // Does not call MarshalJSON!
	fmt.Println()

	fmt.Print("A by pointer: ")
	json.Marshal(&A{}) // Ok.
	fmt.Println()

	fmt.Print("ParentA by value: ")
	json.Marshal(ParentA{}) // Does not call MarshalJSON!
	fmt.Println()

	fmt.Print("ParentA by pointer: ")
	json.Marshal(&ParentA{}) // Ok.
	fmt.Println()
}
```

Unfortunately, in these cases the structs would be serialized using the regular
`encoding/json` approach, thus most likely producing valid JSON but without a
version number and ignoring the conversions done by `vjson` (`Pack` methods).

To avoid these corner cases, follow these two rules:

- If a struct is placed in another struct by value, make its `MarshalJSON()`
  method have a value receiver.
- Always make sure that you pass a pointer to `json.Marshal()` (unless the
  struct already has a `MarshalJSON()` method with a value receiver).

# Future Work

This section contains some ideas for future improvements.
There currently is no timeline for their implementation.

### Fork `encoding/json` package

The idea is to copy the `encoding/json` package and integrate the versioning
features directly into the code instead of the current approach of this package
calling `encoding/json` to do most of the work.

This would have several advantages:

- `vjson` could become a drop-in replacement for `encoding/json`.
- User-defined structs would no longer have to implement
  `MarshalJSON()` / `UnmarshalJSON()` to forward to `vjson`.
- The registration mechanism could be changed, so that instead of calling a
  `Register()` function, user-defined structs implement a specific interface
  method that returns the different possible versions.
- Alternatively, it would be possible to build encoders/decoders with individual
  registries, instead of the global registry system.
- Finally, the serialization speed of the library could be improved. Currently,
  there is some overhead from copying the data multiple times. One copy is used
  to insert the version number (which can be avoided by adding a `Version` field
  to the struct for the latest version). A second copy is necessary to copy the
  result from `MarshalJSON` into the internal output buffer of the `json`
  package. I think both of these could be avoided by integrating the logic
  directly into the encoder.

A disadvantage would be that changes to `encoding/json` would have to be merged
regularly.

### Support Polymorphic Types

By polymorphic types I mean interface types that hold values of several
different implementing types. During serialization the actual type has to be
encoded into the output. During deserialization the type has to be parsed and a
value of the appropriate type created.

I propose to embed the type information into the JSON data using a `"Type"` key,
similar to the `"Version"` key. This feature further requires a mapping between
Go types and string identifiers, necessitating another registration function.

The standard library's json package does not have any support for polymorphic
types. In particular unmarshaling a JSON object into an empty `interface{}`
creates a `map[string]interface{}`, while unmarshaling into any other interface
returns an error. Unfortunately, it is difficult to change this behavior.
Therefore, implementing this feature at the library level most likely requires
forking the `encoding/json` package as described above.

An example where this would be useful:

```go
type Node interface{}

type Parent struct {
	Children []Node
}

type Leaf struct {
	Value int
}
```

And the JSON for a tree:

```
{
	"Type": "parent",
	"Children": [
		{
			"Type": "parent",
			"Children": [
				{
					"Type": "leaf",
					"Value": 9
				},
				{
					"Type": "leaf",
					"Value": 16
				}
			]
		},
		{
			"Type": "leaf",
			"Value": 25
		}
	]
}
```
