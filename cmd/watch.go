package cmd

import (
	"strings"

	"github.com/xalanq/cf-tool/client"
	"github.com/xalanq/cf-tool/config"
)

// Watch command
func Watch(args map[string]interface{}) error {
	contestID, err := getContestID(args)
	if err != nil {
		return err
	}
	problemID, _ := args["<problem-id>"].(string)
	problemID = strings.ToLower(problemID)
	cfg := config.New(config.ConfigPath)
	cln := client.New(config.SessionPath)
	n := 10
	if args["all"].(bool) {
		n = -1
	}
	_, err = cln.WatchSubmission(contestID, problemID, n, false)
	if err != nil {
		if err = loginAgain(cfg, cln, err); err == nil {
			_, err = cln.WatchSubmission(contestID, problemID, n, false)
		}
	}
	if err != nil {
		return err
	}
	return nil
}
