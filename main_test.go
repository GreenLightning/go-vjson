package vjson

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
)

type Simple struct {
	Text   string
	Number int
}

func (s Simple) MarshalJSON() ([]byte, error) {
	return Marshal(s)
}

func (s *Simple) UnmarshalJSON(data []byte) error {
	return Unmarshal(data, s)
}

type SimpleV1 struct {
	Text   string
	Number int
}

func TestMarshalNil(t *testing.T) {
	resetRegistry()
	Register(Simple{}, SimpleV1{})

	data, err := json.Marshal(nil)
	if err != nil {
		t.Fatal("unexpected err:", err)
	}
	if string(data) != "null" {
		t.Error("wrong data:", string(data))
	}
}

func TestMarshalNilValue(t *testing.T) {
	resetRegistry()
	Register(Simple{}, SimpleV1{})

	var value *Simple

	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal("unexpected err:", err)
	}
	if string(data) != "null" {
		t.Error("wrong data:", string(data))
	}
}

func TestMarshalNotRegistered(t *testing.T) {
	resetRegistry()

	value := Simple{Text: "hello", Number: 42}

	_, err := json.Marshal(value)
	if err == nil {
		t.Fatal("missing error")
	}
	if !strings.Contains(err.Error(), "registered") {
		t.Error("unexpected err:", err)
	}
}

func TestUnmarshalNotRegistered(t *testing.T) {
	resetRegistry()

	data := []byte(`{"Version":1,"Text":"hello","Number":42}`)

	var value Simple
	err := json.Unmarshal(data, &value)
	if err == nil {
		t.Fatal("missing error")
	}
	if !strings.Contains(err.Error(), "registered") {
		t.Error("unexpected err:", err)
	}
}

func TestRegisterTwice(t *testing.T) {
	resetRegistry()
	Register(Simple{}, SimpleV1{})
	err := registerError(Simple{}, SimpleV1{})

	if err == nil {
		t.Fatal("missing error")
	}
	if !strings.Contains(err.Error(), "already registered") {
		t.Fatal("unexpected err:", err)
	}
}

func TestRegisterNonStruct(t *testing.T) {
	resetRegistry()
	err := registerError(1, SimpleV1{})

	if err == nil {
		t.Fatal("missing error")
	}
	if !strings.Contains(err.Error(), "only structs are allowed") {
		t.Fatal("unexpected err:", err)
	}
}

func TestRegisterVersionNonStruct(t *testing.T) {
	resetRegistry()
	err := registerError(Simple{}, 1)

	if err == nil {
		t.Fatal("missing error")
	}
	if !strings.Contains(err.Error(), "only structs are allowed") {
		t.Fatal("unexpected err:", err)
	}
}

func TestRegisterNoVersions(t *testing.T) {
	resetRegistry()
	err := registerError(Simple{})

	if err == nil {
		t.Fatal("missing error")
	}
	if !strings.Contains(err.Error(), "at least one version prototype") {
		t.Fatal("unexpected err:", err)
	}
}

func TestNegativeVersion(t *testing.T) {
	resetRegistry()
	Register(Simple{}, SimpleV1{})

	data := []byte(`{"Version":-1,"Text":"hello","Number":42}`)

	var value Simple
	err := json.Unmarshal(data, &value)
	if err == nil {
		t.Fatal("missing error")
	}
	if !strings.Contains(err.Error(), "negative") {
		t.Error("unexpected err:", err)
	}
}

func TestUnsupportedVersion(t *testing.T) {
	resetRegistry()
	Register(Simple{}, SimpleV1{})

	data := []byte(`{"Version":100,"Text":"hello","Number":42}`)

	var value Simple
	err := json.Unmarshal(data, &value)
	if err == nil {
		t.Fatal("missing error")
	}
	if !strings.Contains(err.Error(), "unsupported") {
		t.Error("unexpected err:", err)
	}
}

func TestUnmarshalSimple(t *testing.T) {
	resetRegistry()
	Register(Simple{}, SimpleV1{})

	data := []byte(`{"Version":1,"Text":"hello","Number":42}`)

	var value Simple
	err := json.Unmarshal(data, &value)
	if err != nil {
		t.Fatal("unexpected err:", err)
	}
	if value.Text != "hello" {
		t.Errorf("wrong text: %+v", value)
	}
	if value.Number != 42 {
		t.Errorf("wrong number: %+v", value)
	}
}

func TestUnmarshalSimpleNull(t *testing.T) {
	resetRegistry()
	Register(Simple{}, SimpleV1{})

	data := []byte(`null`)

	var value Simple
	err := json.Unmarshal(data, &value)
	if err != nil {
		t.Fatal("unexpected err:", err)
	}
}

type Empty struct{}

func (value *Empty) MarshalJSON() ([]byte, error) {
	return Marshal(value)
}

type EmptyV1 struct{}

func TestMarshalEmpty(t *testing.T) {
	resetRegistry()
	Register(Empty{}, EmptyV1{})

	value := Empty{}

	data, err := json.Marshal(&value)
	if err != nil {
		t.Fatal("unexpected err:", err)
	}

	str := string(data)
	if str != `{"Version":1}` {
		t.Fatal("wrong data:", str)
	}
}

type Multiple struct {
	B string
	C string
	D string
}

func (value *Multiple) UnmarshalJSON(data []byte) error {
	return Unmarshal(data, value)
}

type MultipleV1 struct {
	A string
	B string
}

type MultipleV2 struct {
	A string
	B string
	C string
}

type MultipleV3 struct {
	B string
	C string
	D string
}

func TestUnmarshalMultiple(t *testing.T) {
	resetRegistry()
	Register(Multiple{}, MultipleV1{}, MultipleV2{}, MultipleV3{})

	data := []byte(`{"Version":2,"A":"a","B":"b","C":"c"}`)

	var value Multiple
	err := json.Unmarshal(data, &value)
	if err != nil {
		t.Fatal("unexpected err:", err)
	}
	if value.B != "b" || value.C != "c" {
		t.Errorf("wrong value: %+v", value)
	}
}

type Renaming struct {
	X string
	Y string
}

func (value *Renaming) UnmarshalJSON(data []byte) error {
	return Unmarshal(data, value)
}

type RenamingV1 struct {
	A string
	B string
}

type RenamingV2 struct {
	X string `vjson:"A"`
	Y string
}

func TestUnmarshalRenaming(t *testing.T) {
	resetRegistry()
	Register(Renaming{}, RenamingV1{}, RenamingV2{})

	data := []byte(`{"Version":1,"A":"x","B":"b"}`)

	var value Renaming
	err := json.Unmarshal(data, &value)
	if err != nil {
		t.Fatal("unexpected err:", err)
	}
	if value.X != "x" || value.Y != "" {
		t.Errorf("wrong value: %+v", value)
	}
}

type Upgrade struct {
	BA string
}

func (value *Upgrade) UnmarshalJSON(data []byte) error {
	return Unmarshal(data, value)
}

type UpgradeV1 struct {
	A string
}

type UpgradeV2 struct {
	BA string
}

func (v2 *UpgradeV2) Upgrade(v1 *UpgradeV1) {
	v2.BA = "b" + v1.A
}

func TestUnmarshalUpgrade(t *testing.T) {
	resetRegistry()
	Register(Upgrade{}, UpgradeV1{}, UpgradeV2{})

	data := []byte(`{"Version":1,"A":"a"}`)

	var value Upgrade
	err := json.Unmarshal(data, &value)
	if err != nil {
		t.Fatal("unexpected err:", err)
	}
	if value.BA != "ba" {
		t.Errorf("wrong value: %+v", value)
	}
}

type UpgradeError struct {
	BA string
}

func (value *UpgradeError) UnmarshalJSON(data []byte) error {
	return Unmarshal(data, value)
}

type UpgradeErrorV1 struct {
	A string
}

type UpgradeErrorV2 struct {
	BA string
}

func (v2 *UpgradeErrorV2) Upgrade(v1 *UpgradeErrorV1) error {
	return fmt.Errorf("upgrade error")
}

func TestUnmarshalUpgradeError(t *testing.T) {
	resetRegistry()
	Register(UpgradeError{}, UpgradeErrorV1{}, UpgradeErrorV2{})

	data := []byte(`{"Version":1,"A":"a"}`)

	var value UpgradeError
	err := json.Unmarshal(data, &value)
	if err == nil {
		t.Fatal("missing error")
	}
	if err.Error() != "upgrade error" {
		t.Fatal("wrong error:", err)
	}
}

type NestedParent struct {
	Child NestedChild
}

type NestedChild struct {
	B string
}

func (value *NestedChild) UnmarshalJSON(data []byte) error {
	return Unmarshal(data, value)
}

type NestedChildV1 struct {
	A string
}

type NestedChildV2 struct {
	B string `vjson:"A"`
}

func TestUnmarshalNested(t *testing.T) {
	resetRegistry()
	Register(NestedChild{}, NestedChildV1{}, NestedChildV2{})

	data := []byte(`{"Child":{"Version":1,"A":"b"}}`)

	var value NestedParent
	err := json.Unmarshal(data, &value)
	if err != nil {
		t.Fatal("unexpected err:", err)
	}
	if value.Child.B != "b" {
		t.Errorf("wrong value: %+v", value)
	}
}

type EmbeddedChild struct {
	A string
}

type EmbeddedParent struct {
	EmbeddedChild
}

func (value *EmbeddedParent) UnmarshalJSON(data []byte) error {
	return Unmarshal(data, value)
}

type EmbeddedParentV1 struct {
	EmbeddedChild
}

type EmbeddedParentV2 struct {
	EmbeddedChild
}

func TestUnmarshalEmbedded(t *testing.T) {
	resetRegistry()
	Register(EmbeddedParent{}, EmbeddedParentV1{}, EmbeddedParentV2{})

	data := []byte(`{"Version":1,"A":"a"}`)

	var value EmbeddedParent
	err := json.Unmarshal(data, &value)
	if err != nil {
		t.Fatal("unexpected err:", err)
	}
	if value.A != "a" {
		t.Errorf("wrong value: %+v", value)
	}
}

type TypeMismatchA struct {
	Message string
}

type TypeMismatchAV1 struct {
	Message int
}

func TestRegisterTypeMismatchA(t *testing.T) {
	resetRegistry()
	err := registerError(TypeMismatchA{}, TypeMismatchAV1{})

	if err == nil {
		t.Fatal("missing error")
	}
	if !strings.Contains(err.Error(), "field Message has different types") {
		t.Fatal("unexpected err:", err)
	}
}

type TypeMismatchB struct {
	Message string
}

type TypeMismatchBV1 struct {
	Message int
}

type TypeMismatchBV2 struct {
	Message string
}

func TestRegisterTypeMismatchB(t *testing.T) {
	resetRegistry()
	err := registerError(TypeMismatchB{}, TypeMismatchBV1{}, TypeMismatchBV2{})

	if err == nil {
		t.Fatal("missing error")
	}
	if !strings.Contains(err.Error(), "field Message has different types") {
		t.Fatal("unexpected err:", err)
	}
}

type TypeMismatchC struct {
	Message string
}

type TypeMismatchCV1 struct {
	OldMessage int
}

type TypeMismatchCV2 struct {
	Message string `vjson:"OldMessage"`
}

func TestRegisterTypeMismatchC(t *testing.T) {
	resetRegistry()
	err := registerError(TypeMismatchC{}, TypeMismatchCV1{}, TypeMismatchCV2{})

	if err == nil {
		t.Fatal("missing error")
	}
	if !strings.Contains(err.Error(), "cannot copy field") || !strings.Contains(err.Error(), "different types") {
		t.Fatal("unexpected err:", err)
	}
}

type TypeConversion struct {
	Message string
}

func (value *TypeConversion) UnmarshalJSON(data []byte) error {
	return Unmarshal(data, value)
}

type TypeConversionV1 struct {
	Message int
}

type TypeConversionV2 struct {
	Message string `vjson:""`
}

func (v2 *TypeConversionV2) Upgrade(v1 *TypeConversionV1) {
	v2.Message = fmt.Sprintf("%d", v1.Message)
}

func TestTypeConversion(t *testing.T) {
	resetRegistry()
	Register(TypeConversion{}, TypeConversionV1{}, TypeConversionV2{})

	data := []byte(`{"Version":1,"Message":42}`)

	var value TypeConversion
	err := json.Unmarshal(data, &value)
	if err != nil {
		t.Fatal("unexpected err:", err)
	}
	if value.Message != "42" {
		t.Errorf("wrong value: %+v", value)
	}
}

type Raw struct {
	Message string
}

func (value *Raw) MarshalJSON() ([]byte, error) {
	return Marshal(value)
}

func (value *Raw) UnmarshalJSON(data []byte) error {
	return Unmarshal(data, value)
}

type RawV1 struct {
	Message json.RawMessage
}

func (latest *RawV1) Pack(value *Raw) error {
	msg, err := json.Marshal(value.Message)
	if err != nil {
		return err
	}

	latest.Message = json.RawMessage(msg)
	return nil
}

func (latest *RawV1) Unpack(value *Raw) error {
	return json.Unmarshal(latest.Message, &value.Message)
}

func TestMarshalRaw(t *testing.T) {
	resetRegistry()
	Register(Raw{}, RawV1{})

	value := Raw{Message: "hello"}

	data, err := json.Marshal(&value)
	if err != nil {
		t.Fatal("unexpected err:", err)
	}

	str := string(data)
	if !strings.Contains(str, `"Version":1`) {
		t.Fatal("wrong data:", str)
	}
	if !strings.Contains(str, `"Message":"hello"`) {
		t.Fatal("wrong data:", str)
	}
}

func TestUnmarshalRaw(t *testing.T) {
	resetRegistry()
	Register(Raw{}, RawV1{})

	data := []byte(`{"Version":1,"Message":"hello"}`)

	var value Raw
	err := json.Unmarshal(data, &value)
	if err != nil {
		t.Fatal("unexpected err:", err)
	}
	if value.Message != "hello" {
		t.Errorf("wrong value: %+v", value)
	}
}

func TestUnmarshalRawNull(t *testing.T) {
	resetRegistry()
	Register(Raw{}, RawV1{})

	data := []byte(`null`)

	var value Raw
	err := json.Unmarshal(data, &value)
	if err != nil {
		t.Fatal("unexpected err:", err)
	}
}

var testError = errors.New("test error")

type RawError struct {
	Message string
}

func (value *RawError) MarshalJSON() ([]byte, error) {
	return Marshal(value)
}

func (value *RawError) UnmarshalJSON(data []byte) error {
	return Unmarshal(data, value)
}

type RawErrorV1 struct {
	Message json.RawMessage
}

func (latest *RawErrorV1) Pack(value *RawError) error {
	return testError
}

func (latest *RawErrorV1) Unpack(value *RawError) error {
	return testError
}

func TestMarshalRawError(t *testing.T) {
	resetRegistry()
	Register(RawError{}, RawErrorV1{})

	value := RawError{Message: "hello"}

	_, err := json.Marshal(&value)
	if err == nil {
		t.Fatal("missing error")
	}
	if !errors.Is(err, testError) {
		t.Fatal("wrong error:", err)
	}
}

func TestUnmarshalRawError(t *testing.T) {
	resetRegistry()
	Register(RawError{}, RawErrorV1{})

	data := []byte(`{"Version":1,"Message":"hello"}`)

	var value RawError
	err := json.Unmarshal(data, &value)
	if err == nil {
		t.Fatal("missing error")
	}
	if !errors.Is(err, testError) {
		t.Fatal("wrong error:", err)
	}
}
