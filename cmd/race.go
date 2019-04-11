package cmd

import (
	"time"

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
	err = cln.RaceContest(contestID)
	if err != nil {
		return err
	}
	time.Sleep(1)
	return Parse(args)
}
