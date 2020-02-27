package client

import (
	"errors"
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"

	"github.com/fatih/color"
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
	reg := regexp.MustCompile(`class="problems"[\s\S]+?</tr>([\s\S]+?)</table>`)
	tmp := reg.FindSubmatch(body)
	if tmp == nil {
		return nil, errors.New("Cannot find any problem statis")
	}
	return tmp[1], nil
}

func findProblems(body []byte) ([]StatisInfo, error) {
	reg := regexp.MustCompile(`<tr[\s\S]*?>`)
	tmp := reg.FindAllIndex(body, -1)
	if tmp == nil {
		return nil, errors.New("Cannot find any problem")
	}
	ret := []StatisInfo{}
	scr := regexp.MustCompile(`<script[\s\S]*?>[\s\S]*?</script>`)
	cls := regexp.MustCompile(`class="(.+?)"`)
	rep := regexp.MustCompile(`<[\s\S]+?>`)
	ton := regexp.MustCompile(`<\s+`)
	rmv := regexp.MustCompile(`<+`)
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
			tot[4] = strings.ReplaceAll(tot[4], "x", "")
			tot[4] = strings.ReplaceAll(tot[4], "&nbsp;", "")
			if tot[4] == "" {
				tot[4] = "0"
			}
			ret = append(ret, StatisInfo{
				tot[0], tot[1], tot[2], tot[3],
				tot[4], state,
			})
		}
	}
	return ret, nil
}

// StatisContest get contest problems statis
func (c *Client) StatisContest(contestID string) (problems []StatisInfo, err error) {
	color.Cyan(ToGym("Get statis in contest %v\n", contestID), contestID)

	URL := ToGym(fmt.Sprintf(c.host+"/contest/%v", contestID), contestID)
	resp, err := c.client.Get(URL)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	_, err = findHandle(body)
	if err != nil {
		return
	}

	block, err := findStatisBlock(body)
	if err != nil {
		return
	}
	return findProblems(block)
}
