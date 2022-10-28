// Package jsonless is a Go library that populates structs from JSON and
// allows serialization back to JSON without losing fields that are
// not explicitly defined in the struct.
package jsonless

import (
	"encoding/json"
	"errors"
	"github.com/bingoohuang/gg/pkg/ss"
	"reflect"
	"strings"

	"github.com/bingoohuang/gg/pkg/mapstruct"
)

// The JSON type contains the state of the decoded data.  Embed this
// type in your type and implement MarshalJSON and UnmarshalJSON
// methods to add lossless encoding and decoding.
//
// Example:
//
//	type Person struct {
//	    lossless.JSON `json:"-"`
//
//	    Name    string
//	    Age     int
//	    Address string
//	}
//
//	func (p *Person) UnmarshalJSON(data []byte) error {
//	    return p.JSON.UnmarshalJSON(p, data)
//	}
//
//	func (p Person) MarshalJSON() ([]byte, error) {
//	    return p.JSON.MarshalJSON(p)
//	}
type JSON struct {
	json *Simple
}

func (js *JSON) maybeInit() {
	if js.json == nil {
		js.json, _ = NewSimple([]byte("{}"))
	}
}

// Set sets a JSON value not represented in the struct type.  The
// argument list is a set of strings referring to the JSON path,
// with the value to be set as the last value.
//
// Example:
//
//	// This sets {"Phone": {"Mobile": "614-555-1212"}} in the JSON
//	p.Set("Phone", "Mobile", "614-555-1212")
func (js *JSON) Set(args ...interface{}) error {
	js.maybeInit()

	if len(args) < 2 {
		return errors.New("rs must contain a path and value")
	}

	v := args[len(args)-1]
	key, ok := args[len(args)-2].(string)
	if !ok {
		return errors.New("all args except last must be strings")
	}

	j := js.json
	for _, p := range args[:len(args)-2] {
		strp, ok := p.(string)
		if !ok {
			return errors.New("all args except last must be strings")
		}

		newj, ok := j.CheckGet(strp)
		if !ok {
			j.Set(strp, make(map[string]interface{}))
			j = j.Get(strp)
		} else {
			j = newj
		}
	}

	j.Set(key, v)

	return nil
}

// UnmarshalJSON unmarshals JSON data into the given destination.  Users should
// call this from their type's UnmarshalJSON method.
//
// Example:
//
//	func (p *Person) UnmarshalJSON(data []byte) error {
//	    return p.JSON.UnmarshalJSON(p, data)
//	}
func (js *JSON) UnmarshalJSON(dest interface{}, data []byte) error {
	j, err := NewSimple(data)
	if err != nil {
		return err
	}

	js.json = j

	config := &mapstruct.Config{
		WeakType: true,
		Result:   dest,
	}

	decoder, err := mapstruct.NewDecoder(config)
	if err != nil {
		panic(err)
	}
	return decoder.Decode(j.data)
}

// MarshalJSON marshals the given source into JSON data.  Users should
// call this from their type's MarshalJSON method.
//
// Example:
//
//	func (p Person) MarshalJSON() ([]byte, error) {
//	    return p.JSON.MarshalJSON(p)
//	}
func (js *JSON) MarshalJSON(src interface{}) ([]byte, error) {
	js.maybeInit()
	if err := syncFromStruct(src, js.json); err != nil {
		return nil, err
	}

	return json.Marshal(js.json)
}

func syncFromStruct(src interface{}, j *Simple) error {
	dv := reflect.Indirect(reflect.ValueOf(src))
	dt := dv.Type()

	// This skips the encoding/json "json" tag's "omitempty" value.
	for i := 0; i < dt.NumField(); i++ {
		sf := dt.Field(i)
		tag := sf.Tag.Get("json")
		if tag == "-" {
			continue
		}

		var f reflect.Value
		var tagName string
		var omitempty bool

		if j := strings.Index(tag, ","); j != -1 {
			tagName = tag[:j]
			if tagName == "-" {
				continue
			}

			f = dv.Field(i)
			// If "omitempty" is specified in the tag, it ignores empty values.
			omitempty = strings.Contains(tag[j+1:], "omitempty") && isEmptyValue(f)
		} else {
			tagName = tag
			f = dv.Field(i)
		}

		name := ss.If(tagName == "", sf.Name, tagName)
		if omitempty {
			j.Del(name)
		} else {
			j.Set(name, f.Interface())
		}
	}

	return nil
}

func isEmptyValue(v reflect.Value) bool {
	switch getKind(v) {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	}
	return false
}

func getKind(val reflect.Value) reflect.Kind {
	kind := val.Kind()

	switch {
	case kind >= reflect.Int && kind <= reflect.Int64:
		return reflect.Int
	case kind >= reflect.Uint && kind <= reflect.Uint64:
		return reflect.Uint
	case kind >= reflect.Float32 && kind <= reflect.Float64:
		return reflect.Float32
	default:
		return kind
	}
}
