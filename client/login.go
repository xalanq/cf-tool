package client

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"

	"github.com/fatih/color"
	"github.com/xalanq/cf-tool/cookiejar"
	"github.com/xalanq/cf-tool/util"
)

// genFtaa generate a random one
func genFtaa() string {
	return util.RandString(18)
}

// genBfaa generate a bfaa
func genBfaa() string {
	return "f1b3f18c715565b589b7823cda7448ce"
}

// ErrorNotLogged not logged in
var ErrorNotLogged = "Not logged in"

// checkLogin if login return nil
func checkLogin(username string, body []byte) error {
	match, err := regexp.Match(fmt.Sprintf(`handle = "%v"`, username), body)
	if err != nil || !match {
		return errors.New(ErrorNotLogged)
	}
	return nil
}

// findCsrf just find
func findCsrf(body []byte) (string, error) {
	reg, _ := regexp.Compile(`csrf='(.+?)'`)
	tmp := reg.FindSubmatch(body)
	if len(tmp) < 2 {
		return "", errors.New("Cannot find csrf")
	}
	return string(tmp[1]), nil
}

// Login codeforces with username(handler) and password
func (c *Client) Login(username, password string) (err error) {
	jar, _ := cookiejar.New(nil)
	color.Cyan("Login %v...\n", username)

	client := &http.Client{Jar: jar}

	resp, err := client.Get("https://codeforces.com/enter")
	if err != nil {
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	csrf, err := findCsrf(body)
	if err != nil {
		return
	}

	ftaa := genFtaa()
	bfaa := genBfaa()

	resp, err = client.PostForm("https://codeforces.com/enter", url.Values{
		"csrf_token":    {csrf},
		"action":        {"enter"},
		"ftaa":          {ftaa},
		"bfaa":          {bfaa},
		"handleOrEmail": {username},
		"password":      {password},
		"_tta":          {"176"},
	})
	if err != nil {
		return
	}
	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	err = checkLogin(username, body)
	if err != nil {
		return
	}

	c.Jar = jar
	c.Ftaa = ftaa
	c.Bfaa = bfaa
	c.Username = username
	color.Green("Succeed!!")
	return c.save()
}
