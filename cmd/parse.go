package cmd

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/xalanq/cf-tool/client"
	"github.com/xalanq/cf-tool/config"
)

// Parse command
func Parse(args map[string]interface{}) error {
	currentPath, err := os.Getwd()
	if err != nil {
		return err
	}
	cfg := config.New(config.ConfigPath)
	cln := client.New(config.SessionPath)
	work := func() error {
		contestID := ""
		problemID := ""
		path := currentPath
		var ok bool
		if contestID, ok = args["<contest-id>"].(string); ok {
			if problemID, ok = args["<problem-id>"].(string); !ok {
				return cln.ParseContest(contestID, filepath.Join(currentPath, contestID), args["race"].(bool))
			}
			problemID = strings.ToLower(problemID)
			path = filepath.Join(currentPath, contestID, problemID)
		} else {
			contestID, err = getContestID(args)
			if err != nil {
				return err
			}
			problemID, err = getProblemID(args)
			if err != nil {
				return err
			}
			if problemID == contestID {
				return cln.ParseContest(contestID, currentPath, args["race"].(bool))
			}
		}
		samples, err := cln.ParseContestProblem(contestID, problemID, path)
		if err != nil {
			color.Red("Failed %v %v", contestID, problemID)
			return err
		}
		color.Green("Parsed %v %v with %v samples", contestID, problemID, samples)
		return nil
	}
	if err = work(); err != nil {
		if err = loginAgain(cfg, cln, err); err == nil {
			err = work()
		}
	}
	return err
}
