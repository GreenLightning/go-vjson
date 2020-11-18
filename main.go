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
	typ          reflect.Type
	mappings     []mapping
	versionField int
}

type unmarshal struct {
	typ reflect.Type
}

type entry struct {
	version             int
	marshal             marshal
	unmarshalsByVersion map[int]unmarshal
}

var entriesByType = make(map[reflect.Type]entry)

func ResetRegistry() {
	entriesByType = make(map[reflect.Type]entry)
}

func Register(prototype interface{}, versionPrototypes ...interface{}) {
	var entry entry
	entry.unmarshalsByVersion = make(map[int]unmarshal)

	var version int
	var typ reflect.Type
	for index, versionPrototype := range versionPrototypes {
		version = index + 1
		typ = reflect.TypeOf(versionPrototype)
		var unmarshal unmarshal
		unmarshal.typ = typ
		entry.unmarshalsByVersion[version] = unmarshal
	}

	entryType := reflect.TypeOf(prototype)

	entry.version = version
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
		elem.Field(entry.marshal.versionField).Set(reflect.ValueOf(entry.version))
		return json.Marshal(value.Interface())
	}

	data, err := json.Marshal(value.Interface())
	if err != nil {
		return nil, err
	}

	if string(data) == "{}" {
		result := fmt.Sprintf(`{"Version":%d}`, entry.version)
		return []byte(result), nil
	}

	var buffer bytes.Buffer
	fmt.Fprintf(&buffer, `{"Version":%d,`, entry.version)
	buffer.Write(data[1:])
	return buffer.Bytes(), nil
}

func Unmarshal(valueInterface interface{}, data []byte) error {
	value := reflect.ValueOf(valueInterface)

	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	}

	entry, ok := entriesByType[value.Type()]
	if !ok {
		return fmt.Errorf("vjson: type not registered: %v", value.Type())
	}

	version, err := unmarshalVersion(data)
	if err != nil {
		return err
	}

	unmarshal, ok := entry.unmarshalsByVersion[version]
	if !ok {
		return fmt.Errorf("vjson: unsupported version for %v: %d", value.Type(), version)
	}

	current := reflect.New(unmarshal.typ)
	err = json.Unmarshal(data, current.Interface())
	if err != nil {
		return err
	}

	elem := current.Elem()
	for _, mapping := range entry.marshal.mappings {
		value.Field(mapping.src).Set(elem.Field(mapping.dst))
	}

	return nil
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
