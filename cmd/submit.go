package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"cf-tool/client"
	"cf-tool/config"
)

// Submit command
func Submit(args map[string]interface{}) error {
	contestID, err := getContestID(args)
	if err != nil {
		return err
	}
	problemID, err := getProblemID(args)
	if err != nil {
		return err
	}
	if problemID == contestID {
		return fmt.Errorf("contestID: %v, problemID: %v is not valid", contestID, problemID)
	}
	cfg := config.New(config.ConfigPath)
	filename, index, err := getOneCode(args, cfg.Template)
	if err != nil {
		return err
	}
	template := cfg.Template[index]

	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	bytes, err := ioutil.ReadAll(file)

	if err != nil {
		return err
	}

	source := string(bytes)
	problemID = strings.ToUpper(problemID)
	lang := template.Lang
	cln := client.New(config.SessionPath)
	if err = cln.SubmitContest(contestID, problemID, lang, source); err != nil {
		if err = loginAgain(cfg, cln, err); err == nil {
			err = cln.SubmitContest(contestID, problemID, lang, source)
		}
	}

	return err
}
