package cmd

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
	"strings"
	"time"
	"unicode"

	"github.com/fatih/color"
	ansi "github.com/k0kubun/go-ansi"
	"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/xalanq/cf-tool/config"
	"github.com/xalanq/cf-tool/util"
)

func splitCmd(s string) (res []string) {
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

func plain(raw []byte) string {
	buf := bufio.NewScanner(bytes.NewReader(raw))
	var b bytes.Buffer
	newline := []byte{'\n'}
	for buf.Scan() {
		b.Write(bytes.TrimSpace(buf.Bytes()))
		b.Write(newline)
	}
	return b.String()
}

func judge(sampleID, command string) error {
	inPath := fmt.Sprintf("in%v.txt", sampleID)
	ansPath := fmt.Sprintf("ans%v.txt", sampleID)
	input, err := os.Open(inPath)
	if err != nil {
		return err
	}
	var o bytes.Buffer
	output := io.Writer(&o)

	cmds := splitCmd(command)

	st := time.Now()
	cmd := exec.Command(cmds[0], cmds[1:]...)
	cmd.Stdin = input
	cmd.Stdout = output
	cmd.Stderr = os.Stderr
	cmd.Run()
	dt := time.Now().Sub(st)

	b, err := ioutil.ReadFile(ansPath)
	if err != nil {
		b = []byte{}
	}
	ans := plain(b)
	out := plain(o.Bytes())

	state := ""
	diff := ""
	if out == ans {
		state = color.New(color.FgGreen).Sprintf("Passed #%v", sampleID)
	} else {
		state = color.New(color.FgRed).Sprintf("Failed #%v", sampleID)
		dmp := diffmatchpatch.New()
		d := dmp.DiffMain(out, ans, true)
		diff = dmp.DiffPrettyText(d) + "\n"
	}
	ansi.Printf("%v .... %.3fs\n%v", state, dt.Seconds(), diff)
	return nil
}

// Test command
func Test(args map[string]interface{}) error {
	cfg := config.New(config.ConfigPath)
	if len(cfg.Template) == 0 {
		return errors.New("You have to add at least one code template by `cf config add`")
	}
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
			err := judge(i, s)
			if err != nil {
				color.Red(err.Error())
			}
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
