package cmd

import (
	"os"
	"path/filepath"

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
	contestID, err := getContestID(args)
	if err != nil {
		return err
	}
	cln := client.New(config.SessionPath)
	work := func() error {
		if probID, ok := args["<problem-id>"].(string); ok {
			return cln.ParseContestProblem(contestID, probID, filepath.Join(currentPath, probID))
		}
		return cln.ParseContest(contestID, currentPath)
	}
	if err := work(); err != nil {
		if err = loginAgain(cfg, cln, err); err == nil {
			err = work()
		}
	}
	return err
}
