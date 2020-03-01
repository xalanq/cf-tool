package cmd

import (
	"io"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/xalanq/cf-tool/client"
)

// Contests command
func Contests() error {
	cln := client.Instance
	contests, err := cln.GetContests()
	if err != nil {
		if err = loginAgain(cln, err); err == nil {
			contests, err = cln.GetContests()
		}
	}
	if err != nil {
		return err
	}
	output := io.Writer(os.Stdout)
	table := tablewriter.NewWriter(output)
	table.SetHeader([]string{"ID", "Name", "Start", "Length", "State", "Registration"})
	table.SetAlignment(tablewriter.ALIGN_CENTER)
	table.SetRowLine(true)
	table.SetRowSeparator("─")
	table.SetColumnSeparator("│")
	table.SetCenterSeparator("┼")
	table.SetAutoWrapText(false)

	colorCell := func(s string, doColor bool) string {
		if doColor {
			return colorText(s, color.GreenString)
		}
		return s
	}
	for _, contest := range contests {
		ok := contest.Registered
		table.Append([]string{
			colorCell(contest.ID, ok),
			colorCell(contest.Name, ok),
			colorCell(contest.Start, ok),
			colorCell(contest.Length, ok),
			colorCell(contest.State, ok),
			colorCell(contest.Registration, ok),
		})

	}
	table.Render()
	return nil
}

func colorText(text string, f func(s string, a ...interface{}) string) string {
	lines := strings.Split(text, "\n")
	out := ""
	for i := 0; i < len(lines); i++ {
		out += f(lines[i])
		if i != len(lines)-1 {
			out += "\n"
		}
	}
	return out
}
