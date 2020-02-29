package client

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/xalanq/cf-tool/util"

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
func (c *Client) RaceContest(info Info) (err error) {
	color.Cyan("Race " + info.Hint())

	URL, err := info.ProblemSetURL(c.host)
	if err != nil {
		return
	}
	if info.ProblemType == "acmsguru" {
		return errors.New(ErrorNotSupportAcmsguru)
	}

	URL = URL + "/countdown"

	body, err := util.GetBody(c.client, URL)
	if err != nil {
		return
	}

	_, err = findHandle(body)
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
