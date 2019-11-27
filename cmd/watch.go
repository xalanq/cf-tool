package cmd

import (
	"github.com/xalanq/cf-tool/client"
	"github.com/xalanq/cf-tool/config"
)

// Watch command
func Watch(args map[string]interface{}) error {
	parsedArgs, err := parseArgs(args, map[string]bool{"<contest-id>": true, "<problem-id>": false})
	if err != nil {
		return err
	}
	contestID, problemID := parsedArgs["<contest-id>"], parsedArgs["<problem-id>"]
	cfg := config.Instance
	cln := client.Instance
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
