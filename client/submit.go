package client

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"

	"github.com/fatih/color"
)

func findErrorSource(body []byte) ([]byte, error) {
	reg := regexp.MustCompile(`"error\sfor__source">(.*?)</span>`)
	tmp := reg.FindSubmatch(body)
	if tmp == nil {
		return nil, errors.New("Cannot find error")
	}
	return tmp[1], nil
}

// SubmitContest submit problem in contest (and block util pending)
func (c *Client) SubmitContest(contestID, problemID, langID, source string) (err error) {
	color.Cyan("Submit %v %v %v", contestID, problemID, Langs[langID])
	submitURL := fmt.Sprintf("https://codeforces.com/contest/%v/submit", contestID)

	client := &http.Client{Jar: c.Jar}
	resp, err := client.Get(submitURL)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	err = checkLogin(c.Username, body)
	if err != nil {
		return
	}

	fmt.Printf("Current user: %v\n", c.Username)

	csrf, err := findCsrf(body)
	if err != nil {
		return
	}

	resp, err = client.PostForm(fmt.Sprintf("%v?csrf=%v", submitURL, csrf), url.Values{
		"csrf_token":            {csrf},
		"ftaa":                  {c.Ftaa},
		"bfaa":                  {c.Bfaa},
		"action":                {"submitSolutionFormSubmitted"},
		"submittedProblemIndex": {problemID},
		"programTypeId":         {langID},
		"source":                {source},
		"tabSize":               {"4"},
		"_tta":                  {"594"},
	})
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	sourceError, err := findErrorSource(body)
	if err == nil {
		return errors.New(string(sourceError))
	}
	color.Green("Submitted")

	return c.WatchSubmission(fmt.Sprintf("https://codeforces.com/contest/%v/my", contestID), 1, true)
}
