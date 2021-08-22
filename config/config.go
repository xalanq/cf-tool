package config

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/xalanq/cf-tool/client"
)

// CodeTemplate config parse code template
type CodeTemplate struct {
	Alias        string   `json:"alias"`
	Lang         string   `json:"lang"`
	Path         string   `json:"path"`
	Suffix       []string `json:"suffix"`
	BeforeScript string   `json:"before_script"`
	Script       string   `json:"script"`
	AfterScript  string   `json:"after_script"`
}

// Config load and save configuration
type Config struct {
	Template      []CodeTemplate    `json:"template"`
	Default       int               `json:"default"`
	GenAfterParse bool              `json:"gen_after_parse"`
	Host          string            `json:"host"`
	Proxy         string            `json:"proxy"`
	FolderName    map[string]string `json:"folder_name"`
	path          string
}

// Instance global configuration
var Instance *Config

// Init initialize
func Init(path string) {
	c := &Config{path: path, Host: "https://codeforces.com", Proxy: ""}
	if err := c.load(); err != nil {
		color.Red(err.Error())
		color.Green("Create a new configuration in %v", path)
	}
	if c.Default < 0 || c.Default >= len(c.Template) {
		c.Default = 0
	}
	if c.FolderName == nil {
		c.FolderName = map[string]string{}
	}
	if _, ok := c.FolderName["root"]; !ok {
		c.FolderName["root"] = "cf"
	}
	for _, problemType := range client.ProblemTypes {
		if _, ok := c.FolderName[problemType]; !ok {
			c.FolderName[problemType] = problemType
		}
	}
	c.save()
	Instance = c
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
	var data bytes.Buffer
	encoder := json.NewEncoder(&data)
	encoder.SetIndent("", "  ")
	encoder.SetEscapeHTML(false)
	err = encoder.Encode(c)
	if err == nil {
		os.MkdirAll(filepath.Dir(c.path), os.ModePerm)
		err = ioutil.WriteFile(c.path, data.Bytes(), 0644)
	}
	if err != nil {
		color.Red("Cannot save config to %v\n%v", c.path, err.Error())
	}
	return
}
