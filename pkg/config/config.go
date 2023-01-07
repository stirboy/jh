package config

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/stirboy/jh/internal/yamlmap"
)

const JhConfigDir = "JH_CONFIG_DIR"

var (
	c         *cfg
	once      sync.Once
	loadError error
)

//go:generate moq -rm -out config_mock.go . Config
type Config interface {
	AuthToken() (string, error)
	Get(string) (string, error)
	Set(string, string)
	Write() error
}

func NewConfig() (Config, error) {
	config, err := Read()
	if err != nil {
		return nil, err
	}
	return config, nil
}

// cfg implements Config Interface
type cfg struct {
	entries *yamlmap.Map
	mu      sync.RWMutex
}

func (c *cfg) AuthToken() (string, error) {
	val, err := c.Get("token")
	if err != nil {
		return "", err
	}

	return val, nil
}

func (c *cfg) Get(key string) (string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	m, err := c.entries.Get(key)
	if err != nil {
		return "", err
	}
	return m.Value, nil
}

func (c *cfg) Set(key, val string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries.Set(key, yamlmap.StringValue(val))
}

func (c *cfg) Write() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	err := writeFile(generalConfigFile(), []byte(c.entries.String()))
	if err != nil {
		return err
	}

	return nil
}

func generalConfigFile() string {
	if c := os.Getenv(JhConfigDir); c != "" {
		return filepath.Join(c, "config.yml")
	}
	d, _ := os.UserHomeDir()
	return filepath.Join(d, ".config", "jh", "config.yml")
}

func writeFile(filename string, data []byte) error {
	err := os.MkdirAll(filepath.Dir(filename), 0771)
	if err != nil {
		return err
	}
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.Write(data)
	return err
}

var Read = func() (*cfg, error) {
	once.Do(func() {
		c, loadError = load(generalConfigFile())
	})
	return c, loadError
}

func load(path string) (*cfg, error) {
	m, err := mapFromFile(path)
	if err != nil && !os.IsNotExist(err) {
		fmt.Println("File does not exist")
		return nil, err
	}

	if m == nil {
		m, _ = yamlmap.Unmarshal([]byte(defaultGeneralEntries))
	}

	return &cfg{entries: m}, nil
}

func mapFromFile(filename string) (*yamlmap.Map, error) {
	data, err := readFile(filename)
	if err != nil {
		return nil, err
	}

	return yamlmap.Unmarshal(data)
}

func readFile(filename string) ([]byte, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	data, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	return data, nil
}

var defaultGeneralEntries = `
# What jira url to use. Ex. my-company.attlasian.net
url:
# What username to use for auth. Ex. my-email@gmail.com
username:
# Jira API Token
token:
`

// ReadFromString takes a yaml string and returns a Config.
// Note: This is only used for testing
func ReadFromString(str string) *cfg {
	m, _ := yamlmap.Unmarshal([]byte(str))
	if m == nil {
		m = yamlmap.MapValue()
	}
	return &cfg{entries: m}
}
