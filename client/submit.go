package client

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"regexp"

	"github.com/fatih/color"
)

// SaveSubmission save it in session
type SaveSubmission struct {
	ContestID    string `json:"contest_id"`
	SubmissionID string `json:"submission_id"`
}

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

	URL := ToGym(fmt.Sprintf("https://codeforces.com/contest/%v/submit", contestID), contestID)
	resp, err := c.client.Get(URL)
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

	resp, err = c.client.PostForm(fmt.Sprintf("%v?csrf=%v", URL, csrf), url.Values{
		"csrf_token":            {csrf},
		"ftaa":                  {c.Ftaa},
		"bfaa":                  {c.Bfaa},
		"action":                {"submitSolutionFormSubmitted"},
		"submittedProblemIndex": {problemID},
		"programTypeId":         {langID},
		"source":                {source},
		"tabSize":               {"4"},
		"_tta":                  {"594"},
		"sourceCodeConfirmed":   {"true"},
	})
	if err != nil {
		return
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

	submissions, err := c.WatchSubmission(contestID, "", 1, true)
	if err != nil {
		return
	}

	c.LastSubmission = &SaveSubmission{
		ContestID:    contestID,
		SubmissionID: submissions[0].ParseID(),
	}
	c.save()

	return
}
