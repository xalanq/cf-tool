package client

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"

	"cf-tool/cookiejar"
	"cf-tool/util"
	"github.com/fatih/color"
)

// Client codeforces client
type Client struct {
	Jar            *cookiejar.Jar  `json:"cookies"`
	Username       string          `json:"username"`
	Ftaa           string          `json:"ftaa"`
	Bfaa           string          `json:"bfaa"`
	LastSubmission *SaveSubmission `json:"last_submission"`
	Host           string          `json:"host"`
	path           string
	client         *http.Client
}

func formatHost(host string) (string, error) {
	if len(host) == 0 {
		return "https://codeforces.com", nil
	}
	reg := regexp.MustCompile(`https?://[a-zA-Z0-9\-]+(\.[a-zA-Z0-9\-]+)+/*`)
	if !reg.MatchString(host) {
		return "", fmt.Errorf(`Invalid host "%v"`, host)
	}
	for host[len(host)-1:] == "/" {
		host = host[:len(host)-1]
	}
	return host, nil
}

// New client
func New(path string) *Client {
	jar, _ := cookiejar.New(nil)
	c := &Client{Jar: jar, LastSubmission: nil, path: path, client: nil}
	if path != "" {
		c.load()
	}
	c.client = &http.Client{Jar: c.Jar}
	var err error
	c.Host, err = formatHost(c.Host)
	if err != nil {
		color.Red(err.Error() + `. Use default host "https://codeforces.com"`)
		color.Red(`Please use "cf config" to set a valid host later`)
		c.Host = "https://codeforces.com"
	}
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

// SetHost set host for Codeforces
func (c *Client) SetHost() (err error) {
	host, err := formatHost(c.Host)
	if err != nil {
		host = "https://codeforces.com"
	}
	color.Green("Current host domain is %v", host)
	color.Cyan(`Set a new host domain (e.g. "https://codeforces.com"`)
	color.Cyan(`Note: Don't forget the "http://" or "https://"`)
	for {
		host, err = formatHost(util.ScanlineTrim())
		if err == nil {
			break
		}
		color.Red(err.Error())
	}
	c.Host = host
	color.Green("New host domain is %v", host)
	return c.save()
}
