package cmd

import (
	"fmt"
	"strconv"

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

// Stand command
func Stand(args map[string]interface{}) error {
	contestID, err := getContestID(args)
	if err != nil {
		return err
	}
	return open.Run(fmt.Sprintf("https://codeforces.com/contest/%v/standings", contestID))
}

// Sid command
func Sid(args map[string]interface{}) error {
	contestID, err := getContestID(args)
	if err != nil {
		return err
	}
	c, _ := args["<submission-id>"].(string)
	if _, err := strconv.Atoi(c); err == nil {
		return open.Run(fmt.Sprintf("https://codeforces.com/contest/%v/submission/%v", contestID, c))
	}
	return fmt.Errorf(`Submission ID should be a number instead of "%v"`, c)
}
