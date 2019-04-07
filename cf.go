package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	docopt "github.com/docopt/docopt-go"
	homedir "github.com/mitchellh/go-homedir"
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
		// cfg := config.New(configPath)
		// cln := client.New(sessionPath)
		contest := ""
		problem := ""
		lang := "0"
		filename, ok := args["<filename>"].(string)
		if ok {
			if tmp, ok := args["<contest-id>"].(string); ok {
				contest = tmp
				problem = args["<problem-id>"].(string)
			} else {
				currentPath, err := os.Getwd()
				if err != nil {
					return err
				}
				fmt.Println(currentPath)
			}
		} else {

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
		fmt.Printf("contest: %v, problem: %v, lang: %v, filename: %v\n%v", contest, problem, lang, filename, source)

		return nil
		// return cln.SubmitContest(contest, problem, lang, source)
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
  cf submit [<filename>] [(<contest-id> <problem-id>)]
  cf parse <contest-id>

Examples:
  cf config   Config(store) username and password(encrypt)
  cf submit   Current path must be <contest-id>/<problem-id>/<file.[suffix]>.
              If there are multiple files which satisfy above condition, you
              have to choose one.
  cf submit a.cpp 100 a
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
	fmt.Println(args)

	if args["config"].(bool) {
		cmdConfig(args)
	} else if args["submit"].(bool) {
		cmdSubmit(args)
	} else if args["parse"].(bool) {
		cmdParse(args)
	}
}
