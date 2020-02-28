package cmd

import (
	"errors"
	"github.com/docopt/docopt-go"
	"io/ioutil"
	"os"
	"strings"

	"github.com/xalanq/cf-tool/client"
	"github.com/xalanq/cf-tool/config"
	"github.com/xalanq/cf-tool/util"
)

// Submit command
func Submit(args interface{}) error {
	contestID, problemID, userFile, err := parseSubmitArgs(args)
	if err != nil {
		return err
	}
	cfg := config.Instance
	filename, index, err := getOneCode(userFile, cfg.Template)
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
	cln := client.Instance
	if err = cln.SubmitContest(contestID, problemID, lang, source); err != nil {
		if err = loginAgain(cfg, cln, err); err == nil {
			err = cln.SubmitContest(contestID, problemID, lang, source)
		}
	}

	return err
}

func parseSubmitArgs(args interface{}) (string, string, string, error) {
	opts, ok := args.(docopt.Opts)
	if !ok {
		return "", "", "", errors.New("args must be docopt.Opts type")
	}
	isFilename := func(str string) bool {
		if str == "" || util.IsUrl(str) {
			return false
		}
		if _, ok := os.Stat(str); strings.Contains(str, ".") || ok == nil {
			return true
		}
		return false
	}
	if _, ok := opts["<filename>"].(string); !ok {
		if p, ok := opts["<problem-id>"].(string); ok {
			if isFilename(p) {
				opts["<filename>"] = p
				opts["<problem-id>"] = nil
			}
		} else if c, ok := opts["<url | contest-id>"].(string); ok {
			if isFilename(c) {
				opts["<filename>"] = c
				opts["<url | contest-id>"] = nil
			}
		}
	}
	parsedArgs, err := parseArgs(opts, ParseRequirement{
		ContestID: true,
		ProblemID: true,
	})
	if err != nil {
		return "", "", "", err
	}
	return parsedArgs.ContestID, parsedArgs.ProblemID, parsedArgs.Filename, nil
}
