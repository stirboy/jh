package utils

import (
	"errors"
	"io"
	"net/http"
	"reflect"

	jira "github.com/andygrunwald/go-jira/v2/cloud"
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

func ParseJiraResponse(resp *jira.Response) error {
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.New("can't parse response body")
	}
	bodyString := string(bodyBytes)
	return errors.New(bodyString)
}

func ParseResponse(resp *http.Response) ([]byte, error) {
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return bodyBytes, nil
}

func getMapType(val interface{}) (reflect.Type, bool) {
	t := reflect.TypeOf(val)
	return t, t.Kind() == reflect.Map
}
