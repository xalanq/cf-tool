package client

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
)

// StatisInfo statis information
type StatisInfo struct {
	ID     string
	Name   string
	IO     string
	Limit  string
	Passed string
	State  string
}

func findStatisBlock(body []byte) ([]byte, error) {
	reg, _ := regexp.Compile(`class="problems"[\s\S]+?</tr>([\s\S]+?)</table>`)
	tmp := reg.FindSubmatch(body)
	if tmp == nil {
		return nil, errors.New("Cannot find any problem statis")
	}
	return tmp[1], nil
}

func findProblems(body []byte) ([]StatisInfo, error) {
	reg, _ := regexp.Compile(`<tr[\s\S]*?>`)
	tmp := reg.FindAllIndex(body, -1)
	if tmp == nil {
		return nil, errors.New("Cannot find any problem")
	}
	ret := []StatisInfo{}
	scr, _ := regexp.Compile(`<script[\s\S]*?>[\s\S]*?</script>`)
	cls, _ := regexp.Compile(`class="(.+?)"`)
	rep, _ := regexp.Compile(`<[\s\S]+?>`)
	ton, _ := regexp.Compile(`<\s+`)
	rmv, _ := regexp.Compile(`<+`)
	tmp = append(tmp, []int{len(body), 0})
	for i := 1; i < len(tmp); i++ {
		state := ""
		if x := cls.FindSubmatch(body[tmp[i-1][0]:tmp[i-1][1]]); x != nil {
			state = string(x[1])
		}
		b := scr.ReplaceAll(body[tmp[i-1][0]:tmp[i][0]], []byte{})
		b = rep.ReplaceAll(b, []byte("<"))
		b = ton.ReplaceAll(b, []byte("<"))
		b = rmv.ReplaceAll(b, []byte("<"))
		data := strings.Split(string(b), "<")
		tot := []string{}
		for j := 0; j < len(data); j++ {
			s := strings.TrimSpace(data[j])
			if s != "" {
				tot = append(tot, s)
			}
		}
		if len(tot) >= 5 {
			ret = append(ret, StatisInfo{
				tot[0], tot[1], tot[2], tot[3],
				strings.ReplaceAll(tot[4], "&nbsp;x", ""),
				state,
			})
		}
	}
	return ret, nil
}

// StatisContest get contest problems statis
func (c *Client) StatisContest(contestID string) (probs []StatisInfo, err error) {
	fmt.Printf("Get statis in contest %v\n", contestID)
	statisURL := fmt.Sprintf("https://codeforces.com/contest/%v", contestID)

	client := &http.Client{Jar: c.Jar}
	resp, err := client.Get(statisURL)
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

	block, err := findStatisBlock(body)
	if err != nil {
		return
	}
	return findProblems(block)
}
