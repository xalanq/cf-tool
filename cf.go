package main

import (
	"github.com/fatih/color"
	ansi "github.com/k0kubun/go-ansi"
	"github.com/xalanq/cf-tool/cmd"
	"github.com/xalanq/cf-tool/config"

	docopt "github.com/docopt/docopt-go"
)

func main() {
	usage := `Codeforces Tool (cf). https://github.com/xalanq/cf-tool

File:
  cf will save some data in following files:

  "~/.cfconfig"   config file, including username, encrypted password, etc.
  "~/.cfsession"  session file, including cookies, username, etc.

  "~" is the homedir in your system

Usage:
  cf config (login | add | default)
  cf submit [<filename>]
  cf submit [(<contest-id> <problem-id>)] [<filename>]
  cf list [<contest-id>]
  cf parse <contest-id> [<problem-id>]
  cf gen [<alias>]
  cf test [<filename>]
  cf watch [<contest-id>]
  cf open [<contest-id>] [<problem-id>]
  cf hack [<contest-id>]

Examples:
  cf config login      Config username and password(encrypt).
  cf config add        Add template.
  cf config default    Set default template.
  cf submit            Current path must be <contest-id>/<problem-id>/<file.[suffix]>.
                       If there are multiple files which satisfy above condition, you
                       have to choose one.
  cf submit a.cpp
  cf submit 100 a
  cf submit 100 a a.cpp
  cf list              List current contest or <contest-id> problems' information.
  cf parse 100         Parse contest 100, all problems, including samples,
                       into ./100/<problem-id>.
  cf parse 100 a       Parse contest 100, problem a, including samples,
                       into current path
  cf gen               Generate default template in current path (name as current path).
  cf gen cpp           Generate template which alias is cpp in current path (same above).
  cf test              Compile the source with build config first. Then test all samples.
                       If there are multiple files, you have to choose one.
  cf watch             Watch the first 10 submissions
  cf open              Open page with default browser
  cf hack              Open standing page with default browser

Notes:
  <problem-id>         Could be "a" or "A", case-insensitive.
  <contest-id>         Should be a number, you could find it in codeforces contest url.
                       E.g. 1119 in https://codeforces.com/contest/1119.
  <alias>              Template's alias.

Template:
  You can insert some placeholders in your template code. When generate a code from a
  template, cf will replace all placeholders by following rules:

  $%U%$   Username
  $%Y%$   Year   (e.g. 2019)
  $%M%$   Month  (e.g. 04)
  $%D%$   Day    (e.g. 09)
  $%h%$   Hour   (e.g. 08)
  $%m%$   Minute (e.g. 05)
  $%s%$   Second (e.g. 00)

Command:
  Execution order is:
    - before_script   (execute once)
    - script          (execute number of samples times)
    - after_script    (execute once)
  You can set one of before_script and after_script to empty string,
  meaning not executing. You have to run your program in script(standard input/output).

  You can insert some placeholders in your commands. When execute these commands,
  cf will replace all placeholders by following rules:

  $%path%$   Path of test file (Excluding $%full%$, e.g. /home/xalanq/)
  $%full%$   Full name of test file (e.g. a.cpp)
  $%file%$   Name of testing file (Excluding suffix, e.g. a)
  $%rand%$   Random string with 8 character (including a-z 0-9)

Options:
  -h --help
  --version`

	args, _ := docopt.Parse(usage, nil, true, "Codeforces Tool (cf) v0.2.1", false)
	color.Output = ansi.NewAnsiStdout()
	config.Init()
	err := cmd.Eval(args)
	if err != nil {
		color.Red(err.Error())
	}
	color.Unset()
}
