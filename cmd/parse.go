package cmd

import (
	"fmt"
	"os"
	"path/filepath"

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
	contestID, err := getContestID(args)
	if err != nil {
		return err
	}
	cln := client.New(config.SessionPath)
	work := func() error {
		if problemID, ok := args["<problem-id>"].(string); ok {
			samples, err := cln.ParseContestProblem(contestID, problemID, filepath.Join(currentPath, problemID))
			if err != nil {
				return fmt.Errorf("Failed %v %v", contestID, problemID)
			}
			color.Green("Parsed %v %v with %v samples", contestID, problemID, samples)
			return nil
		}
		return cln.ParseContest(contestID, currentPath)
	}
	if err = work(); err != nil {
		if err = loginAgain(cfg, cln, err); err == nil {
			err = work()
		}
	}
	return err
}
