package cmd

import (
	"io/ioutil"
	"os"
	"strings"

	"github.com/xalanq/cf-tool/client"
	"github.com/xalanq/cf-tool/config"
	"github.com/xalanq/cf-tool/util"
)

// Submit command
func Submit(args map[string]interface{}) error {
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

func parseSubmitArgs(args map[string]interface{}) (string, string, string, error) {
	isFilename := func(str string) bool {
		if str == "" || util.IsUrl(str) {
			return false
		}
		if _, ok := os.Stat(str); strings.Contains(str, ".") || ok == nil {
			return true
		}
		return false
	}
	var newArgs = make(map[string]interface{})
	for key, value := range args {
		newArgs[key] = value
	}
	if _, ok := args["<filename>"].(string); !ok {
		if p, ok := args["<problem-id>"].(string); ok {
			if isFilename(p) {
				newArgs["<filename>"] = p
				newArgs["<problem-id>"] = nil
			}
		} else if c, ok := args["<url | contest-id>"].(string); ok {
			if isFilename(c) {
				newArgs["<filename>"] = c
				newArgs["<url | contest-id>"] = nil
			}
		}
	}
	parsedArgs, err := parseArgs(newArgs, map[string]bool{"<contest-id>": true, "<problem-id>": true, "<filename>": false})
	if err != nil {
		return "", "", "", err
	}
	return parsedArgs["<contest-id>"], parsedArgs["<problem-id>"], parsedArgs["<filename>"], nil
}
