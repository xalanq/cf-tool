package config

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/fatih/color"
)

// CodeTemplate config parse code template
type CodeTemplate struct {
	Alias  string   `json:"alias"`
	Lang   string   `json:"lang"`
	Path   string   `json:"path"`
	Suffix []string `json:"suffix"`
	Build  string   `json:"build"`
}

// Config load and save configuration
type Config struct {
	Username string         `json:"username"`
	Password string         `json:"password"`
	Template []CodeTemplate `json:"template"`
	Default  int            `json:"default"`
	path     string
}

// New an empty config
func New(path string) *Config {
	c := &Config{path: path}
	if err := c.load(); err != nil {
		return &Config{path: path}
	}
	return c
}

// load from path
func (c *Config) load() (err error) {
	file, err := os.Open(c.path)
	if err != nil {
		return
	}
	defer file.Close()

	bytes, err := ioutil.ReadAll(file)

	if err != nil {
		return err
	}

	return json.Unmarshal(bytes, c)
}

// save file to path
func (c *Config) save() (err error) {
	data, err := json.MarshalIndent(c, "", "  ")
	if err == nil {
		err = ioutil.WriteFile(c.path, data, 0644)
	}
	if err != nil {
		color.Red("Cannot save config to %v\n%v", c.path, err.Error())
	}
	return
}
