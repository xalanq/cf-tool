package cmd

import (
	"bufio"
	"bytes"
	"io"
	"strings"

	"github.com/fatih/color"
	ansi "github.com/k0kubun/go-ansi"
	"github.com/olekukonko/tablewriter"
	"github.com/xalanq/cf-tool/client"
	"github.com/xalanq/cf-tool/config"
)

// List command
func List(args map[string]interface{}) error {
	contestID, err := getContestID(args)
	if err != nil {
		return err
	}
	cfg := config.Instance
	cln := client.Instance
	problems, err := cln.StatisContest(contestID)
	if err != nil {
		if err = loginAgain(cfg, cln, err); err == nil {
			problems, err = cln.StatisContest(contestID)
		}
	}
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	output := io.Writer(&buf)
	table := tablewriter.NewWriter(output)
	table.SetHeader([]string{"#", "problem", "passed", "limit", "IO"})
	table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	table.SetAlignment(tablewriter.ALIGN_CENTER)
	table.SetCenterSeparator("|")
	table.SetAutoWrapText(false)
	for _, prob := range problems {
		table.Append([]string{
			prob.ID,
			prob.Name,
			prob.Passed,
			prob.Limit,
			prob.IO,
		})
	}
	table.Render()

	scanner := bufio.NewScanner(io.Reader(&buf))
	for i := -2; scanner.Scan(); i++ {
		line := scanner.Text()
		if i >= 0 {
			if strings.Contains(problems[i].State, "accepted") {
				line = color.New(color.BgGreen).Sprint(line)
			} else if strings.Contains(problems[i].State, "rejected") {
				line = color.New(color.BgRed).Sprint(line)
			}
		}
		ansi.Println(line)
	}

	return nil
}
