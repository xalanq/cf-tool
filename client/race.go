package client

import (
	"errors"
	"fmt"
	"io/ioutil"
	"regexp"
	"strconv"
	"time"

	"github.com/fatih/color"
	ansi "github.com/k0kubun/go-ansi"
	"github.com/skratchdot/open-golang/open"
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

func raceContest(contestID string) (err error) {
	for _, problemID := range []string{"A", "B", "C", "D", "E"} {
		open.Run(ToGym(fmt.Sprintf("https://codeforces.com/contest/%v/problem/%v", contestID, problemID), contestID))
	}
	return nil
}

// RaceContest wait for contest starting
func (c *Client) RaceContest(contestID string) (err error) {
	color.Cyan(ToGym("Race for contest %v\n", contestID), contestID)

	URL := ToGym(fmt.Sprintf("https://codeforces.com/contest/%v", contestID), contestID)
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

	_, err = findStatisBlock(body)
	if err != nil {
		count, err := findCountdown(body)
		if err != nil {
			return err
		}
		color.Green("Count down: ")
		count--
		time.Sleep(900 * time.Millisecond)
		for count > 0 {
			time.Sleep(time.Second)
			count--
			h := count / 60 / 60
			m := count/60 - h*60
			s := count - h*60*60 - m*60
			fmt.Printf("%02d:%02d:%02d\n", h, m, s)
			ansi.CursorUp(1)
		}
	}
	return raceContest(contestID)
}
