package cmd

import (
	"path/filepath"

	"github.com/xalanq/cf-tool/client"
	"github.com/xalanq/cf-tool/config"
)

// Pull command
func Pull(args interface{}) error {
	cfg := config.Instance
	cln := client.Instance
	var err error
	work := func() error {
		parsedArgs, err := parseArgs(args, ParseRequirement{ContestID: true})
		if err != nil {
			return err
		}
		contestID, problemID := parsedArgs.ContestID, parsedArgs.ProblemID
		path := filepath.Join(parsedArgs.ContestRootPath, contestID, problemID)
		return cln.PullContest(contestID, problemID, path, parsedArgs.Accepted)
	}
	if err = work(); err != nil {
		if err = loginAgain(cfg, cln, err); err == nil {
			err = work()
		}
	}
	return err
}
