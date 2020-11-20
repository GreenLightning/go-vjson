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

type marshalContext struct {
	rtype        reflect.Type
	mappings     []mapping
	versionField int
}

type unmarshalContext struct {
	mappings []mapping
}

type versionContext struct {
	rtype    reflect.Type
	mappings []mapping
}

type entry struct {
	latestVersion int
	versions      map[int]versionContext
	marshal       marshalContext
	unmarshal     unmarshalContext
}

var entryByType = make(map[reflect.Type]entry)

func ResetRegistry() {
	entryByType = make(map[reflect.Type]entry)
}

func Register(prototype interface{}, versionPrototypes ...interface{}) {
	entryType := reflect.TypeOf(prototype)

	var entry entry
	entry.latestVersion = len(versionPrototypes)
	entry.versions = make(map[int]versionContext)

	var lastType reflect.Type
	for index, versionPrototype := range versionPrototypes {
		var context versionContext
		context.rtype = reflect.TypeOf(versionPrototype)

		if lastType != nil {
			for i := 0; i < lastType.NumField(); i++ {
				srcField := lastType.Field(i)
				dstField, ok := context.rtype.FieldByName(srcField.Name)
				if ok {
					mapping := mapping{src: srcField.Index[0], dst: dstField.Index[0]}
					context.mappings = append(context.mappings, mapping)
				}
			}
		}

		entry.versions[index+1] = context
		lastType = context.rtype
	}

	entry.marshal.rtype = lastType
	for i := 0; i < entryType.NumField(); i++ {
		srcField := entryType.Field(i)
		dstField, ok := lastType.FieldByName(srcField.Name)
		if ok {
			mapping := mapping{src: srcField.Index[0], dst: dstField.Index[0]}
			entry.marshal.mappings = append(entry.marshal.mappings, mapping)
		}
	}

	if field, ok := lastType.FieldByName("Version"); ok {
		entry.marshal.versionField = field.Index[0]
	} else {
		entry.marshal.versionField = -1
	}

	for i := 0; i < lastType.NumField(); i++ {
		srcField := lastType.Field(i)
		dstField, ok := entryType.FieldByName(srcField.Name)
		if ok {
			mapping := mapping{src: srcField.Index[0], dst: dstField.Index[0]}
			entry.unmarshal.mappings = append(entry.unmarshal.mappings, mapping)
		}
	}

	entryByType[entryType] = entry
}

func Marshal(inputInterface interface{}) ([]byte, error) {
	input := reflect.ValueOf(inputInterface)

	if input.Kind() == reflect.Ptr {
		input = input.Elem()
	}

	entry, ok := entryByType[input.Type()]
	if !ok {
		return nil, fmt.Errorf("vjson: type not registered: %v", input.Type())
	}

	value := reflect.New(entry.marshal.rtype)
	elem := value.Elem()
	copyFields(input, elem, entry.marshal.mappings)

	if entry.marshal.versionField >= 0 {
		elem.Field(entry.marshal.versionField).Set(reflect.ValueOf(entry.latestVersion))
		return json.Marshal(value.Interface())
	}

	data, err := json.Marshal(value.Interface())
	if err != nil {
		return nil, err
	}

	if string(data) == "{}" {
		result := fmt.Sprintf(`{"Version":%d}`, entry.latestVersion)
		return []byte(result), nil
	}

	var buffer bytes.Buffer
	fmt.Fprintf(&buffer, `{"Version":%d,`, entry.latestVersion)
	buffer.Write(data[1:])
	return buffer.Bytes(), nil
}

func Unmarshal(valueInterface interface{}, data []byte) error {
	value := reflect.ValueOf(valueInterface)

	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	}

	entry, ok := entryByType[value.Type()]
	if !ok {
		return fmt.Errorf("vjson: type not registered: %v", value.Type())
	}

	version, err := unmarshalVersion(data)
	if err != nil {
		return err
	}

	currentContext, ok := entry.versions[version]
	if !ok {
		return fmt.Errorf("vjson: unsupported version for %v: %d", value.Type(), version)
	}

	current := reflect.New(currentContext.rtype)
	err = json.Unmarshal(data, current.Interface())
	if err != nil {
		return err
	}

	for version < entry.latestVersion {
		version++
		nextContext := entry.versions[version]
		next := reflect.New(nextContext.rtype)
		copyFields(current.Elem(), next.Elem(), nextContext.mappings)
		currentContext = nextContext
		current = next
	}

	copyFields(current.Elem(), value, entry.unmarshal.mappings)
	return nil
}

func copyFields(src, dst reflect.Value, mappings []mapping) {
	for _, mapping := range mappings {
		dst.Field(mapping.dst).Set(src.Field(mapping.src))
	}
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
