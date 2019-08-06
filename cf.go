package main

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	ansi "github.com/k0kubun/go-ansi"
	"github.com/xalanq/cf-tool/cmd"
	"github.com/xalanq/cf-tool/config"

	docopt "github.com/docopt/docopt-go"
)

const version = "v0.7.2"

func main() {
	usage := `Codeforces Tool $%version%$ (cf). https://github.com/xalanq/cf-tool

You should run "cf config login" and "cf config add" at first.

If you want to compete, the best command is "cf race 1111" where "1111" is the contest id.

Usage:
  cf config (login | add | del | default)
  cf submit [(<contest-id> <problem-id>)] [<filename>]
  cf list [<contest-id>]
  cf parse [<contest-id>] [<problem-id>]
  cf gen [<alias>]
  cf test [<filename>]
  cf watch [all] [<contest-id>]
  cf open [<contest-id>] [<problem-id>]
  cf stand [<contest-id>]
  cf sid [<submission-id>] [<contest-id>]
  cf race <contest-id>
  cf pull [ac] [<contest-id>] [<problem-id>]
  cf clone [ac] <username>
  cf upgrade

Examples:
  cf config login      Config your username and password.
  cf config add        Add a template.
  cf config del        Remove a template.
  cf config default    Set default template.
  cf submit            If current path is "<contest-id>/<problem-id>", cf will find the
                       code which can be submitted. Then submit to <contest-id> <problem-id>.
  cf submit a.cpp
  cf submit 100 a
  cf submit 100 a a.cpp
  cf list              List all problems' stats of a contest.
  cf list 1119
  cf parse 100         Fetch all problems' samples of contest 100 into "./100/<problem-id>".
  cf parse 100 a       Fetch samples of problem "a" of contest 100 into "./100/a".
  cf parse             Fetch samples of current problem into current path.
  cf gen               Generate a code from default template.
  cf gen cpp           Generate a code from the template whose alias is "cpp" into current path.
  cf test              Run the commands of a template in current path. Then test all samples.
                       If you want to add a new testcase, create two files "inK.txt" and "ansK.txt"
                       where K is a string with 0~9.
  cf watch             Watch the first 10 submissions of current contest.
  cf watch all         Watch all submissions of current contest.
  cf open 1136 a       Use default web browser to open the page of contest 1136, problem a.
  cf open 1136         Use default web browser to open the page of contest 1136.
  cf stand             Use default web browser to open the standing page.
  cf sid 52531875      Use default web browser to open the submission 52531875's page.
  cf sid               Open the last submission's page.
  cf race 1136         If the contest 1136 has not started yet, it will countdown. When the
                       countdown ends, it will open all problems' pages and parse samples.
  cf pull 100          Pull all problems' latest codes of contest 100 into "./100/<problem-id>".
  cf pull 100 a        Pull the latest code of problem "a" of contest 100 into "./100/<problem-id>".
  cf pull ac 100 a     Pull the "Accepted" or "Pretests passed" code of problem "a" of contest 100.
  cf pull              Pull the latest code of current problem into current path.
  cf clone xalanq      Clone all codes of xalanq.
  cf upgrade           Upgrade the "cf" to the latest version from GitHub.

Notes:
  <problem-id>         "a" or "A", case-insensitive.
  <contest-id>         A number. You can find it in codeforces contest url. E.g. "1119" in
                       "https://codeforces.com/contest/1119".
  <alias>              Template's alias.

File:
  cf will save some data in some files:

  "~/.cfconfig"        Configuration file, including username, encrypted password, etc.
  "~/.cfsession"       Session file, including cookies, username, etc.

  "~" is the home directory of current user in your system.

Template:
  You can insert some placeholders into your template code. When generate a code from the
  template, cf will replace all placeholders by following rules:

  $%U%$   Username
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
  You could set "before_script" or "after_script" to empty string, meaning not executing.
  You have to run your program in "script" with standard input/output (no need to redirect).

  You can insert some placeholders in your scripts. When execute a script,
  cf will replace all placeholders by following rules:

  $%path%$   Path to source file (Excluding $%full%$, e.g. "/home/xalanq/")
  $%full%$   Full name of source file (e.g. "a.cpp")
  $%file%$   Name of source file (Excluding suffix, e.g. "a")
  $%rand%$   Random string with 8 character (including "a-z" "0-9")

Options:
  -h --help
  --version`
	usage = strings.Replace(usage, `$%version%$`, version, 1)

	args, _ := docopt.Parse(usage, nil, true, fmt.Sprintf("Codeforces Tool (cf) %v", version), false)
	args[`{version}`] = version
	color.Output = ansi.NewAnsiStdout()
	config.Init()
	err := cmd.Eval(args)
	if err != nil {
		color.Red(err.Error())
	}
	color.Unset()
}
