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

You should run "cf config login" and "cf config add" at first.

If you want to compete, the best command is "cf race 1111", where "1111" is the contest id.

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
  cf stand [<contest-id>]
  cf race <contest-id>

Examples:
  cf config login      Config your username and password.
  cf config add        Add a template.
  cf config default    Set default template.
  cf submit            Current path must be "<contest-id>/<problem-id>", cf will find which
                       file can be submitted.
  cf submit a.cpp
  cf submit 100 a
  cf submit 100 a a.cpp
  cf list              List problems' stats of current contest.
  cf list 1119         
  cf parse 100         Parse all problems of contest 100, including samples, into
                       "./100/<problem-id>".
  cf parse 100 a       Parse problem "a" of contest 100, including samples, into current path.
  cf gen               Generate default template into current path.
  cf gen cpp           Generate the template which's alias is "cpp" into current path.
  cf test              Compile a source which satisfy at least one template's suffix.
                       Then test all samples.
  cf watch             Watch the first 10 submissionso of current contest.
  cf open 1136 a       Use default web browser to open the page of contest 1136, problem a.
  cf open 1136         Use default web browser to open the page of contest 1136.
  cf stand             Use default web browser to open the standing page.
  cf race 1136         Count down before contest 1136 begins. Then it will run 'cf open 1136 a',
                       'cf open 1136 b', ..., 'cf open 1136 e', 'cf parse 1136' when the contest
                       begins.

Notes:
  <problem-id>         Could be "a" or "A", case-insensitive.
  <contest-id>         Should be a number, you could find it in codeforces contest url.
                       E.g. "1119" in "https://codeforces.com/contest/1119".
  <alias>              Template's alias.

File:
  cf will save some data in following files:

  "~/.cfconfig"        configuration file, including username, encrypted password, etc.
  "~/.cfsession"       session file, including cookies, username, etc.

  "~" is the home directory of current user in your system.

Template:
  You can insert some placeholders in your template code. When generate a code from the
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
  You could set "before_script" or "after_script" to empty string if you want,
  meaning not executing.
  You have to run your program in "script" with standard input/output (no need to
  redirect).

  You can insert some placeholders in your scripts. When execute a script,
  cf will replace all placeholders by following rules:

  $%path%$   Path to source file (Excluding $%full%$, e.g. "/home/xalanq/")
  $%full%$   Full name of source file (e.g. "a.cpp")
  $%file%$   Name of source file (Excluding suffix, e.g. "a")
  $%rand%$   Random string with 8 character (including "a-z" "0-9")

Options:
  -h --help
  --version`

	args, _ := docopt.Parse(usage, nil, true, "Codeforces Tool (cf) v0.3.3", false)
	color.Output = ansi.NewAnsiStdout()
	config.Init()
	err := cmd.Eval(args)
	if err != nil {
		color.Red(err.Error())
	}
	color.Unset()
}
