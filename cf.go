package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/fatih/color"
	ansi "github.com/k0kubun/go-ansi"

	docopt "github.com/docopt/docopt-go"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/xalanq/cf-tool/client"
	"github.com/xalanq/cf-tool/config"
	"github.com/xalanq/cf-tool/util"
)

var configPath = "~/.cfconfig"
var sessionPath = "~/.cfsession"

func cmdConfig(args map[string]interface{}) error {
	cfg := config.New(configPath)
	if args["login"].(bool) {
		return cfg.Login(sessionPath)
	} else if args["add"].(bool) {
		return cfg.Add()
	} else if args["default"].(bool) {
		return cfg.SetDefault()
	}
	return nil
}

func cmdSubmit(args map[string]interface{}) error {
	cfg := config.New(configPath)
	cln := client.New(sessionPath)
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
		return errors.New("Cannot find any supported file to submit\nYou can add the suffix with `cf config add`")
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

func cmdList(args map[string]interface{}) error {
	contest := ""
	if tmp, ok := args["<contest-id>"].(string); ok {
		contest = tmp
	} else {
		currentPath, err := os.Getwd()
		if err != nil {
			return err
		}
		if _, err := strconv.Atoi(filepath.Base(currentPath)); err == nil {
			contest = filepath.Base(currentPath)
		} else {
			contest = filepath.Base(filepath.Dir(currentPath))
		}
	}
	if _, err := strconv.Atoi(contest); err != nil {
		return fmt.Errorf(`Contest should be a number instead of "%v"`, contest)
	}
	cfg := config.New(configPath)
	cln := client.New(sessionPath)
	for T := 1; T <= 3; T++ {
		probs, err := cln.StatisContest(contest)
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
			info := fmt.Sprintf(format, prob.ID, prob.Name, prob.Passed, prob.Limit, prob.IO)
			if strings.Contains(prob.State, "accepted") {
				info = color.New(color.BgGreen).Sprint(info)
			} else if strings.Contains(prob.State, "rejected") {
				info = color.New(color.BgRed).Sprint(info)
			}
			ansi.Print(info)
		}
		break
	}
	return nil
}

func cmdParse(args map[string]interface{}) error {
	currentPath, err := os.Getwd()
	if err != nil {
		return err
	}
	cfg := config.New(configPath)
	cln := client.New(sessionPath)
	contestID := args["<contest-id>"].(string)
	for T := 1; T <= 3; T++ {
		if probID, ok := args["<problem-id>"].(string); ok {
			err = cln.ParseContestProblem(contestID, probID, currentPath)
		} else {
			err = cln.ParseContest(contestID, currentPath)
		}
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

func cmdGen(args map[string]interface{}) error {
	currentPath, err := os.Getwd()
	if err != nil {
		return err
	}
	savePath := filepath.Join(currentPath, filepath.Base(currentPath))
	path := ""
	cfg := config.New(configPath)
	if alias, ok := args["<alias>"].(string); ok {
		templates := cfg.Alias(alias)
		if len(templates) < 1 {
			return fmt.Errorf("Cannot find any template with alias %v", alias)
		} else if len(templates) == 1 {
			path = templates[0].Path
		} else {
			fmt.Printf("There are multiple templates with alias %v\n", alias)
			for i, template := range templates {
				fmt.Printf(`%3v: "%v"`, i, template.Path)
				fmt.Println()
			}
			i := util.ChooseIndex(len(templates))
			path = templates[i].Path
		}
	} else {
		if cfg.Default < 0 || cfg.Default >= len(cfg.Template) {
			return fmt.Errorf("Invalid default value %v in config file", cfg.Default)
		}
		path = cfg.Template[cfg.Default].Path
	}
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	now := time.Now()
	source := string(b)
	source = strings.ReplaceAll(source, "$%U%$", cfg.Username)
	source = strings.ReplaceAll(source, "$%Y%$", fmt.Sprintf("%v", now.Year()))
	source = strings.ReplaceAll(source, "$%M%$", fmt.Sprintf("%02v", int(now.Month())))
	source = strings.ReplaceAll(source, "$%D%$", fmt.Sprintf("%02v", now.Day()))
	source = strings.ReplaceAll(source, "$%h%$", fmt.Sprintf("%02v", now.Hour()))
	source = strings.ReplaceAll(source, "$%m%$", fmt.Sprintf("%02v", now.Minute()))
	source = strings.ReplaceAll(source, "$%s%$", fmt.Sprintf("%02v", now.Second()))
	ext := filepath.Ext(path)
	tmpPath := savePath + ext
	_, err = os.Stat(tmpPath)
	for i := 1; err == nil; i++ {
		nxtPath := fmt.Sprintf("%v%v%v", savePath, i, ext)
		fmt.Printf("%v is existed. Rename to %v\n", filepath.Base(tmpPath), filepath.Base(nxtPath))
		tmpPath = nxtPath
		_, err = os.Stat(tmpPath)
	}
	savePath = tmpPath
	err = ioutil.WriteFile(savePath, []byte(source), 0644)
	if err == nil {
		color.Green("Generated! See %v", filepath.Base(savePath))
	}
	return err
}

func cmdTest(args map[string]interface{}) error {
	cfg := config.New(configPath)
	var template config.CodeTemplate
	ava := []string{}
	mp := make(map[string]int)
	samples := []string{}
	for i, temp := range cfg.Template {
		for _, suffix := range temp.Suffix {
			mp["."+suffix] = i
		}
	}
	filename, ok := args["<filename>"].(string)
	currentPath, err := os.Getwd()
	if err != nil {
		return err
	}
	paths, err := ioutil.ReadDir(currentPath)
	if err != nil {
		return err
	}
	sampleReg, _ := regexp.Compile(`in(\d+).txt`)
	for _, path := range paths {
		name := path.Name()
		tmp := sampleReg.FindSubmatch([]byte(name))
		if tmp != nil {
			idx := string(tmp[1])
			ans := fmt.Sprintf("ans%v.txt", idx)
			if _, err := os.Stat(ans); err == nil {
				samples = append(samples, idx)
			}
		}
		if !ok {
			ext := filepath.Ext(name)
			if _, ok := mp[ext]; ok {
				ava = append(ava, name)
			}
		}
	}
	if ok {
		ext := filepath.Ext(filename)
		if _, ok := mp[ext]; ok {
			ava = append(ava, filename)
		}
	}
	if len(ava) < 1 {
		return errors.New("Cannot find any supported file to test\nYou can add the suffix with `cf config add`")
	}
	if len(ava) > 1 {
		color.Cyan("There are multiple files can be tested.")
		for i, name := range ava {
			fmt.Printf("%3v: %v\n", i, name)
		}
		i := util.ChooseIndex(len(ava))
		filename = ava[i]
		template = cfg.Template[mp[filepath.Ext(filename)]]
	} else {
		filename = ava[0]
		template = cfg.Template[mp[filepath.Ext(filename)]]
	}
	path, full := filepath.Split(filename)
	ext := filepath.Ext(filename)
	file := full[:len(full)-len(ext)]
	rand := util.RandString(8)
	filter := func(cmd string) string {
		cmd = strings.ReplaceAll(cmd, "$%rand%$", rand)
		cmd = strings.ReplaceAll(cmd, "$%path%$", path)
		cmd = strings.ReplaceAll(cmd, "$%full%$", full)
		cmd = strings.ReplaceAll(cmd, "$%file%$", file)
		return cmd
	}
	splitCmd := func(s string) (res []string) {
		// https://github.com/vrischmann/shlex/blob/master/shlex.go
		var buf bytes.Buffer
		insideQuotes := false
		for _, r := range s {
			switch {
			case unicode.IsSpace(r) && !insideQuotes:
				if buf.Len() > 0 {
					res = append(res, buf.String())
					buf.Reset()
				}
			case r == '"' || r == '\'':
				if insideQuotes {
					res = append(res, buf.String())
					buf.Reset()
					insideQuotes = false
					continue
				}
				insideQuotes = true
			default:
				buf.WriteRune(r)
			}
		}
		if buf.Len() > 0 {
			res = append(res, buf.String())
		}
		return
	}
	plain := func(raw []byte) string {
		buf := bufio.NewScanner(bytes.NewReader(raw))
		var b bytes.Buffer
		newline := []byte{'\n'}
		for buf.Scan() {
			b.Write(bytes.TrimSpace(buf.Bytes()))
			b.Write(newline)
		}
		return b.String()
	}
	if len(samples) == 0 {
		color.Red("There is no sample data")
		return nil
	}
	if s := filter(template.BeforeScript); len(s) > 0 {
		fmt.Println(s)
		cmds := splitCmd(s)
		cmd := exec.Command(cmds[0], cmds[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Run()
	}
	if s := filter(template.Script); len(s) > 0 {
		for _, i := range samples {
			inFile := fmt.Sprintf("in%v.txt", i)
			ansFile := fmt.Sprintf("ans%v.txt", i)
			input, err := os.Open(inFile)
			if err != nil {
				color.Red(err.Error())
				continue
			}
			var o bytes.Buffer
			output := io.Writer(&o)

			cmds := splitCmd(s)

			st := time.Now()
			cmd := exec.Command(cmds[0], cmds[1:]...)
			cmd.Stdin = input
			cmd.Stdout = output
			cmd.Stderr = os.Stderr
			cmd.Run()
			dt := time.Now().Sub(st)

			b, err := ioutil.ReadFile(ansFile)
			if err != nil {
				b = []byte{}
			}
			ans := plain(b)
			out := plain(o.Bytes())

			state := ""
			diff := ""
			if out == ans {
				state = color.New(color.FgGreen).Sprintf("Passed #%v", i)
			} else {
				state = color.New(color.FgRed).Sprintf("Failed #%v", i)
				dmp := diffmatchpatch.New()
				d := dmp.DiffMain(out, ans, true)
				diff = dmp.DiffPrettyText(d) + "\n"
			}
			ansi.Printf("%v .... %.3f sec\n%v", state, dt.Seconds(), diff)
		}
	} else {
		color.Red("Invalid script command. Please check config file")
		return nil
	}
	if s := filter(template.AfterScript); len(s) > 0 {
		fmt.Println(s)
		cmds := splitCmd(s)
		cmd := exec.Command(cmds[0], cmds[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Run()
	}
	return nil
}

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

Examples:
  cf config login      Config username and password(encrypt).
  cf config add        Add template.
  cf config default    Set default template.
  cf submit            Current path must be <contest-id>/<problem-id>/<file.[suffix]>.
                       If there are multiple files which satisfy above condition, you
                       have to choose one.
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
  meaning not executing.

  You can insert some placeholders in your commands. When execute these commands,
  cf will replace all placeholders by following rules:

  $%path%$   Path of test file (Excluding $%full%$, e.g. /home/xalanq/)
  $%full%$   Full name of test file (e.g. a.cpp)
  $%file%$   Name of testing file (Excluding suffix, e.g. a)
  $%rand%$   Random string with 8 character (including a-z 0-9)

Options:
  -h --help
  --version`

	args, _ := docopt.Parse(usage, nil, true, "Codeforces Tool (cf) v0.1.0", false)
	color.Output = ansi.NewAnsiStdout()
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
		} else if args["gen"].(bool) {
			return cmdGen(args)
		} else if args["test"].(bool) {
			return cmdTest(args)
		}
		return nil
	}()
	if e != nil {
		fmt.Println(e.Error())
	}
}
