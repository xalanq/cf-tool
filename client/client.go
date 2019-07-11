package client

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/fatih/color"
	"github.com/xalanq/cf-tool/cookiejar"
)

// Client codeforces client
type Client struct {
	Jar            *cookiejar.Jar  `json:"cookies"`
	Username       string          `json:"username"`
	Ftaa           string          `json:"ftaa"`
	Bfaa           string          `json:"bfaa"`
	LastSubmission *SaveSubmission `json:"last_submission"`
	path           string
}

// New client
func New(path string) *Client {
	jar, _ := cookiejar.New(nil)
	c := &Client{Jar: jar, path: path, LastSubmission: nil}
	c.load()
	return c
}

// load from path
func (c *Client) load() (err error) {
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
func (c *Client) save() (err error) {
	data, err := json.MarshalIndent(c, "", "  ")
	if err == nil {
		err = ioutil.WriteFile(c.path, data, 0644)
	}
	if err != nil {
		color.Red("Cannot save session to %v\n%v", c.path, err.Error())
	}
	return
}
