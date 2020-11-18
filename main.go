package vjson

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
)

type mapping struct {
	src int
	dst int
}

type marshal struct {
	version      int
	typ          reflect.Type
	mappings     []mapping
	versionField int
}

type entry struct {
	typesByVersion map[int]reflect.Type
	marshal        marshal
}

var entriesByType = make(map[reflect.Type]entry)

func ResetRegistry() {
	entriesByType = make(map[reflect.Type]entry)
}

func Register(prototype interface{}, versionPrototypes ...interface{}) {
	var entry entry
	entry.typesByVersion = make(map[int]reflect.Type)

	var version int
	var typ reflect.Type
	for index, versionPrototype := range versionPrototypes {
		version = index + 1
		typ = reflect.TypeOf(versionPrototype)
		entry.typesByVersion[version] = typ
	}

	entryType := reflect.TypeOf(prototype)

	entry.marshal.version = version
	entry.marshal.typ = typ
	for i := 0; i < entryType.NumField(); i++ {
		srcField := entryType.Field(i)
		dstField, ok := typ.FieldByName(srcField.Name)
		if ok {
			mapping := mapping{src: srcField.Index[0], dst: dstField.Index[0]}
			entry.marshal.mappings = append(entry.marshal.mappings, mapping)
		}
	}

	if field, ok := typ.FieldByName("Version"); ok {
		entry.marshal.versionField = field.Index[0]
	} else {
		entry.marshal.versionField = -1
	}

	entriesByType[entryType] = entry
}

func Marshal(inputInterface interface{}) ([]byte, error) {
	input := reflect.ValueOf(inputInterface)

	if input.Kind() == reflect.Ptr {
		input = input.Elem()
	}

	entry, ok := entriesByType[input.Type()]
	if !ok {
		return nil, fmt.Errorf("vjson: type not registered: %v", input.Type())
	}

	value := reflect.New(entry.marshal.typ)
	elem := value.Elem()
	for _, mapping := range entry.marshal.mappings {
		elem.Field(mapping.dst).Set(input.Field(mapping.src))
	}

	if entry.marshal.versionField >= 0 {
		elem.Field(entry.marshal.versionField).Set(reflect.ValueOf(entry.marshal.version))
		return json.Marshal(value.Interface())
	}

	data, err := json.Marshal(value.Interface())
	if err != nil {
		return nil, err
	}

	if string(data) == "{}" {
		result := fmt.Sprintf(`{"Version":%d}`, entry.marshal.version)
		return []byte(result), nil
	}

	var buffer bytes.Buffer
	fmt.Fprintf(&buffer, `{"Version":%d,`, entry.marshal.version)
	buffer.Write(data[1:])
	return buffer.Bytes(), nil
}

type versionContainer struct {
	Version int
}

func unmarshalVersion(data []byte) (int, error) {
	var container versionContainer
	err := json.Unmarshal(data, &container)
	if err != nil {
		return 0, err
	}
	if container.Version < 0 {
		return 0, fmt.Errorf("vjson: cannot unmarshal object: negative version number")
	}
	if container.Version == 0 {
		// If the version field is omitted, version 1 is implied.
		container.Version = 1
	}
	return container.Version, nil
}
