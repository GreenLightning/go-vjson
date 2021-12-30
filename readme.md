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
