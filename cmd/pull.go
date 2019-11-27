package cmd

import (
	"path/filepath"

	"github.com/xalanq/cf-tool/client"
	"github.com/xalanq/cf-tool/config"
)

// Pull command
func Pull(args map[string]interface{}) error {
	cfg := config.Instance
	cln := client.Instance
	ac := args["ac"].(bool)
	var err error
	work := func() error {
		parsedArgs, err := parseArgs(args, map[string]bool{"<contest-id>": true, "<problem-id>": false})
		if err != nil {
			return err
		}
		contestID, problemID := parsedArgs["<contest-id>"], parsedArgs["<problem-id>"]
		path := filepath.Join(parsedArgs["contestRootPath"], contestID, problemID)
		return cln.PullContest(contestID, problemID, path, ac)
	}
	if err = work(); err != nil {
		if err = loginAgain(cfg, cln, err); err == nil {
			err = work()
		}
	}
	return err
}
