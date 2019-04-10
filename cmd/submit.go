package cmd

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/xalanq/cf-tool/client"
	"github.com/xalanq/cf-tool/config"
	"github.com/xalanq/cf-tool/util"
)

// Submit command
func Submit(args map[string]interface{}) error {
	cfg := config.New(config.ConfigPath)
	cln := client.New(config.SessionPath)
	contest := ""
	problem := ""
	lang := ""
	ava := []string{}
	mp := make(map[string]int)
	for i, temp := range cfg.Template {
		for _, suffix := range temp.Suffix {
			mp["."+suffix] = i
		}
	}
	filename, ok := args["<filename>"].(string)
	if !ok {
		currentPath, err := os.Getwd()
		if err != nil {
			return err
		}
		paths, err := ioutil.ReadDir(currentPath)
		if err != nil {
			return err
		}
		for _, path := range paths {
			name := path.Name()
			ext := filepath.Ext(name)
			if _, ok := mp[ext]; ok {
				ava = append(ava, name)
			}
		}
	} else {
		ext := filepath.Ext(filename)
		if _, ok := mp[ext]; ok {
			ava = append(ava, filename)
		}
	}
	if len(ava) < 1 {
		return errors.New("Cannot find any supported file to submit\nYou can add some suffixes by `cf config add`")
	}
	if len(ava) > 1 {
		color.Cyan("There are multiple files can be submitted.")
		for i, name := range ava {
			fmt.Printf("%3v: %v\n", i, name)
		}
		i := util.ChooseIndex(len(ava))
		filename = ava[i]
		i = mp[filepath.Ext(filename)]
		lang = cfg.Template[i].Lang
	} else {
		filename = ava[0]
		i := mp[filepath.Ext(filename)]
		lang = cfg.Template[i].Lang
	}
	if tmp, ok := args["<contest-id>"].(string); ok {
		contest = tmp
		problem = args["<problem-id>"].(string)
	} else {
		currentPath, err := os.Getwd()
		if err != nil {
			return err
		}
		problem = filepath.Base(currentPath)
		contest = filepath.Base(filepath.Dir(currentPath))
	}

	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	bytes, err := ioutil.ReadAll(file)

	if err != nil {
		return err
	}

	problem = strings.ToUpper(problem)
	source := string(bytes)
	if _, err := strconv.Atoi(contest); err != nil {
		return fmt.Errorf(`Contest should be a number instead of "%v"`, contest)
	}

	for T := 1; T <= 3; T++ {
		err = cln.SubmitContest(contest, problem, lang, source)
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
