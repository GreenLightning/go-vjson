ALPHA SOFTWARE
==============

**This package is still in alpha. It has no documentation and will most likely panic if you don't know what you're doing.**

# Introduction

This package adds versioning on top of the standard library's json package for
building backward-compatible formats.

Each struct is versioned independently using an integer version number, which is
added to the serialized JSON under the `"Version"` key. In the code, the data
format of each version is explicitly defined in a separate struct declaration.
Advanced features include automatic renaming using tags and running arbitrary
code during version upgrades.

For example (`examples/introduction.go`):

```
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
	return vjson.Unmarshal(u, data)
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
- Finally, the serialization speed of the library would match that of
  `encoding/json`, since the additional copy to add the version can be avoided.

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
types. In particular unmarshaling into an empty `interface{}` creates a
`map[string]interface{}`, while unmarshaling into any other interface returns an
error. Unfortunately, it is difficult to change this behavior. Therefore,
implementing this feature at the library level most likely requires forking the
`encoding/json` package as described above.

An example where this would be useful:

```
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
