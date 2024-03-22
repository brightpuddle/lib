// Package json provides a set of functions to manipulate JSON data
package json

import (
	"errors"
	"fmt"

	"lib/logger"

	"github.com/goccy/go-reflect"

	"github.com/segmentio/encoding/json"
	"github.com/tidwall/sjson"
)

// Set sets a json value at path
func Set(body, path string, value interface{}) string {
	res, err := sjson.Set(body, path, value)
	if err != nil {
		logger.Get().
			Error().
			Err(err).
			Msgf("cannot set JSON value")
	}
	return res
}

// SetRaw sets raw json at path
func SetRaw(body, path, value string) string {
	res, err := sjson.SetRaw(body, path, value)
	if err != nil {
		logger.Get().
			Error().
			Err(err).
			Msgf("cannot set JSON value")
	}
	return res
}

// ptr wraps the given value with pointer: V => *V, *V => **V, etc.
func ptr(v reflect.Value) reflect.Value {
	pt := reflect.PtrTo(v.Type())
	pv := reflect.New(pt.Elem())
	pv.Elem().Set(v)
	return pv
}

func allocNull(res interface{}) error {
	v := reflect.ValueOf(res)
	return allocNullValue(v)
}

func allocNullValue(v reflect.Value) error {
	if v.Kind() != reflect.Ptr {
		return fmt.Errorf("%s is not a pointer", v.Kind())
	}
	v = v.Elem()
	if v.Kind() != reflect.Struct {
		return fmt.Errorf("%s is not a struct", v.Kind())
	}
	if !v.CanSet() {
		return errors.New("value not settable")
	}

	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		if f.Kind() == reflect.Struct {
			allocNullValue(f.Addr())
		}
		switch f.Kind() {
		case reflect.Struct:
			allocNullValue(f.Addr())
			continue
		case reflect.Slice, reflect.Map, reflect.Ptr:
			if !f.IsNil() {
				continue
			}
		}
		switch f.Kind() {
		case reflect.Slice:
			f.Set(reflect.MakeSlice(f.Type(), 0, 0))
		case reflect.Map:
			f.Set(reflect.MakeMap(f.Type()))
		case reflect.Ptr:
			switch f.Type().Elem().Kind() {
			case reflect.Slice:
				f.Set(ptr(reflect.MakeSlice(f.Type().Elem(), 0, 0)))
			case reflect.Map:
				f.Set(ptr(reflect.MakeMap(f.Type().Elem())))
			}
		}
	}
	return nil
}

// Marshal obj into a JSON string
func Marshal(obj interface{}) string {
	// Silently fails if obj is not a pointer
	allocNull(obj)
	res, err := json.Marshal(obj)
	if err != nil {
		logger.Get().
			Error().
			Err(err).
			Msgf("cannot marshal JSON")
	}
	return string(res)
}

// Unmarshal JSON data into obj
func Unmarshal(data string, obj interface{}) error {
	return json.Unmarshal([]byte(data), obj)
}
