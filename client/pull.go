package client

import (
	"errors"
	"fmt"
	"html"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/fatih/color"
)

func findCode(body []byte) (string, error) {
	reg := regexp.MustCompile(`<pre[\s\S]*?>([\s\S]*?)</pre>`)
	tmp := reg.FindSubmatch(body)
	if tmp == nil {
		return "", errors.New("Cannot find code")
	}
	return html.UnescapeString(string(tmp[1])), nil
}

// PullCode pull problem's code to path
func (c *Client) PullCode(codeURL, path, ext string) (filename string, err error) {
	client := &http.Client{Jar: c.Jar}
	resp, err := client.Get(codeURL)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	code, err := findCode(body)
	if err != nil {
		return
	}

	filename = path + ext
	i := 1
	for _, err := os.Stat(filename); err == nil; _, err = os.Stat(filename) {
		tmpPath := fmt.Sprintf("%v%v%v", path, i, ext)
		fmt.Printf("%v is existed. Rename to %v\n", filepath.Base(filename), filepath.Base(tmpPath))
		filename = tmpPath
		i++
	}

	err = os.MkdirAll(filepath.Dir(filename), os.ModePerm)
	if err != nil {
		return
	}

	err = ioutil.WriteFile(filename, []byte(code), 0644)
	return
}

// PullContest pull all latest codes or ac codes of contest's problem
func (c *Client) PullContest(contestID, problemID, rootPath string, ac bool) (err error) {
	color.Cyan("Pull code from %v%v, accepted: %v", contestID, problemID, ac)
	submissions, _, err := c.getSubmissions(fmt.Sprintf("https://codeforces.com/contest/%v/my", contestID), -1)
	if err != nil {
		return err
	}

	saved := map[string](map[string]bool){}
	used := []Submission{}

	for _, submission := range submissions {
		pid := strings.ToLower(strings.Split(submission.name, " ")[0])
		if problemID != "" && problemID != pid {
			continue
		}
		if ac && !(strings.Contains(submission.status, "Accepted") || strings.Contains(submission.status, "Pretests passed")) {
			continue
		}
		ext, ok := LangsExt[submission.lang]
		if !ok {
			continue
		}
		if _, ok = saved[pid]; !ok {
			saved[pid] = map[string]bool{}
		}
		if _, ok = saved[pid][ext]; ok {
			continue
		}
		path := ""
		if problemID == "" {
			path = filepath.Join(rootPath, pid, pid)
		} else {
			path = filepath.Join(rootPath, strings.ToLower(problemID))
		}
		filename, err := c.PullCode(
			fmt.Sprintf("https://codeforces.com/contest/%v/submission/%v", contestID, submission.id),
			path,
			"."+ext,
		)
		if err == nil {
			saved[pid][ext] = true
			color.Green(fmt.Sprintf(`Downloaded code of %v %v into %v`, contestID, problemID, filepath.Base(filename)))
			used = append(used, submission)
		}
	}

	if len(used) == 0 {
		return errors.New("Cannot find any code to save")
	}

	color.Cyan("These submissions' codes have been saved.")
	maxline := 0
	display(used, true, &maxline, false)
	return nil
}
