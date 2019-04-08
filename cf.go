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
	"github.com/xalanq/codeforces/client"
	"github.com/xalanq/codeforces/config"
)

var configPath = "~/.cfconfig"
var sessionPath = "~/.cfsession"

func cmdConfig(args map[string]interface{}) {
	e := func() error {
		cfg := config.New(configPath)
		if args["login"].(bool) {
			return cfg.Login(sessionPath)
		} else if args["add"].(bool) {
			return cfg.Add()
		}
		return nil
	}()
	if e != nil {
		fmt.Println(e.Error())
	}
}

func cmdSubmit(args map[string]interface{}) {
	e := func() error {
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
	}()
	if e != nil {
		fmt.Println(e.Error())
	}
}

func cmdParse(args map[string]interface{}) {

}

func main() {
	usage := `Codeforces Tool (cf). https://github.com/xalanq/codeforces

Usage:
  cf config [login | add]
  cf submit [<filename>]
  cf submit [(<contest-id> <problem-id>)] [<filename>]
  cf parse <contest-id>

Examples:
  cf config   Config(store) username and password(encrypt)
  cf submit   Current path must be <contest-id>/<problem-id>/<file.[suffix]>.
              If there are multiple files which satisfy above condition, you
              have to choose one.
  cf submit 100 a
  cf submit 100 a a.cp
  cf parse 100

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

	if args["config"].(bool) {
		cmdConfig(args)
	} else if args["submit"].(bool) {
		cmdSubmit(args)
	} else if args["parse"].(bool) {
		cmdParse(args)
	}
}
