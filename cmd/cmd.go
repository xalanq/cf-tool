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

func parseArgs(args map[string]interface{}, required map[string]bool) (map[string]string, error) {
	result := make(map[string]string)
	contestID, problemID, lastDir := "", "", ""
	path, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	result["contestRootPath"] = path
	for {
		c := filepath.Base(path)
		if _, err := strconv.Atoi(c); err == nil {
			contestID, problemID = c, strings.ToLower(lastDir)
			if _, ok := args["<url | contest-id>"].(string); !ok {
				result["contestRootPath"] = filepath.Dir(path)
			}
			break
		}
		if filepath.Dir(path) == path {
			break
		}
		path, lastDir = filepath.Dir(path), c
	}
	if p, ok := args["<problem-id>"].(string); ok {
		problemID = strings.ToLower(p)
	}
	if c, ok := args["<url | contest-id>"].(string); ok {
		if util.IsUrl(c) {
			parsed, err := parseUrl(c)
			if err != nil {
				return nil, err
			}
			if value, ok := parsed["contestID"]; ok {
				contestID = value
			}
			if value, ok := parsed["problemID"]; ok {
				problemID = strings.ToLower(value)
			}
		} else if _, err := strconv.Atoi(c); err == nil {
			contestID = c
		}
	}
	if req, ok := required["<contest-id>"]; ok {
		result["<contest-id>"] = contestID
		if contestID == "" && req {
			return nil, errors.New("Unable to find <contest-id>")
		}
	}
	if req, ok := required["<problem-id>"]; ok {
		result["<problem-id>"] = problemID
		if problemID == "" && req {
			return nil, errors.New("Unable to find <problem-id>")
		}
	}
	for key, req := range required {
		if _, ok := result[key]; ok {
			continue
		}
		value, ok := args[key].(string)
		if req && !ok {
			return nil, errors.New("Unable to find " + key)
		}
		result[key] = value
	}
	return result, nil
}

func parseUrl(url string) (map[string]string, error) {
	reg := regexp.MustCompile(`(https?:\/\/)?(www\.)?([a-zA-Z\d\-\.]+)\/(?P<type>problemset|gym|contest|group)`)
	url_type := ""
	for i, val := range reg.FindStringSubmatch(url) {
		if reg.SubexpNames()[i] == "type" {
			url_type = val
			break
		}
	}

	reg_str := ""
	switch url_type {
	case "contest":
		reg_str = `(https?:\/\/)?(www\.)?([a-zA-Z\d\-\.]+)\/contest\/(?P<contestID>\d+)(\/problem\/(?P<problemID>[\w\d]+))?`
	case "gym":
		reg_str = `(https?:\/\/)?(www\.)?([a-zA-Z\d\-\.]+)\/gym\/(?P<contestID>\d+)(\/problem\/(?P<problemID>[\w\d]+))?`
	case "problemset":
		reg_str = `(https?:\/\/)?(www\.)?([a-zA-Z\d\-\.]+)\/problemset\/problem\/(?P<contestID>\d+)\/(?P<problemID>[\w\d]+)?`
	case "group":
		return nil, errors.New("Groups are not supported")
	default:
		return nil, errors.New("Invalid url")
	}

	output := make(map[string]string)
	reg = regexp.MustCompile(reg_str)
	names := reg.SubexpNames()
	for i, val := range reg.FindStringSubmatch(url) {
		if names[i] != "" && val != "" {
			output[names[i]] = val
		}
	}
	output["type"] = url_type
	return output, nil
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

func getCode(filename string, templates []config.CodeTemplate) (codes []CodeList) {
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

	if filename != "" {
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

func getOneCode(filename string, templates []config.CodeTemplate) (name string, index int, err error) {
	codes := getCode(filename, templates)
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
		err = cln.Login()
	}
	return err
}
