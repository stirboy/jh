package utils

import (
	"errors"
	"reflect"
)

func MapStringKeys(val interface{}) ([]string, error) {
	if val == nil {
		return nil, errors.New("nil cannot be passed")
	}

	t, isMap := getMapType(val)
	if !isMap {
		return nil, errors.New("val passed is not a map")
	}

	if t.Key().Kind() != reflect.String {
		return nil, errors.New("map key should be string")
	}

	v := reflect.ValueOf(val)
	keys := v.MapKeys()

	length := len(keys)

	resultSlice := reflect.MakeSlice(reflect.SliceOf(t.Key()), length, length)

	for i, key := range keys {
		resultSlice.Index(i).Set(key)
	}

	return resultSlice.Interface().([]string), nil
}

func getMapType(val interface{}) (reflect.Type, bool) {
	t := reflect.TypeOf(val)
	return t, t.Kind() == reflect.Map
}
