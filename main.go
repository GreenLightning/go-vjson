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

func resetRegistry() {
	entryByType = make(map[reflect.Type]entry)
}

// Register registers a type for serialization.
//
// The first parameter is the target type, while the following parameters
// correspond to individual version starting from v1, v2, etc. The concrete
// values passed to this function are ignored, only their types are considered.
//
// Register panics if an error is encountered.
// Register must not be called concurrently with any other call to Register, Marshal or Unmarshal.
// (Marshal and Unmarshal can be called concurrently with themselves.)
//
// (Register is intended to be only called from init functions, where the panic
// and concurrency limitations are not a concern.)
func Register(prototype interface{}, versionPrototypes ...interface{}) {
	err := registerError(prototype, versionPrototypes...)
	if err != nil {
		panic(err)
	}
}

func registerError(prototype interface{}, versionPrototypes ...interface{}) error {
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

	if _, ok := entryType.FieldByName("Version"); ok {
		return fmt.Errorf("type %v must not contain a field named Version, as it is reserved for vjson", entryType)
	}

	var entry entry
	entry.latestVersion = len(versionPrototypes)
	entry.versions = make(map[int]versionContext)

	seenTypes := make(map[reflect.Type]bool)
	seenTypes[entryType] = true

	var lastType reflect.Type
	for index, versionPrototype := range versionPrototypes {
		var context versionContext
		context.rtype = reflect.TypeOf(versionPrototype)

		if context.rtype.Kind() != reflect.Struct {
			return fmt.Errorf("only structs are allowed, but found %v for version %d", context.rtype, index+1)
		}

		if seenTypes[context.rtype] {
			return fmt.Errorf("struct %v for version %d was already passed earlier in the same call to register", context.rtype, index+1)
		}

		seenTypes[context.rtype] = true

		if lastType != nil {
			for i := 0; i < context.rtype.NumField(); i++ {
				dstField := context.rtype.Field(i)
				srcName := dstField.Name
				required := false
				if tag, ok := dstField.Tag.Lookup("vjson"); ok {
					if tag == "" {
						continue
					}
					srcName = tag
					required = true
				}
				srcField, ok := lastType.FieldByName(srcName)
				// ignore fields of embedded structs
				if ok && len(srcField.Index) != 1 {
					ok = false
				}
				if !ok {
					if required {
						return fmt.Errorf("field %s in %v has tag %s, but there is no such field in %v", dstField.Name, context.rtype, srcName, lastType)
					}
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

		if index+1 < len(versionPrototypes) {
			if _, ok := reflect.PtrTo(context.rtype).MethodByName("Pack"); ok {
				return fmt.Errorf("detected Pack method on %v, which is not the latest version", context.rtype)
			}
			if _, ok := reflect.PtrTo(context.rtype).MethodByName("Unpack"); ok {
				return fmt.Errorf("detected Unpack method on %v, which is not the latest version", context.rtype)
			}
		}

		entry.versions[index+1] = context
		lastType = context.rtype
	}

	entry.marshal.rtype = lastType
	if field, ok := lastType.FieldByName("Version"); ok {
		if len(field.Index) != 1 {
			return fmt.Errorf("Version field in %v must be a top-level field, but is in an embedded struct", lastType)
		}
		if field.Type.Kind() != reflect.Int {
			return fmt.Errorf("Version field in %v must have type int but is %v", lastType, field.Type)
		}
		entry.marshal.versionField = field.Index[0]
	} else {
		entry.marshal.versionField = -1
	}

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

// Marshal is like json.Marshal but adds a version number to the generated JSON.
// The type of the data passed to Marshal must have previously been registered
// with the vjson package or else an error is returned.
// Marshal always serializes to the latest known version.
func Marshal(v interface{}) ([]byte, error) {
	input := reflect.ValueOf(v)

	if input.Kind() == reflect.Ptr {
		input = input.Elem()
	}

	entry, ok := entryByType[input.Type()]
	if !ok {
		return nil, fmt.Errorf("vjson: type not registered: %v", input.Type())
	}

	value := reflect.New(entry.marshal.rtype)
	if entry.marshal.packFunc.IsValid() {
		var pointer reflect.Value
		if input.CanAddr() {
			pointer = input.Addr()
		} else {
			// Workaround for the case where input is not addressable,
			// but we have found a pack method, which expects a pointer.
			//
			// We want to allow this because json.Marshal() allows unaddressable
			// values as well and we don't want to make a special exception for
			// types that have a pack method, because then adding a pack method
			// could introduce errors at runtime.
			//
			// Therefore we have to take the hit and make an addressable copy
			// of input.
			//
			// See TestMarshalUnaddressableWithPack.
			pointer = reflect.New(input.Type())
			pointer.Elem().Set(input)
		}
		err := callErrorFunction(entry.marshal.packFunc, value, pointer)
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

// Unmarshal is like json.Unmarshal but respects the version number contained in the JSON.
// The type of the data passed to Unmarshal must have previously been registered with
// the vjson package and the version number contained in the JSON must be within the range
// of versions given to the Register function. Otherwise an error is returned.
// Unmarshal upgrades the data to the latest version.
func Unmarshal(data []byte, v interface{}) error {
	value := reflect.ValueOf(v)

	if kind := value.Kind(); kind != reflect.Ptr || value.IsNil() {
		if kind == reflect.Invalid {
			return fmt.Errorf("vjson: Unmarshal(nil)")
		}
		if kind != reflect.Ptr {
			return fmt.Errorf("vjson: Unmarshal(non-pointer %v)", value.Type())
		}
		return fmt.Errorf("vjson: Unmarshal(nil %v)", value.Type())
	}

	value = value.Elem()

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
