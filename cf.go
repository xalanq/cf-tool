package main

import (
	"github.com/xalanq/codeforces/config"

	docopt "github.com/docopt/docopt-go"
	homedir "github.com/mitchellh/go-homedir"
)

func cmdConfig(args map[string]interface{}) {
	path, _ := homedir.Expand("~/.cfconfig")
	cfg := config.Load(path)
	if args["login"].(bool) {
		cfg.Login()
	} else if args["add"].(bool) {
		cfg.Add()
	}
	cfg.Save(path)
}

func cmdSubmit(args map[string]interface{}) {

}

func cmdParse(args map[string]interface{}) {

}

func main() {
	usage := `Codeforces Tool (cf). https://github.com/xalanq/codeforces

Usage:
  cf config [login | add]
  cf submit [<filename>] [--contest=<cid> --problem=<pid>]
  cf parse <cid>

Examples:
  cf config    config(store) username and password(encrypt)
  cf submit    submit file which parent dir is ./<cid>/<pid>/<valid file>
               if there are multiple avalible files. You have to choose one.
  cf parse 100
  cf submit a.cpp --contest=100 --problem=A 

Options:
  -h --help
  --version`

	args, _ := docopt.Parse(usage, nil, true, "Codeforces Tool (cf) v0.1.0", false)

	if args["config"].(bool) {
		cmdConfig(args)
	} else if args["submit"].(bool) {
		cmdSubmit(args)
	} else if args["parse"].(bool) {
		cmdParse(args)
	}
}
