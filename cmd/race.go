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
	contestID, err := getContestID(args)
	if err != nil {
		return err
	}
	cln := client.New(config.SessionPath)
	if err = cln.RaceContest(contestID); err != nil {
		cfg := config.New(config.ConfigPath)
		if err = loginAgain(cfg, cln, err); err == nil {
			err = cln.RaceContest(contestID)
		}
		if err != nil {
			return err
		}
	}
	time.Sleep(1)
	open.Run(client.ToGym(fmt.Sprintf(cln.Host+"/contest/%v", contestID), contestID))
	open.Run(client.ToGym(fmt.Sprintf(cln.Host+"/contest/%v/problems", contestID), contestID))
	return Parse(args)
}
