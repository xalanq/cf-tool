package cmd

import (
	"fmt"

	"github.com/skratchdot/open-golang/open"
)

// Open command
func Open(args map[string]interface{}) error {
	contestID, err := getContestID(args)
	if err != nil {
		return err
	}
	problemID, err := getProblemID(args)
	if err != nil {
		return err
	}
	if problemID == contestID {
		return open.Run(fmt.Sprintf("https://codeforces.com/contest/%v", contestID))
	}
	return open.Run(fmt.Sprintf("https://codeforces.com/contest/%v/problem/%v", contestID, problemID))
}

// Hack command
func Hack(args map[string]interface{}) error {
	contestID, err := getContestID(args)
	if err != nil {
		return err
	}
	return open.Run(fmt.Sprintf("https://codeforces.com/contest/%v/standings", contestID))
}
