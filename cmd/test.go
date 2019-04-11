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

	samples := getSampleID()
	if len(samples) == 0 {
		color.Red("There is no sample data")
		return nil
	}

	filename, index, err := getOneCode(args, cfg.Template)
	if err != nil {
		return err
	}
	template := cfg.Template[index]
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

	run := func(script string) error {
		if s := filter(script); len(s) > 0 {
			fmt.Println(s)
			cmds := splitCmd(s)
			cmd := exec.Command(cmds[0], cmds[1:]...)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			return cmd.Run()
		}
		return nil
	}

	if err := run(template.BeforeScript); err != nil {
		return err
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
	if err := run(template.AfterScript); err != nil {
		return err
	}

	return nil
}
