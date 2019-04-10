package cmd

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	ansi "github.com/k0kubun/go-ansi"
	"github.com/xalanq/cf-tool/client"
	"github.com/xalanq/cf-tool/config"
)

// List command
func List(args map[string]interface{}) error {
	contest, err := getContestID(args)
	if err != nil {
		return err
	}
	cfg := config.New(config.ConfigPath)
	cln := client.New(config.SessionPath)
	probs, err := cln.StatisContest(contest)
	if err != nil {
		if err = loginAgain(cfg, cln, err); err == nil {
			probs, err = cln.StatisContest(contest)
		}
	}
	if err != nil {
		return err
	}
	maxLen := make([]int, 5)
	for _, prob := range probs {
		if len := len(prob.ID); len > maxLen[0] {
			maxLen[0] = len
		}
		if len := len(prob.Name); len > maxLen[1] {
			maxLen[1] = len
		}
		if len := len(prob.Passed); len > maxLen[2] {
			maxLen[2] = len
		}
		if len := len(prob.Limit); len > maxLen[3] {
			maxLen[3] = len
		}
		if len := len(prob.IO); len > maxLen[4] {
			maxLen[4] = len
		}
	}
	format := "  "
	for _, i := range maxLen {
		format += "%-" + fmt.Sprintf("%v", i+2) + "v"
	}
	format += "\n"
	fmt.Printf(format, "#", "Name", "AC", "Limit", "IO")
	for _, prob := range probs {
		info := fmt.Sprintf(format, prob.ID, prob.Name, prob.Passed, prob.Limit, prob.IO)
		if strings.Contains(prob.State, "accepted") {
			info = color.New(color.BgGreen).Sprint(info)
		} else if strings.Contains(prob.State, "rejected") {
			info = color.New(color.BgRed).Sprint(info)
		}
		ansi.Print(info)
	}
	return nil
}
