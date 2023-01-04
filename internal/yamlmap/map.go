package yamlmap

import (
	"errors"

	"gopkg.in/yaml.v3"
)

var (
	ErrNotFound      = errors.New("not found")
	ErrInvalidYaml   = errors.New("invalid yaml")
	ErrInvalidFormat = errors.New("invalid format")
)

// Map is an abstraction to work with yaml
type Map struct {
	*yaml.Node
}

func StringValue(value string) *Map {
	return &Map{&yaml.Node{
		Kind:  yaml.ScalarNode,
		Tag:   "!!str",
		Value: value,
	}}
}

func MapValue() *Map {
	return &Map{&yaml.Node{
		Kind: yaml.MappingNode,
		Tag:  "!!map",
	}}
}

func Marshal(m *Map) ([]byte, error) {
	return yaml.Marshal(m.Node)
}

func Unmarshal(data []byte) (*Map, error) {
	var root yaml.Node
	err := yaml.Unmarshal(data, &root)
	if err != nil {
		return nil, ErrInvalidYaml
	}
	if len(root.Content) == 0 {
		return MapValue(), nil
	}
	if root.Content[0].Kind != yaml.MappingNode {
		return nil, ErrInvalidFormat
	}
	return &Map{root.Content[0]}, nil
}

func (m *Map) Get(key string) (*Map, error) {
	// Note: The content slice of a yamlMap looks like [key1, value1, key2, value2, ...].
	for i, v := range m.Content {
		if i%2 != 0 {
			continue
		}

		if v.Value == key {
			if i+1 < len(m.Content) {
				return &Map{m.Content[i+1]}, nil
			}
		}
	}

	return nil, ErrNotFound
}

func (m *Map) Set(key string, value *Map) {
	// Note: The content slice of a yamlMap looks like [key1, value1, key2, value2, ...].
	for i, v := range m.Content {
		if i%2 != 0 {
			continue
		}

		if v.Value == key {
			if i+1 < len(m.Content) {
				m.Content[i+1] = value.Node
				return
			}
		}
	}
	m.Add(key, value)
}

func (m *Map) Add(key string, value *Map) {
	// Note: The content slice of a yamlMap looks like [key1, value1, key2, value2, ...].
	// we need to add 2 values [key, value]
	keyNode := &yaml.Node{
		Kind:  yaml.ScalarNode,
		Tag:   "!!str",
		Value: key,
	}

	m.Content = append(m.Content, keyNode, value.Node)
}

func (m *Map) String() string {
	data, err := Marshal(m)
	if err != nil {
		return ""
	}
	return string(data)
}
