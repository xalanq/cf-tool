package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	ansi "github.com/k0kubun/go-ansi"
	"github.com/mitchellh/go-homedir"
	"github.com/xalanq/cf-tool/client"
	"github.com/xalanq/cf-tool/cmd"
	"github.com/xalanq/cf-tool/config"

	docopt "github.com/docopt/docopt-go"
)

const version = "v1.0.0"
const configPath = "~/.cf/config"
const sessionPath = "~/.cf/session"

func main() {
	usage := `Codeforces Tool $%version%$ (cf). https://github.com/xalanq/cf-tool

You should run "cf config" to configure your handle, password and code
templates at first.

If you want to compete, the best command is "cf race"

Usage:
  cf config
  cf submit [-f <file>] [<specifier>...]
  cf list [<specifier>...]
  cf parse [<specifier>...]
  cf gen [<alias>]
  cf test [<file>]
  cf customtest [-l <language-id>] <file> [<input-file>]
  cf watch [all] [<specifier>...]
  cf open [<specifier>...]
  cf stand [<specifier>...]
  cf sid [<specifier>...]
  cf race [<specifier>...]
  cf pull [ac] [<specifier>...]
  cf clone [ac] [<handle>]
  cf upgrade

Options:
  -h --help            Show this screen.
  --version            Show version.
  -f <file>, --file <file>, <file>
                       Path to file. E.g. "a.cpp", "./temp/a.cpp"
  -l <language-id>, --language-id <language-id>
                       Language ID. Choose "Add a template" option in "cf config"
                       to view the list of available language ID.
  <specifier>          Any useful text. E.g.
                       "https://codeforces.com/contest/100",
                       "https://codeforces.com/contest/180/problem/A",
                       "https://codeforces.com/group/Cw4JRyRGXR/contest/269760"
                       "1111A", "1111", "a", "Cw4JRyRGXR"
                       You can combine multiple specifiers to specify what you
                       want.
  <alias>              Template's alias. E.g. "cpp"
  ac                   The status of the submission is Accepted.

Examples:
  cf config            Configure the cf-tool.
  cf submit            cf will detect what you want to submit automatically.
  cf submit -f a.cpp
  cf submit https://codeforces.com/contest/100/A
  cf submit -f a.cpp 100A 
  cf submit -f a.cpp 100 a
  cf submit contest 100 a
  cf submit gym 100001 a
  cf list              List all problems' stats of a contest.
  cf list 1119
  cf parse 100         Fetch all problems' samples of contest 100 into
                       "{cf}/{contest}/100/<problem-id>".
  cf parse gym 100001a
                       Fetch samples of problem "a" of gym 100001 into
                       "{cf}/{gym}/100001/a".
  cf parse gym 100001
                       Fetch all problems' samples of gym 100001 into
                       "{cf}/{gym}/100001".
  cf parse             Fetch samples of current problem into current path.
  cf gen               Generate a code from default template.
  cf gen cpp           Generate a code from the template whose alias is "cpp"
                       into current path.
  cf test              Run the commands of a template in current path. Then
                       test all samples. If you want to add a new testcase,
                       create two files "inK.txt" and "ansK.txt" where K is
                       a string with 0~9.
  cf customtest a.py 31
  cf customtest a.py 31 input.txt
  cf watch             Watch the first 10 submissions of current contest.
  cf watch all         Watch all submissions of current contest.
  cf open 1136a        Use default web browser to open the page of contest
                       1136, problem a.
  cf open gym 100136   Use default web browser to open the page of gym
                       100136.
  cf stand             Use default web browser to open the standing page.
  cf sid 52531875      Use default web browser to open the submission
                       52531875's page.
  cf sid               Open the last submission's page.
  cf race 1136         If the contest 1136 has not started yet, it will
                       countdown. When the countdown ends, it will open all
                       problems' pages and parse samples.
  cf pull 100          Pull all problems' latest codes of contest 100 into
                       "./100/<problem-id>".
  cf pull 100 a        Pull the latest code of problem "a" of contest 100 into
                       "./100/<problem-id>".
  cf pull ac 100 a     Pull the "Accepted" or "Pretests passed" code of problem
                       "a" of contest 100.
  cf pull              Pull the latest codes of current problem into current
                       path.
  cf clone xalanq      Clone all codes of xalanq.
  cf upgrade           Upgrade the "cf" to the latest version from GitHub.

File:
  cf will save some data in some files:

  "~/.cf/config"        Configuration file, including templates, etc.
  "~/.cf/session"       Session file, including cookies, handle, password, etc.

  "~" is the home directory of current user in your system.

Template:
  You can insert some placeholders into your template code. When generate a code
  from the template, cf will replace all placeholders by following rules:

  $%U%$   Handle (e.g. xalanq)
  $%Y%$   Year   (e.g. 2019)
  $%M%$   Month  (e.g. 04)
  $%D%$   Day    (e.g. 09)
  $%h%$   Hour   (e.g. 08)
  $%m%$   Minute (e.g. 05)
  $%s%$   Second (e.g. 00)

Script in template:
  Template will run 3 scripts in sequence when you run "cf test":
    - before_script   (execute once)
    - script          (execute the number of samples times)
    - after_script    (execute once)
  You could set "before_script" or "after_script" to empty string, meaning
  not executing.
  You have to run your program in "script" with standard input/output (no
  need to redirect).

  You can insert some placeholders in your scripts. When execute a script,
  cf will replace all placeholders by following rules:

  $%path%$   Path to source file (Excluding $%full%$, e.g. "/home/xalanq/")
  $%full%$   Full name of source file (e.g. "a.cpp")
  $%file%$   Name of source file (Excluding suffix, e.g. "a")
  $%rand%$   Random string with 8 character (including "a-z" "0-9")`
	color.Output = ansi.NewAnsiStdout()

	usage = strings.Replace(usage, `$%version%$`, version, 1)
	opts, _ := docopt.ParseArgs(usage, os.Args[1:], fmt.Sprintf("Codeforces Tool (cf) %v", version))
	opts[`{version}`] = version

	cfgPath, _ := homedir.Expand(configPath)
	clnPath, _ := homedir.Expand(sessionPath)
	config.Init(cfgPath)
	client.Init(clnPath, config.Instance.Host, config.Instance.Proxy)

	err := cmd.Eval(opts)
	if err != nil {
		color.Red(err.Error())
	}
	color.Unset()
}
