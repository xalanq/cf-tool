package client

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/xalanq/cf-tool/util"

	"github.com/fatih/color"
)

func findErrorMessage(body []byte) (string, error) {
	reg := regexp.MustCompile(`error[a-zA-Z_\-\ ]*">(.*?)</span>`)
	tmp := reg.FindSubmatch(body)
	if tmp == nil {
		return "", errors.New("Cannot find error")
	}
	return string(tmp[1]), nil
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
		"contestId":             {info.ContestID},
		"source":                {source},
		"tabSize":               {"4"},
		"_tta":                  {"594"},
		"sourceCodeConfirmed":   {"true"},
	})
	if err != nil {
		return
	}

	errMsg, err := findErrorMessage(body)
	if err == nil {
		return errors.New(errMsg)
	}

	msg, err := findMessage(body)
	if err != nil {
		return errors.New("Submit failed")
	}
	if !strings.Contains(msg, "submitted successfully") {
		return errors.New(msg)
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
