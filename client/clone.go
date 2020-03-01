package client

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/xalanq/cf-tool/util"
)

type cloneData struct {
	url  string
	path string
	ext  string
}

// Clone all ac codes of all contests
func (c *Client) Clone(handle, rootPath string, ac bool) (err error) {
	color.Cyan("Clone all codes of %v. Only Accepted: %v", handle, ac)

	if handle == c.Handle {
		body, err := util.GetBody(c.client, c.host)
		if err != nil {
			return err
		}

		if _, err = findHandle(body); err != nil {
			return err
		}
	}

	data, err := util.GetJSONBody(c.client, fmt.Sprintf(c.host+"/api/user.status?handle=%v", handle))
	if err != nil {
		return
	}

	if status, ok := data["status"].(string); !ok || status != "OK" {
		return fmt.Errorf("Cannot get any submission")
	}
	submissions := data["result"].([]interface{})
	total := len(submissions)
	count := 0
	color.Cyan("Total submissions: %v", total)

	threadNumber := 16
	ch := make(chan cloneData, threadNumber)
	again := make(chan cloneData, threadNumber)
	wg := sync.WaitGroup{}
	wg.Add(threadNumber + 1)
	mu := sync.Mutex{}

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
				s, ok := <-ch
				if !ok {
					wg.Done()
					return
				}
				filename, err := c.PullCode(
					s.url,
					s.path,
					s.ext,
					false,
				)
				if err == nil {
					mu.Lock()
					count++
					color.Green(fmt.Sprintf(`%v/%v Saved %v`, count, total, filename))
					mu.Unlock()
				} else {
					if err.Error() == ErrorSkip {
						mu.Lock()
						count++
						color.Yellow(fmt.Sprintf(`%v/%v Exist %v: Skip.`, count, total, s.url))
						mu.Unlock()
					} else if err.Error() == ErrorTooManyRequest {
						mu.Lock()
						count++
						const WAIT int = 500
						color.Red(fmt.Sprintf(`%v/%v Error %v: %v. Waiting for %v seconds to continue.`,
							count, total, s.url, err.Error(), WAIT))
						mu.Unlock()
						time.Sleep(time.Duration(WAIT) * time.Second)
						mu.Lock()
						count--
						mu.Unlock()
						again <- s
					} else {
						mu.Lock()
						count++
						color.Red(fmt.Sprintf(`%v/%v Error %v: %v`, count, total, s.url, err.Error()))
						mu.Unlock()
					}
				}
			}
		}()
	}
	for _, _submission := range submissions {
		func() {
			defer func() {
				if r := recover(); r != nil {
					color.Red("Error: %v", r)
					color.Red("%v", _submission)
				}
			}()
			submission := _submission.(map[string]interface{})
			verdict := submission["verdict"].(string)
			lang := submission["programmingLanguage"].(string)
			contestID := ""
			if v, ok := submission["contestId"].(float64); ok {
				contestID = fmt.Sprintf("%v", int64(v))
			} else {
				contestID = "99999"
			}
			submissionID := fmt.Sprintf("%v", int64(submission["id"].(float64)))
			problemID := strings.ToLower(submission["problem"].(map[string]interface{})["index"].(string))
			info := Info{ProblemType: "contest", ContestID: contestID, ProblemID: problemID, SubmissionID: submissionID}
			if contestID == "99999" {
				info.ProblemType = "acmsguru"
			} else if len(contestID) >= 6 {
				info.ProblemType = "gym"
			}
			if ac && verdict != "OK" {
				mu.Lock()
				count++
				color.Green(fmt.Sprintf(`%v/%v Skip %v: Not an accepted code`, count, total, info.Hint()))
				mu.Unlock()
				return
			}
			ext, ok := LangsExt[lang]
			if !ok {
				mu.Lock()
				count++
				color.Red(fmt.Sprintf(`%v/%v Error %v: Language "%v" is not supported`, count, total, info.Hint(), lang))
				mu.Unlock()
				return
			}
			filename := submissionID
			if verdict != "OK" {
				testCount := int64(submission["passedTestCount"].(float64))
				filename = fmt.Sprintf("%v_%v_%v", submissionID, strings.ToLower(verdict), testCount)
			}
			/*
			info.RootPath = filepath.Join(rootPath, handle, info.ProblemType)
			// NOTE this path scheme is incompatible with other commands. "handle/cf/contest/..." would be compatible, however.
			*/
			URL, _ := info.SubmissionURL(c.host)
			data := cloneData{URL, filepath.Join(info.Path(), filename), "." + ext}
			ch <- data
		}()
	}
	close(ch)
	close(again)
	wg.Wait()

	return nil
}
