package cmd

import (
	"fmt"
	"time"

	"github.com/skratchdot/open-golang/open"
	"github.com/xalanq/cf-tool/client"
	"github.com/xalanq/cf-tool/config"
)

// Race command
func Race(args interface{}) error {
	parsedArgs, err := parseArgs(args, ParseRequirement{ContestID: true})
	if err != nil {
		return err
	}
	contestID := parsedArgs.ContestID
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
	return _parse(contestID, "", parsedArgs.ContestRootPath)
}
