package cmd

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/xalanq/cf-tool/client"
	"github.com/xalanq/cf-tool/config"
)

// Pull command
func Pull(args map[string]interface{}) error {
	currentPath, err := os.Getwd()
	if err != nil {
		return err
	}
	cfg := config.New(config.ConfigPath)
	cln := client.New(config.SessionPath)
	ac := args["ac"].(bool)
	work := func() error {
		contestID := ""
		problemID := ""
		path := currentPath
		ok := false
		if contestID, ok = args["<contest-id>"].(string); ok {
			if problemID, ok = args["<problem-id>"].(string); !ok {
				return cln.PullContest(contestID, "", filepath.Join(currentPath, contestID), ac)
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
				return cln.PullContest(contestID, "", currentPath, ac)
			}
		}
		return cln.PullContest(contestID, problemID, path, ac)
	}
	if err = work(); err != nil {
		if err = loginAgain(cfg, cln, err); err == nil {
			err = work()
		}
	}
	return err
}
