package cmd

import (
	"fmt"
	"strconv"

	"cf-tool/client"
	"cf-tool/config"
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
		return open.Run(client.ToGym(fmt.Sprintf(client.New(config.SessionPath).Host+"/contest/%v", contestID), contestID))
	}
	return open.Run(client.ToGym(fmt.Sprintf(client.New(config.SessionPath).Host+"/contest/%v/problem/%v", contestID, problemID), contestID))
}

// Stand command
func Stand(args map[string]interface{}) error {
	contestID, err := getContestID(args)
	if err != nil {
		return err
	}
	return open.Run(client.ToGym(fmt.Sprintf(client.New(config.SessionPath).Host+"/contest/%v/standings", contestID), contestID))
}

// Sid command
func Sid(args map[string]interface{}) error {
	contestID := ""
	submissionID := ""
	cln := client.New(config.SessionPath)
	if args["<submission-id>"] == nil {
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
	return open.Run(client.ToGym(fmt.Sprintf(cln.Host+"/contest/%v/submission/%v", contestID, submissionID), contestID))
}
