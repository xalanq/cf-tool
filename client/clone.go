package client

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"cf-tool/cookiejar"
	"github.com/fatih/color"
)

type cloneData struct {
	contestID    string
	submissionID string
	path         string
	ext          string
}

// Clone all ac codes of all contests
func (c *Client) Clone(username, rootPath string, ac bool) (err error) {
	color.Cyan("Clone codes of %v, ac: %v", username, ac)

	jar, _ := cookiejar.New(nil)
	if username == c.Username {
		resp, err := c.client.Get(c.Host)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		if err = checkLogin(c.Username, body); err != nil {
			return err
		}
		jar = c.Jar.Copy()
	}

	resp, err := c.client.Get(fmt.Sprintf(c.Host+"/api/user.status?handle=%v", username))
	if err != nil {
		return
	}
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	var data map[string]interface{}
	if err = decoder.Decode(&data); err != nil {
		return
	}

	c.Jar = jar

	if status, ok := data["status"].(string); !ok || status != "OK" {
		return fmt.Errorf("Cannot get any submission")
	}
	submissions := data["result"].([]interface{})
	total := len(submissions)
	count := 0
	color.Cyan("Total submission: %v", total)

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
					s.contestID,
					s.submissionID,
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
					if username == c.Username {
						err = fmt.Errorf("Too many requests")
					}
					if err.Error() == "Too many requests" {
						mu.Lock()
						count++
						const WAIT int = 500
						color.Red(fmt.Sprintf(`%v/%v Error in %v|%v: %v. Waiting for %v seconds to continue.`,
							count, total, s.contestID, s.submissionID, err.Error(), WAIT))
						mu.Unlock()
						time.Sleep(time.Duration(WAIT) * time.Second)
						mu.Lock()
						count--
						mu.Unlock()
						again <- s
					} else {
						mu.Lock()
						count++
						color.Red(fmt.Sprintf(`%v/%v Error in %v|%v: %v`, count, total, s.contestID, s.submissionID, err.Error()))
						mu.Unlock()
					}
				}
			}
		}()
	}
	for _, _submission := range submissions {
		submission := _submission.(map[string]interface{})
		verdict := submission["verdict"].(string)
		contestID := fmt.Sprintf("%v", int64(submission["contestId"].(float64)))
		submissionID := fmt.Sprintf("%v", int64(submission["id"].(float64)))
		if ac && verdict != "OK" {
			mu.Lock()
			count++
			color.Green(fmt.Sprintf(`%v/%v Skip %v|%v: Not an accepted code`, count, total, contestID, submissionID))
			mu.Unlock()
			continue
		}
		lang := submission["programmingLanguage"].(string)
		ext, ok := LangsExt[lang]
		if !ok {
			mu.Lock()
			count++
			color.Red(fmt.Sprintf(`%v/%v Error in %v|%v: Language "%v" is not supported`, count, total, contestID, submissionID, lang))
			mu.Unlock()
			continue
		}
		problemID := strings.ToLower(submission["problem"].(map[string]interface{})["index"].(string))
		filename := submissionID
		if verdict != "OK" {
			testCount := int64(submission["passedTestCount"].(float64))
			filename = fmt.Sprintf("%v_%v_%v", submissionID, strings.ToLower(verdict), testCount)
		}
		which := "contest"
		if len(contestID) >= 6 {
			which = "gym"
		}
		path := filepath.Join(rootPath, username, which, contestID, problemID, filename)
		data := cloneData{contestID, submissionID, path, "." + ext}
		ch <- data
	}
	close(ch)
	close(again)
	wg.Wait()

	return nil
}
