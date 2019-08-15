package client

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"regexp"
	"strconv"
	"time"

	"github.com/fatih/color"
	ansi "github.com/k0kubun/go-ansi"
)

func findCountdown(body []byte) (int, error) {
	reg := regexp.MustCompile(`class=["']countdown["'][\s\S]*?(\d+):(\d+):(\d+)`)
	tmp := reg.FindSubmatch(body)
	if tmp == nil {
		return 0, errors.New("Cannot find any countdown")
	}
	h, _ := strconv.Atoi(string(tmp[1]))
	m, _ := strconv.Atoi(string(tmp[2]))
	s, _ := strconv.Atoi(string(tmp[3]))
	return h*60*60 + m*60 + s, nil
}

// RaceContest wait for contest starting
func (c *Client) RaceContest(contestID string) (err error) {
	color.Cyan(ToGym("Race for contest %v\n", contestID), contestID)

	URL := ToGym(fmt.Sprintf(c.Host+"/contest/%v/countdown", contestID), contestID)
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

	if !bytes.Contains(body, []byte(`Go!</a>`)) {
		count, err := findCountdown(body)
		if err != nil {
			return err
		}
		color.Green("Countdown: ")
		for count > 0 {
			h := count / 60 / 60
			m := count/60 - h*60
			s := count - h*60*60 - m*60
			fmt.Printf("%02d:%02d:%02d\n", h, m, s)
			ansi.CursorUp(1)
			count--
			time.Sleep(time.Second)
		}
		time.Sleep(900 * time.Millisecond)
	}

	return
}
