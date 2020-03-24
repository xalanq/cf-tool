package client

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"time"
	"strings"
	"errors"
	"net/url"
	"encoding/json"

	"github.com/fatih/color"
)

func (c *Client) CustomTest(langId int, source, input string) (err error) {
	color.Cyan("Custom Test %v", Langs[strconv.Itoa(langId)])

	resp, err := c.client.Get(c.host+"/problemset/customtest")
	if err != nil {
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	_, err = findHandle(body)
	if err != nil {
		return
	}

	fmt.Printf("Current user: %v\n", c.Handle)

	csrf, err := findCsrf(body)
	if err != nil {
		return
	}

	// fmt.Sprintf("%v?csrf_token=%v", URL, csrf)
	resp, err = c.client.PostForm(c.host+"/data/customtest", url.Values{
		"csrf_token":            {csrf},
		"source":                {source},
		"programTypeId":         {strconv.Itoa(langId)},
		"input":                 {input},
		//"_tta":                  {440},
		//"communityCode":         {},
		"action":                {"submitSourceCode"},
		"sourceCode":            {source},
	})
	if err != nil {
		return
	}

	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil { return }

	var customTestInfo map[string]string
	json.Unmarshal(body, &customTestInfo)

	var customTestSubmitId string
	var ok bool
	if customTestSubmitId, ok = customTestInfo["customTestSubmitId"]; !ok {
		errorMessage := ""
		for key, message := range customTestInfo {
			if errorMessage != "" {
				errorMessage += "; "
			}
			errorMessage += strings.TrimPrefix(key, "error__") + ": " + message
		}
		return errors.New(errorMessage)
	}

	color.Green("Submitted")
	for {
		time.Sleep(2500 * time.Millisecond)

		resp, err = c.client.PostForm(c.host+"/data/customtest", url.Values{
			"customTestSubmitId":    {customTestSubmitId},
			"csrf_token":            {csrf},
			//"communityCode":         {},
			"action":                {"getVerdict"},
		})
		if err != nil { return }

		defer resp.Body.Close()
		body, err = ioutil.ReadAll(resp.Body)
		if err != nil { return }

		var output struct {
			CustomTestSubmitId string
			Output string
			Stat string
			Verdict string
		}
		json.Unmarshal(body, &output)
		if output.CustomTestSubmitId != customTestSubmitId {
			color.Red("Error: Expected %v, actual %v", customTestSubmitId, output.CustomTestSubmitId)
		}
		if output.Verdict == "OK" {
			fmt.Printf("%v\n=====\nUsed: %v\n", output.Output, output.Stat)
			return
		}
		// otherwise there will be only CustomTestSubmitId field
	}
}
