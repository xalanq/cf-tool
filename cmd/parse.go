package cmd

import (
	"fmt"
	"os"

	"github.com/xalanq/cf-tool/client"
	"github.com/xalanq/cf-tool/config"
)

// Parse command
func Parse(args map[string]interface{}) error {
	currentPath, err := os.Getwd()
	if err != nil {
		return err
	}
	cfg := config.New(config.ConfigPath)
	cln := client.New(config.SessionPath)
	contestID := args["<contest-id>"].(string)
	for T := 1; T <= 3; T++ {
		if probID, ok := args["<problem-id>"].(string); ok {
			err = cln.ParseContestProblem(contestID, probID, currentPath)
		} else {
			err = cln.ParseContest(contestID, currentPath)
		}
		if err != nil {
			if err.Error() == client.ErrorNotLogged {
				fmt.Printf("Not logged. %v try to re-login\n", T)
				password, err := cfg.DecryptPassword()
				if err != nil {
					return err
				}
				cln.Login(cfg.Username, password)
				continue
			}
			return err
		}
		break
	}
	return nil
}
