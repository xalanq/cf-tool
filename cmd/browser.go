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
	parsedArgs, err := parseArgs(args, map[string]bool{"<contest-id>": true, "<problem-id>": false})
	if err != nil {
		return err
	}
	contestID, problemID := parsedArgs["<contest-id>"], parsedArgs["<problem-id>"]
	if problemID == "" {
		return open.Run(client.ToGym(fmt.Sprintf(config.Instance.Host+"/contest/%v", contestID), contestID))
	}
	return open.Run(client.ToGym(fmt.Sprintf(config.Instance.Host+"/contest/%v/problem/%v", contestID, problemID), contestID))
}

// Stand command
func Stand(args map[string]interface{}) error {
	parsedArgs, err := parseArgs(args, map[string]bool{"<contest-id>": true})
	if err != nil {
		return err
	}
	contestID := parsedArgs["<contest-id>"]
	return open.Run(client.ToGym(fmt.Sprintf(config.Instance.Host+"/contest/%v/standings", contestID), contestID))
}

// Sid command
func Sid(args map[string]interface{}) error {
	parsedArgs, err := parseArgs(args, map[string]bool{"<contest-id>": false, "<submission-id>": false})
	contestID, submissionID := parsedArgs["<contest-id>"], parsedArgs["<submission-id>"]
	cfg := config.Instance
	cln := client.Instance
	if submissionID == "" {
		if cln.LastSubmission != nil {
			contestID = cln.LastSubmission.ContestID
			submissionID = cln.LastSubmission.SubmissionID
		} else {
			return fmt.Errorf(`You have not submitted any problem yet`)
		}
	} else {
		if err != nil {
			return err
		}
		if _, err = strconv.Atoi(submissionID); err != nil {
			return err
		}
	}
	return open.Run(client.ToGym(fmt.Sprintf(cfg.Host+"/contest/%v/submission/%v", contestID, submissionID), contestID))
}
