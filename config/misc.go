package config

import (
	"fmt"
	"regexp"

	"github.com/fatih/color"
	"github.com/xalanq/cf-tool/util"
)

// SetGenAfterParse set it yes or no
func (c *Config) SetGenAfterParse() (err error) {
	c.GenAfterParse = util.YesOrNo(`Run "cf gen" after "cf parse" (y/n)? `)
	return c.save()
}

func formatHost(host string) (string, error) {
	reg := regexp.MustCompile(`https?://[\w\-]+(\.[\w\-]+)+/?`)
	if !reg.MatchString(host) {
		return "", fmt.Errorf(`Invalid host "%v"`, host)
	}
	for host[len(host)-1:] == "/" {
		host = host[:len(host)-1]
	}
	return host, nil
}

func formatProxy(proxy string) (string, error) {
	reg := regexp.MustCompile(`[\w\-]+?://[\w\-]+(\.[\w\-]+)*(:\d+)?`)
	if !reg.MatchString(proxy) {
		return "", fmt.Errorf(`Invalid proxy "%v"`, proxy)
	}
	return proxy, nil
}

// SetHost set host for Codeforces
func (c *Config) SetHost() (err error) {
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
func (c *Config) SetProxy() (err error) {
	proxy, err := formatProxy(c.Proxy)
	if err != nil {
		proxy = ""
	}
	if len(proxy) == 0 {
		color.Green("Current proxy is based on environment")
	} else {
		color.Green("Current proxy is %v", proxy)
	}
	color.Cyan(`Set a new proxy (e.g. "http://127.0.0.1:2333", "socks5://127.0.0.1:1080"`)
	color.Cyan(`Enter empty line if you want to use default proxy from environment`)
	color.Cyan(`Note: Proxy URL should match "protocol://host[:port]"`)
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

/*
// SetFolderName set folder name
func (c *Config) SetFolderName() (err error) {
	color.Cyan(`Set folders' name`)
	color.Cyan(`Enter empty line if you don't want to change the value`)
	color.Green(`Root path (current: %v)`, c.FolderName["root"])
	if value := util.ScanlineTrim(); value != "" {
		c.FolderName["root"] = value
	}
	for _, problemType := range client.ProblemTypes {
		color.Green(`%v path (current: %v)`, problemType, c.FolderName[problemType])
		if value := util.ScanlineTrim(); value != "" {
			c.FolderName[problemType] = value
		}
	}
	return c.save()
}
*/
