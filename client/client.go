package client

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"regexp"

	"github.com/fatih/color"
	"github.com/xalanq/cf-tool/cookiejar"
	"github.com/xalanq/cf-tool/util"
)

// Client codeforces client
type Client struct {
	Jar            *cookiejar.Jar  `json:"cookies"`
	Username       string          `json:"username"`
	Ftaa           string          `json:"ftaa"`
	Bfaa           string          `json:"bfaa"`
	LastSubmission *SaveSubmission `json:"last_submission"`
	Host           string          `json:"host"`
	Proxy          string          `json:"proxy"`
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

func formatProxy(proxy string) (string, error) {
	if len(proxy) == 0 {
		return "", nil
	}
	reg := regexp.MustCompile(`(https?|socks5)://[\w\-\.]+:\d+`)
	if !reg.MatchString(proxy) {
		return "", fmt.Errorf(`Invalid proxy "%v"`, proxy)
	}
	for proxy[len(proxy)-1:] == "/" {
		proxy = proxy[:len(proxy)-1]
	}
	return proxy, nil
}

// New client
func New(path string) *Client {
	jar, _ := cookiejar.New(nil)
	c := &Client{Jar: jar, LastSubmission: nil, path: path, client: nil}
	if path != "" {
		c.load()
	}
	proxyURL, err := formatProxy(c.Proxy)
	proxy := http.ProxyFromEnvironment
	if err != nil {
		color.Red(err.Error() + `. Use default proxy from environment`)
		color.Red(`Please use "cf config" to set a valid proxy later`)
	} else if proxyURL != "" {
		proxyURL, err := url.Parse(c.Proxy)
		if err == nil {
			proxy = http.ProxyURL(proxyURL)
		}
	}
	c.client = &http.Client{Jar: c.Jar, Transport: &http.Transport{Proxy: proxy}}
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

// SetProxy set proxy for client
func (c *Client) SetProxy() (err error) {
	proxy, err := formatProxy(c.Proxy)
	if err != nil {
		proxy = ""
	}
	if len(proxy) == 0 {
		color.Green("Current proxy is based on environment")
	} else {
		color.Green("Current proxy is %v", proxy)
	}
	color.Cyan(`Set a new proxy (e.g. "http://127.0.0.1:80", "socks5://127.0.0.1:1080"`)
	color.Cyan(`Enter empty line if you want to use default proxy from environment`)
	color.Cyan(`Note: Proxy URL should match "proxyProtocol://proxyIp:proxyPort"`)
	for {
		proxy, err = formatProxy(util.ScanlineTrim())
		if err == nil {
			break
		}
		color.Red(err.Error())
	}
	c.Proxy = proxy
	if len(proxy) == 0 {
		color.Green("Current proxy is based on environment")
	} else {
		color.Green("Current proxy is %v", proxy)
	}
	return c.save()
}
