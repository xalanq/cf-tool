package client

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/fatih/color"
)

func findSample(body []byte) (input [][]byte, output [][]byte, err error) {
	irg, _ := regexp.Compile(`class="input"[\s\S]*?<pre>\s*([\s\S]*?)\s*</pre>`)
	org, _ := regexp.Compile(`class="output"[\s\S]*?<pre>\s*([\s\S]*?)\s*</pre>`)
	a := irg.FindAllSubmatch(body, -1)
	b := org.FindAllSubmatch(body, -1)
	if a == nil || b == nil || len(a) != len(b) {
		return nil, nil, fmt.Errorf("Cannot parse sample with input %v and output %v", len(a), len(b))
	}
	for i := 0; i < len(a); i++ {
		input = append(input, a[i][1])
		output = append(output, b[i][1])
	}
	return
}

// ParseProblem parse problem to path
func (c *Client) ParseProblem(probURL, path string) (err error) {
	client := &http.Client{Jar: c.Jar}
	resp, err := client.Get(probURL)
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

	input, output, err := findSample(body)
	if err != nil {
		return
	}

	for i := 0; i < len(input); i++ {
		fileIn := filepath.Join(path, fmt.Sprintf("in%v.txt", i+1))
		fileOut := filepath.Join(path, fmt.Sprintf("ans%v.txt", i+1))
		e := ioutil.WriteFile(fileIn, input[i], 0644)
		if e != nil {
			color.Red(e.Error())
		}
		e = ioutil.WriteFile(fileOut, output[i], 0644)
		if e != nil {
			color.Red(e.Error())
		}
	}
	return nil
}

// ParseContestProblem parse contest problem
func (c *Client) ParseContestProblem(contestID, probID, path string) (err error) {
	err = os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return
	}
	probURL := fmt.Sprintf("https://codeforces.com/contest/%v/problem/%v", contestID, probID)
	return c.ParseProblem(probURL, path)
}

// ParseContest parse for contest
func (c *Client) ParseContest(contestID, rootPath string) (err error) {
	color.Cyan("Try to parse contest %v to %v", contestID, rootPath)
	probs, err := c.StatisContest(contestID)
	if err != nil {
		return err
	}
	wg := sync.WaitGroup{}
	wg.Add(len(probs))
	for t := range probs {
		prob := probs[t]
		go func() {
			defer wg.Done()
			fmt.Printf("Parsing %v\n", prob.ID)
			probID := strings.ToLower(prob.ID)
			path := filepath.Join(rootPath, contestID, probID)
			err := c.ParseContestProblem(contestID, prob.ID, path)
			if err != nil {
				color.Red("%v: %v", prob.ID, err.Error())
			} else {
				color.Green("%v: Done!", prob.ID)
			}
		}()
	}
	wg.Wait()
	return nil
}
