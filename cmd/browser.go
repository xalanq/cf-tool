package cmd

import (
	"fmt"
	"strconv"

	"github.com/skratchdot/open-golang/open"
	"github.com/xalanq/cf-tool/client"
	"github.com/xalanq/cf-tool/config"
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
		return open.Run(client.ToGym(fmt.Sprintf("https://codeforces.com/contest/%v", contestID), contestID))
	}
	return open.Run(client.ToGym(fmt.Sprintf("https://codeforces.com/contest/%v/problem/%v", contestID, problemID), contestID))
}

// Stand command
func Stand(args map[string]interface{}) error {
	contestID, err := getContestID(args)
	if err != nil {
		return err
	}
	return open.Run(client.ToGym(fmt.Sprintf("https://codeforces.com/contest/%v/standings", contestID), contestID))
}

// Sid command
func Sid(args map[string]interface{}) error {
	contestID := ""
	submissionID := ""
	if args["<submission-id>"] == nil {
		cln := client.New(config.SessionPath)
		if cln.LastSubmission != nil {
			contestID = cln.LastSubmission.ContestID
			submissionID = cln.LastSubmission.SubmissionID
		} else {
			return fmt.Errorf(`You have not submitted any problem yet`)
		}
	} else {
		var err error
		contestID, err = getContestID(args)
		if err != nil {
			return err
		}
		submissionID, _ = args["<submission-id>"].(string)
		if _, err = strconv.Atoi(submissionID); err != nil {
			return err
		}
	}
	return open.Run(client.ToGym(fmt.Sprintf("https://codeforces.com/contest/%v/submission/%v", contestID, submissionID), contestID))
}
