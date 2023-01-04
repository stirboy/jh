package yamlmap

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMapGet(t *testing.T) {
	tests := []struct {
		name      string
		key       string
		wantValue string
		wantError bool
	}{
		{
			name:      "get key",
			key:       "default",
			wantValue: "default value",
		},
		{
			name:      "get blank key",
			key:       "blank",
			wantValue: "",
		},
	}

	for _, tt := range tests {
		m := testMap()
		t.Run(tt.name, func(t *testing.T) {
			v, err := m.Get(tt.key)
			if tt.wantError {
				assert.EqualError(t, err, "not found")
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantValue, v.Value)
		})
	}
}

func TestMapSet(t *testing.T) {
	tests := []struct {
		name       string
		key        string
		value      *Map
		wantValue  string
		wantLength int
	}{
		{
			name:       "set key",
			key:        "default",
			value:      StringValue("value"),
			wantValue:  "value",
			wantLength: 6,
		},
		{
			name:       "set key that does not exist",
			key:        "nonExistentKey",
			value:      StringValue("value"),
			wantValue:  "value",
			wantLength: 8,
		},
	}

	for _, tt := range tests {
		m := testMap()
		t.Run(tt.name, func(t *testing.T) {
			m.Set(tt.key, tt.value)
			assert.Equal(t, tt.wantLength, len(m.Content))

			v, err := m.Get(tt.key)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantValue, v.Node.Value)
		})
	}
}

func TestMapAdd(t *testing.T) {
	tests := []struct {
		name       string
		key        string
		value      string
		wantValue  string
		wantLength int
	}{
		{
			name:       "add entry with key that is already present",
			key:        "default",
			value:      "some value",
			wantValue:  "default value",
			wantLength: 8,
		},
		{
			name:       "add entry with key that is not present",
			key:        "notPresent",
			value:      "some value",
			wantValue:  "some value",
			wantLength: 8,
		},
	}

	for _, tt := range tests {
		m := testMap()
		t.Run(tt.name, func(t *testing.T) {
			m.Add(tt.key, StringValue(tt.value))
			entry, err := m.Get(tt.key)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantValue, entry.Value)
			assert.Equal(t, tt.wantLength, len(m.Content))
		})

	}

}

func testMap() *Map {
	var data = `
default: default value
blank:
dog: who let the dogs out?
`
	m, _ := Unmarshal([]byte(data))
	return m
}
