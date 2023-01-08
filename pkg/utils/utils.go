package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	jira "github.com/andygrunwald/go-jira/v2/cloud"
)

func MapKeys[K string, V any](m map[K]V) ([]K, error) {
	if m == nil {
		return nil, errors.New("nil cannot be passed")
	}

	slice := []K{}
	for k := range m {
		slice = append(slice, k)
	}

	return slice, nil
}

func ParseJiraResponse(resp *jira.Response) error {
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.New("can't parse response body")
	}

	// format
	formattedBody := &bytes.Buffer{}
	if err = json.Indent(formattedBody, bodyBytes, "", "  "); err != nil {
		// let's return unformatted body
		return errors.New(string(bodyBytes))
	}

	return errors.New("\n" + formattedBody.String())
}

func ParseResponse(resp *http.Response) ([]byte, error) {
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return bodyBytes, nil
}
