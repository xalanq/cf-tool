package client

import (
	"encoding/json"
	"fmt"
	"html"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
)

type Handle struct {
	Handle string
	Color  string
}

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
	URL := ToGym(fmt.Sprintf(c.Host+"/contest/%v/problem/%v", contestID, problemID), contestID)
	samples, err = c.ParseProblem(URL, path)
	if err != nil {
		return
	}
	return
}

// ParseContest parse for contest
func (c *Client) ParseContest(contestID, rootPath string) (problems []StatisInfo, err error) {
	problems, err = c.StatisContest(contestID)
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

func (c *Client) findHandles(body []byte) (handles []Handle, err error) {
	handleRegex := regexp.MustCompile(`class="rated-user ([\s\S]*?)">([\s\S]*?)</a>`)
	legendaryRegex := regexp.MustCompile(`<span class="legendary-user-first-letter">([\s\S]*?)</span>([\s\S]*?)$`)

	handlesMatch := handleRegex.FindAllSubmatch(body, -1)
	if handlesMatch == nil {
		return nil, fmt.Errorf("cannot find handles")
	}

	for i := 0; i < len(handlesMatch); i += 1 {
		handle := Handle{Handle: string(handlesMatch[i][2]), Color: string(handlesMatch[i][1])[5:]}
		if handle.Color == "legendary" {
			legendaryMatch := legendaryRegex.FindAllSubmatch([]byte(handle.Handle), -1)
			if legendaryMatch == nil {
				return nil, fmt.Errorf("cannot find handle of legendary: %v", handle.Handle)
			}
			handle.Handle = string(legendaryMatch[0][1]) + string(legendaryMatch[0][2])
		}
		handles = append(handles, handle)
	}

	return
}

func (c *Client) ParseHandlesPage(page int) (handles []Handle, err error) {
	resp, err := c.client.Get(fmt.Sprintf("%s/problemset/standings/page/%d", c.Host, page))
	if err != nil {
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	return c.findHandles(body)
}

func (c *Client) ParseHandles() (result []Handle, err error) {
	threadNumber := 16

	ch := make(chan int, threadNumber)
	again := make(chan int, threadNumber)

	wg := sync.WaitGroup{}
	wg.Add(threadNumber + 1)
	mu := sync.Mutex{}

	count := 0
	//total := 1945
	total := 10

	go func() {
		for {
			s, ok := <-again
			if !ok {
				wg.Done()
				return
			}
			ch <- s
		}
	}()

	for gid := 0; gid < threadNumber; gid++ {
		go func() {
			for {
				page, ok := <-ch
				if !ok {
					wg.Done()
					return
				}
				handles, err := c.ParseHandlesPage(page)
				if err == nil {
					mu.Lock()
					count++
					color.Green(fmt.Sprintf(`%v/%v Saved`, count, total))
					result = append(result, handles...)
					mu.Unlock()
				} else {
					color.Red("%v", err.Error())
					err = fmt.Errorf("Too many requests")

					if err.Error() == "Too many requests" {
						mu.Lock()
						count++
						const WAIT int = 120
						color.Red(fmt.Sprintf(`%v/%v Error in %v: %v. Waiting for %v seconds to continue.`,
							count, total, page, err.Error(), WAIT))
						mu.Unlock()
						time.Sleep(time.Duration(WAIT) * time.Second)
						mu.Lock()
						count--
						mu.Unlock()
						again <- page
					}
				}
			}
		}()
	}

	for page := 1; page <= total; page++ {
		ch <- page
	}

	close(ch)
	close(again)
	wg.Wait()
	return
}

func (c *Client) SaveHandles(path string) error {
	handles, err := c.ParseHandles()
	if err != nil {
		return err
	}
	finalPath := filepath.Join(path, "data")
	if err := os.MkdirAll(finalPath, os.ModePerm); err != nil {
		return err
	}
	finalPath = filepath.Join(finalPath, "handles.json")
	b, err := json.Marshal(handles)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(finalPath, b, 0644)
}
