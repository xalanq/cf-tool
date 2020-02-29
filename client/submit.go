package client

import (
	"bytes"
	"errors"
	"fmt"
	"net/url"
	"regexp"

	"github.com/xalanq/cf-tool/util"

	"github.com/fatih/color"
)

func findErrorMessage(body []byte) ([]byte, error) {
	reg := regexp.MustCompile(`error[a-zA-Z_\-\ ]*">(.*?)</span>`)
	tmp := reg.FindSubmatch(body)
	if tmp == nil {
		return nil, errors.New("Cannot find error")
	}
	return tmp[1], nil
}

// Submit submit (block while pending)
func (c *Client) Submit(info Info, langID, source string) (err error) {
	color.Cyan("Submit " + info.Hint())

	URL, err := info.SubmitURL(c.host)
	if err != nil {
		return
	}

	body, err := util.GetBody(c.client, URL)
	if err != nil {
		return
	}

	handle, err := findHandle(body)
	if err != nil {
		return
	}

	fmt.Printf("Current user: %v\n", handle)

	csrf, err := findCsrf(body)
	if err != nil {
		return
	}

	body, err = util.PostBody(c.client, fmt.Sprintf("%v?csrf_token=%v", URL, csrf), url.Values{
		"csrf_token":            {csrf},
		"ftaa":                  {c.Ftaa},
		"bfaa":                  {c.Bfaa},
		"action":                {"submitSolutionFormSubmitted"},
		"submittedProblemIndex": {info.ProblemID},
		"programTypeId":         {langID},
		"source":                {source},
		"tabSize":               {"4"},
		"_tta":                  {"594"},
		"sourceCodeConfirmed":   {"true"},
	})
	if err != nil {
		return
	}

	errorMessage, err := findErrorMessage(body)
	if err == nil {
		return errors.New(string(errorMessage))
	}
	if !bytes.Contains(body, []byte("submitted successfully")) {
		return errors.New("Submit failed")
	}
	color.Green("Submitted")

	submissions, err := c.WatchSubmission(info, 1, true)
	if err != nil {
		return
	}

	info.SubmissionID = submissions[0].ParseID()
	c.Handle = handle
	c.LastSubmission = &info
	return c.save()
}
