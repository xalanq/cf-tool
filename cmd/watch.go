package cmd

import (
	"github.com/xalanq/cf-tool/client"
	"github.com/xalanq/cf-tool/config"
)

// Watch command
func Watch(args interface{}) error {
	parsedArgs, err := parseArgs(args, ParseRequirement{ContestID: true})
	if err != nil {
		return err
	}
	contestID, problemID := parsedArgs.ContestID, parsedArgs.ProblemID
	cfg := config.Instance
	cln := client.Instance
	n := 10
	if parsedArgs.All {
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
