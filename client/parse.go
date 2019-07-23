package client

import (
	"fmt"
	"html"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/fatih/color"
)

// ToGym if length of contestID >= 6, replace contest to gym
func ToGym(URL, contestID string) string {
	if len(contestID) >= 6 {
		URL = strings.Replace(URL, "contest", "gym", -1)
	}
	return URL
}

func findSample(body []byte) (input [][]byte, output [][]byte, err error) {
	irg := regexp.MustCompile(`class="input"[\s\S]*?<pre>([\s\S]*?)</pre>`)
	org := regexp.MustCompile(`class="output"[\s\S]*?<pre>([\s\S]*?)</pre>`)
	a := irg.FindAllSubmatch(body, -1)
	b := org.FindAllSubmatch(body, -1)
	if a == nil || b == nil || len(a) != len(b) {
		return nil, nil, fmt.Errorf("Cannot parse sample with input %v and output %v", len(a), len(b))
	}
	newline := regexp.MustCompile(`<[\s/br]+?>`)
	filter := func(src []byte) []byte {
		src = newline.ReplaceAll(src, []byte("\n"))
		s := html.UnescapeString(string(src))
		return []byte(strings.TrimSpace(s) + "\n")
	}
	for i := 0; i < len(a); i++ {
		input = append(input, filter(a[i][1]))
		output = append(output, filter(b[i][1]))
	}
	return
}

// ParseProblem parse problem to path
func (c *Client) ParseProblem(URL, path string) (samples int, err error) {
	client := &http.Client{Jar: c.Jar.Copy()}
	resp, err := client.Get(URL)
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
	return len(input), nil
}

// ParseContestProblem parse contest problem
func (c *Client) ParseContestProblem(contestID, problemID, path string) (samples int, err error) {
	err = os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return
	}
	URL := ToGym(fmt.Sprintf("https://codeforces.com/contest/%v/problem/%v", contestID, problemID), contestID)
	samples, err = c.ParseProblem(URL, path)
	if err != nil {
		return
	}
	return
}

// ParseContest parse for contest
func (c *Client) ParseContest(contestID, rootPath string) (err error) {
	problems, err := c.StatisContest(contestID)
	if err != nil {
		return
	}
	wg := sync.WaitGroup{}
	wg.Add(len(problems))
	mu := sync.Mutex{}
	for t := range problems {
		problem := problems[t]
		go func() {
			defer wg.Done()
			mu.Lock()
			fmt.Printf("Parsing %v %v\n", contestID, problem.ID)
			mu.Unlock()
			problemID := strings.ToLower(problem.ID)
			path := filepath.Join(rootPath, problemID)
			samples, err := c.ParseContestProblem(contestID, problem.ID, path)
			mu.Lock()
			if err != nil {
				color.Red("Failed %v %v. Error: %v", contestID, problem.ID, err.Error())
			} else {
				color.Green("Parsed %v %v with %v samples", contestID, problemID, samples)
			}
			mu.Unlock()
		}()
	}
	wg.Wait()
	return
}
