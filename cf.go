package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	docopt "github.com/docopt/docopt-go"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/xalanq/cf-tool/client"
	"github.com/xalanq/cf-tool/config"
)

var configPath = "~/.cfconfig"
var sessionPath = "~/.cfsession"

func cmdConfig(args map[string]interface{}) error {
	cfg := config.New(configPath)
	if args["login"].(bool) {
		return cfg.Login(sessionPath)
	} else if args["add"].(bool) {
		return cfg.Add()
	}
	return nil
}

func cmdSubmit(args map[string]interface{}) error {
	cfg := config.New(configPath)
	cln := client.New(sessionPath)
	contest := ""
	problem := ""
	lang := ""
	filename, ok := args["<filename>"].(string)
	ava := []string{}
	mp := make(map[string]int)
	for i, temp := range cfg.Template {
		for _, suffix := range temp.Suffix {
			mp["."+suffix] = i
		}
	}
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
		return errors.New("Cannot find any supported file to submit\nYou can add the suffix with `cf config add`")
	}
	if len(ava) > 1 {
		fmt.Println("There are multiple files can be submitted.")
		for i, name := range ava {
			fmt.Printf("%3v: %v\n", i, name)
		}
		fmt.Print("Please choose one(index): ")
		for {
			var index string
			_, err := fmt.Scanln(&index)
			if err == nil {
				i, err := strconv.Atoi(index)
				if err == nil && i >= 0 && i < len(ava) {
					filename = ava[i]
					i = mp[filepath.Ext(filename)]
					lang = cfg.Template[i].Lang
					break
				}
			}
			fmt.Println("Invalid index! Please try again: ")
		}
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

func cmdList(args map[string]interface{}) error {
	contest := ""
	if tmp, ok := args["<contest-id>"].(string); ok {
		contest = tmp
	} else {
		currentPath, err := os.Getwd()
		if err != nil {
			return err
		}
		contest = filepath.Base(filepath.Dir(currentPath))
	}
	if _, err := strconv.Atoi(contest); err != nil {
		return fmt.Errorf(`Contest should be a number instead of "%v"`, contest)
	}
	cfg := config.New(configPath)
	cln := client.New(sessionPath)
	for T := 1; T <= 3; T++ {
		probs, err := cln.Statis(contest)
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
		maxLen := make([]int, 5)
		for _, prob := range probs {
			if len := len(prob.ID); len > maxLen[0] {
				maxLen[0] = len
			}
			if len := len(prob.Name); len > maxLen[1] {
				maxLen[1] = len
			}
			if len := len(prob.Passed); len > maxLen[2] {
				maxLen[2] = len
			}
			if len := len(prob.Limit); len > maxLen[3] {
				maxLen[3] = len
			}
			if len := len(prob.IO); len > maxLen[4] {
				maxLen[4] = len
			}
		}
		format := "  "
		for _, i := range maxLen {
			format += "%-" + fmt.Sprintf("%v", i+2) + "v"
		}
		format += "\n"
		fmt.Printf(format, "#", "Name", "AC", "Limit", "IO")
		for _, prob := range probs {
			fmt.Printf(format, prob.ID, prob.Name, prob.Passed, prob.Limit, prob.IO)
		}
		break
	}
	return nil
}

func cmdParse(args map[string]interface{}) error {
	return nil
}

func main() {
	usage := `Codeforces Tool (cf). https://github.com/codeforces/codeforces

Usage:
  cf config [login | add]
  cf submit [<filename>]
  cf submit [(<contest-id> <problem-id>)] [<filename>]
  cf list [<contest-id>]
  cf parse <contest-id>

Examples:
  cf config      Config(store) username and password(encrypt)
  cf submit      Current path must be <contest-id>/<problem-id>/<file.[suffix]>.
                 If there are multiple files which satisfy above condition, you
                 have to choose one.
  cf list        List current contest or <contest-id> problems' infomation
  cf parse 100   Generate Round, include sample
  cf submit 100 a
  cf submit 100 a a.cp

Notes:
  <problem-id>   could be "a" or "A", case-insensitive
  <contest-id>   should be a number, you could find it in codeforces contest url.
                 E.g. 1119 in https://codeforces.com/contest/1119

Options:
  -h --help
  --version`

	args, _ := docopt.Parse(usage, nil, true, "Codeforces Tool (cf) v0.1.0", false)
	configPath, _ = homedir.Expand(configPath)
	sessionPath, _ = homedir.Expand(sessionPath)

	e := func() error {
		if args["config"].(bool) {
			return cmdConfig(args)
		} else if args["submit"].(bool) {
			return cmdSubmit(args)
		} else if args["list"].(bool) {
			return cmdList(args)
		} else if args["parse"].(bool) {
			return cmdParse(args)
		}
		return nil
	}()
	if e != nil {
		fmt.Println(e.Error())
	}
}
