package vjson

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
)

type mapping struct {
	src int
	dst int
}

type marshalContext struct {
	rtype        reflect.Type
	packFunc     reflect.Value
	mappings     []mapping
	versionField int
}

type unmarshalContext struct {
	unpackFunc reflect.Value
	mappings   []mapping
}

type versionContext struct {
	rtype       reflect.Type
	mappings    []mapping
	upgradeFunc reflect.Value
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
	err := RegisterError(prototype, versionPrototypes...)
	if err != nil {
		panic(err)
	}
}

func RegisterError(prototype interface{}, versionPrototypes ...interface{}) error {
	entryType := reflect.TypeOf(prototype)

	if entryType.Kind() != reflect.Struct {
		return fmt.Errorf("only structs are allowed, but found %v", entryType)
	}

	if _, ok := entryByType[entryType]; ok {
		return fmt.Errorf("type %v already registered", entryType)
	}

	if len(versionPrototypes) == 0 {
		return fmt.Errorf("must provide at least one version prototype")
	}

	var entry entry
	entry.latestVersion = len(versionPrototypes)
	entry.versions = make(map[int]versionContext)

	var lastType reflect.Type
	for index, versionPrototype := range versionPrototypes {
		var context versionContext
		context.rtype = reflect.TypeOf(versionPrototype)

		if context.rtype.Kind() != reflect.Struct {
			return fmt.Errorf("only structs are allowed, but found %v for version %d", context.rtype, index+1)
		}

		if lastType != nil {
			for i := 0; i < context.rtype.NumField(); i++ {
				dstField := context.rtype.Field(i)
				srcName := dstField.Name
				if tag, ok := dstField.Tag.Lookup("vjson"); ok {
					if tag == "" {
						continue
					}
					srcName = tag
				}
				srcField, ok := lastType.FieldByName(srcName)
				if !ok {
					continue
				}
				if srcField.Type != dstField.Type {
					if srcField.Name != dstField.Name {
						return fmt.Errorf("cannot copy field %s (%v) in %v to field %s (%v) in %v because they have different types", srcField.Name, srcField.Type, lastType, dstField.Name, dstField.Type, context.rtype)
					}
					return fmt.Errorf("field %s has different types in %v (%v) and %v (%v)", srcField.Name, lastType, srcField.Type, context.rtype, dstField.Type)
				}
				mapping := mapping{src: srcField.Index[0], dst: dstField.Index[0]}
				context.mappings = append(context.mappings, mapping)
			}
			sort.Slice(context.mappings, func(i, j int) bool { return context.mappings[i].src < context.mappings[j].src })
		}

		// The upgrade method must have a pointer receiver,
		// because it is meant to modify the receiver.
		if upgradeMethod, ok := reflect.PtrTo(context.rtype).MethodByName("Upgrade"); ok {
			context.upgradeFunc = upgradeMethod.Func
		}

		entry.versions[index+1] = context
		lastType = context.rtype
	}

	entry.marshal.rtype = lastType
	if packMethod, ok := reflect.PtrTo(lastType).MethodByName("Pack"); ok {
		entry.marshal.packFunc = packMethod.Func
	} else {
		for i := 0; i < entryType.NumField(); i++ {
			srcField := entryType.Field(i)
			dstField, ok := lastType.FieldByName(srcField.Name)
			if !ok {
				continue
			}
			if srcField.Type != dstField.Type {
				return fmt.Errorf("field %s has different types in %v (%v) and %v (%v)", srcField.Name, entryType, srcField.Type, lastType, dstField.Type)
			}
			mapping := mapping{src: srcField.Index[0], dst: dstField.Index[0]}
			entry.marshal.mappings = append(entry.marshal.mappings, mapping)
		}
	}

	if field, ok := lastType.FieldByName("Version"); ok {
		entry.marshal.versionField = field.Index[0]
	} else {
		entry.marshal.versionField = -1
	}

	if unpackMethod, ok := reflect.PtrTo(lastType).MethodByName("Unpack"); ok {
		entry.unmarshal.unpackFunc = unpackMethod.Func
	} else {
		for i := 0; i < lastType.NumField(); i++ {
			srcField := lastType.Field(i)
			dstField, ok := entryType.FieldByName(srcField.Name)
			if !ok {
				continue
			}
			if srcField.Type != dstField.Type {
				return fmt.Errorf("field %s has different types in %v (%v) and %v (%v)", srcField.Name, entryType, dstField.Type, lastType, srcField.Type)
			}
			mapping := mapping{src: srcField.Index[0], dst: dstField.Index[0]}
			entry.unmarshal.mappings = append(entry.unmarshal.mappings, mapping)
		}
	}

	entryByType[entryType] = entry
	return nil
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
	if entry.marshal.packFunc.IsValid() {
		err := callErrorFunction(entry.marshal.packFunc, value, input.Addr())
		if err != nil {
			return nil, err
		}
	} else {
		copyFields(input, value.Elem(), entry.marshal.mappings)
	}

	if entry.marshal.versionField >= 0 {
		value.Elem().Field(entry.marshal.versionField).Set(reflect.ValueOf(entry.latestVersion))
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

	if string(data) == "null" {
		return nil
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
		if nextContext.upgradeFunc.IsValid() {
			err := callErrorFunction(nextContext.upgradeFunc, next, current)
			if err != nil {
				return err
			}
		}
		currentContext = nextContext
		current = next
	}

	if entry.unmarshal.unpackFunc.IsValid() {
		err := callErrorFunction(entry.unmarshal.unpackFunc, current, value.Addr())
		if err != nil {
			return err
		}
	} else {
		copyFields(current.Elem(), value, entry.unmarshal.mappings)
	}
	return nil
}

func copyFields(src, dst reflect.Value, mappings []mapping) {
	for _, mapping := range mappings {
		dst.Field(mapping.dst).Set(src.Field(mapping.src))
	}
}

func callErrorFunction(f reflect.Value, params ...reflect.Value) error {
	returnValues := f.Call(params)
	if len(returnValues) == 0 {
		return nil
	}
	errorValue := returnValues[len(returnValues)-1]
	errorInterface := errorValue.Interface()
	if errorInterface == nil {
		return nil
	}
	return errorInterface.(error)
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
