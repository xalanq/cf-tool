package cmd

import (
	"fmt"

	"github.com/xalanq/cf-tool/client"
	"github.com/xalanq/cf-tool/config"
)

// Watch command
func Watch(args map[string]interface{}) error {
	contest, err := getContestID(args)
	if err != nil {
		return err
	}
	cfg := config.New(config.ConfigPath)
	cln := client.New(config.SessionPath)
	URL := fmt.Sprintf("https://codeforces.com/contest/%v/my", contest)
	err = cln.WatchSubmission(URL, 10, false)
	if err != nil {
		if err = loginAgain(cfg, cln, err); err == nil {
			err = cln.WatchSubmission(URL, 10, false)
		}
	}
	if err != nil {
		return err
	}
	return nil
}
