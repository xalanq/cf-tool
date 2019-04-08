package client

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
)

// StatisInfo statis infomation
type StatisInfo struct {
	ID     string
	Name   string
	IO     string
	Limit  string
	Passed string
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
	rep, _ := regexp.Compile(`<[\s\S]+?>`)
	ton, _ := regexp.Compile(`<\s+`)
	rmv, _ := regexp.Compile(`<+`)
	tmp = append(tmp, []int{len(body), 0})
	st := tmp[0][0]
	for i := 1; i < len(tmp); i++ {
		b := rep.ReplaceAll(body[st:tmp[i][0]], []byte("<"))
		st = tmp[i][0]
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
			})
		}
	}
	return ret, nil
}

// Statis get prblem statis
func (c *Client) Statis(contestID string) (probs []StatisInfo, err error) {
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

	block, err := findStatisBlock(body)
	if err != nil {
		return
	}
	return findProblems(block)
}
