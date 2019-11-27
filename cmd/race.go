package cmd

import (
	"fmt"
	"time"

	"github.com/skratchdot/open-golang/open"
	"github.com/xalanq/cf-tool/client"
	"github.com/xalanq/cf-tool/config"
)

// Race command
func Race(args map[string]interface{}) error {
	parsedArgs, err := parseArgs(args, map[string]bool{"<contest-id>": true})
	if err != nil {
		return err
	}
	contestID := parsedArgs["<contest-id>"]
	cfg := config.Instance
	cln := client.Instance
	if err = cln.RaceContest(contestID); err != nil {
		if err = loginAgain(cfg, cln, err); err == nil {
			err = cln.RaceContest(contestID)
		}
		if err != nil {
			return err
		}
	}
	time.Sleep(1)
	open.Run(client.ToGym(fmt.Sprintf(cfg.Host+"/contest/%v", contestID), contestID))
	open.Run(client.ToGym(fmt.Sprintf(cfg.Host+"/contest/%v/problems", contestID), contestID))
	return _Parse(contestID, "", parsedArgs["contestRootPath"])
}
