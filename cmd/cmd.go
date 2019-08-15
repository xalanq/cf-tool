package cmd

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/xalanq/cf-tool/client"
	"github.com/xalanq/cf-tool/config"
	"github.com/xalanq/cf-tool/util"
)

// Eval args
func Eval(args map[string]interface{}) error {
	if args["config"].(bool) {
		return Config(args)
	} else if args["submit"].(bool) {
		return Submit(args)
	} else if args["list"].(bool) {
		return List(args)
	} else if args["parse"].(bool) {
		return Parse(args)
	} else if args["gen"].(bool) {
		return Gen(args)
	} else if args["test"].(bool) {
		return Test(args)
	} else if args["watch"].(bool) {
		return Watch(args)
	} else if args["open"].(bool) {
		return Open(args)
	} else if args["stand"].(bool) {
		return Stand(args)
	} else if args["sid"].(bool) {
		return Sid(args)
	} else if args["race"].(bool) {
		return Race(args)
	} else if args["pull"].(bool) {
		return Pull(args)
	} else if args["clone"].(bool) {
		return Clone(args)
	} else if args["upgrade"].(bool) {
		return Upgrade(args["{version}"].(string))
	}
	return nil
}

func getContestID(args map[string]interface{}) (string, error) {
	if c, ok := args["<contest-id>"].(string); ok {
		if _, err := strconv.Atoi(c); err == nil {
			return c, nil
		}
		return "", fmt.Errorf(`Contest should be a number instead of "%v"`, c)
	}
	path, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		c := filepath.Base(path)
		if _, err := strconv.Atoi(c); err == nil {
			return c, nil
		}
		if filepath.Dir(path) == path {
			break
		}
		path = filepath.Dir(path)
	}
	return "", errors.New("Cannot find any valid contest id")
}

func getProblemID(args map[string]interface{}) (string, error) {
	if p, ok := args["<problem-id>"].(string); ok {
		return strings.ToLower(p), nil
	}
	path, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return strings.ToLower(filepath.Base(path)), nil
}

func getSampleID() (samples []string) {
	path, err := os.Getwd()
	if err != nil {
		return
	}
	paths, err := ioutil.ReadDir(path)
	if err != nil {
		return
	}
	reg := regexp.MustCompile(`in(\d+).txt`)
	for _, path := range paths {
		name := path.Name()
		tmp := reg.FindSubmatch([]byte(name))
		if tmp != nil {
			idx := string(tmp[1])
			ans := fmt.Sprintf("ans%v.txt", idx)
			if _, err := os.Stat(ans); err == nil {
				samples = append(samples, idx)
			}
		}
	}
	return
}

// CodeList Name matches some template suffix, index are template array indexes
type CodeList struct {
	Name  string
	Index []int
}

func getCode(args map[string]interface{}, templates []config.CodeTemplate) (codes []CodeList) {
	mp := make(map[string][]int)
	for i, temp := range templates {
		suffixMap := map[string]bool{}
		for _, suffix := range temp.Suffix {
			if _, ok := suffixMap[suffix]; !ok {
				suffixMap[suffix] = true
				sf := "." + suffix
				mp[sf] = append(mp[sf], i)
			}
		}
	}

	if filename, ok := args["<filename>"].(string); ok {
		ext := filepath.Ext(filename)
		if idx, ok := mp[ext]; ok {
			return []CodeList{CodeList{filename, idx}}
		}
		return
	}

	path, err := os.Getwd()
	if err != nil {
		return
	}
	paths, err := ioutil.ReadDir(path)
	if err != nil {
		return
	}

	for _, path := range paths {
		name := path.Name()
		ext := filepath.Ext(name)
		if idx, ok := mp[ext]; ok {
			codes = append(codes, CodeList{name, idx})
		}
	}

	return codes
}

func getOneCode(args map[string]interface{}, templates []config.CodeTemplate) (name string, index int, err error) {
	codes := getCode(args, templates)
	if len(codes) < 1 {
		return "", 0, errors.New("Cannot find any code.\nMaybe you should add a new template by `cf config`")
	}
	if len(codes) > 1 {
		color.Cyan("There are multiple files can be selected.")
		for i, code := range codes {
			fmt.Printf("%3v: %v\n", i, code.Name)
		}
		i := util.ChooseIndex(len(codes))
		codes[0] = codes[i]
	}
	if len(codes[0].Index) > 1 {
		color.Cyan("There are multiple languages match the file.")
		for i, idx := range codes[0].Index {
			fmt.Printf("%3v: %v\n", i, client.Langs[templates[idx].Lang])
		}
		i := util.ChooseIndex(len(codes[0].Index))
		codes[0].Index[0] = codes[0].Index[i]
	}
	return codes[0].Name, codes[0].Index[0], nil
}

func loginAgain(cfg *config.Config, cln *client.Client, err error) error {
	if err != nil && err.Error() == client.ErrorNotLogged {
		color.Red("Not logged. Try to login\n")
		password, e := cfg.DecryptPassword()
		if e != nil {
			return e
		}
		err = cln.Login(cfg.Username, password)
	}
	return err
}
