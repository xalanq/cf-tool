package cmd

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
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
	source := ""
	ext := ""
	if cfg.GenAfterParse {
		if len(cfg.Template) == 0 {
			return errors.New("You have to add at least one code template by `cf config add`")
		}
		path := cfg.Template[cfg.Default].Path
		ext = filepath.Ext(path)
		if source, err = readTemplateSource(path, cfg); err != nil {
			return err
		}
	}
	cln := client.New(config.SessionPath)
	parseContest := func(contestID, rootPath string, race bool) error {
		problems, err := cln.ParseContest(contestID, rootPath, race)
		if err == nil && cfg.GenAfterParse {
			for _, problem := range problems {
				problemID := strings.ToLower(problem.ID)
				path := filepath.Join(rootPath, problemID)
				gen(source, path, ext)
			}
		}
		return err
	}
	work := func() error {
		contestID := ""
		problemID := ""
		path := currentPath
		var ok bool
		if contestID, ok = args["<contest-id>"].(string); ok {
			if problemID, ok = args["<problem-id>"].(string); !ok {
				return parseContest(contestID, filepath.Join(currentPath, contestID), args["race"].(bool))
			}
			problemID = strings.ToLower(problemID)
			path = filepath.Join(currentPath, contestID, problemID)
		} else {
			contestID, err = getContestID(args)
			if err != nil {
				return err
			}
			problemID, err = getProblemID(args)
			if err != nil {
				return err
			}
			if problemID == contestID {
				return parseContest(contestID, currentPath, args["race"].(bool))
			}
		}
		samples, err := cln.ParseContestProblem(contestID, problemID, path)
		if err != nil {
			color.Red("Failed %v %v", contestID, problemID)
			return err
		}
		color.Green("Parsed %v %v with %v samples", contestID, problemID, samples)
		if cfg.GenAfterParse {
			gen(source, path, ext)
		}
		return nil
	}
	if err = work(); err != nil {
		if err = loginAgain(cfg, cln, err); err == nil {
			err = work()
		}
	}
	return err
}
