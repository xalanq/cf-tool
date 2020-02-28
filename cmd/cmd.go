package cmd

import (
	"errors"
	"fmt"
	"github.com/docopt/docopt-go"
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
func Eval(args docopt.Opts) error {
	parsed := ParsedArgs{}
	args.Bind(&parsed)
	if parsed.Config {
		return Config()
	} else if parsed.Submit {
		return Submit(args)
	} else if parsed.List {
		return List(args)
	} else if parsed.Parse {
		return Parse(args)
	} else if parsed.Generate {
		return Gen(args)
	} else if parsed.Test {
		return Test(args)
	} else if parsed.Watch {
		return Watch(args)
	} else if parsed.Open {
		return Open(args)
	} else if parsed.Standings {
		return Stand(args)
	} else if parsed.Sid {
		return Sid(args)
	} else if parsed.Race {
		return Race(args)
	} else if parsed.Pull {
		return Pull(args)
	} else if parsed.Clone {
		return Clone(args)
	} else if parsed.Upgrade {
		return Upgrade(parsed.Version)
	}
	return nil
}

type ParseRequirement struct {
	ContestID, ProblemID, SubmissionID, Filename, Alias, Username bool
}

type ParsedArgs struct {
	ContestID       string `docopt:"<url | contest-id>"`
	ProblemID       string `docopt:"<problem-id>"`
	SubmissionID    string `docopt:"<submission-id>"`
	Filename        string `docopt:"<filename>"`
	Alias           string `docopt:"<alias>"`
	Accepted        bool   `docopt:"ac"`
	All             bool   `docopt:"all"`
	Handle          string `docopt:"<handle>"`
	Version         string `docopt:"{version}"`
	Config          bool   `docopt:"config"`
	Submit          bool   `docopt:"submit"`
	List            bool   `docopt:"list"`
	Parse           bool   `docopt:"parse"`
	Generate        bool   `docopt:"gen"`
	Test            bool   `docopt:"test"`
	Watch           bool   `docopt:"watch"`
	Open            bool   `docopt:"open"`
	Standings       bool   `docopt:"stand"`
	Sid             bool   `docopt:"sid"`
	Race            bool   `docopt:"race"`
	Pull            bool   `docopt:"pull"`
	Clone           bool   `docopt:"clone"`
	Upgrade         bool   `docopt:"upgrade"`
	ContestRootPath string
}

func parseArgs(args interface{}, required ParseRequirement) (ParsedArgs, error) {
	opts, ok := args.(docopt.Opts)
	result := ParsedArgs{}
	if !ok {
		return result, errors.New("args must be docopt.Opts type")
	}
	opts.Bind(&result)
	contestID, problemID, lastDir := "", "", ""
	path, err := os.Getwd()
	if err != nil {
		return result, err
	}
	result.ContestRootPath = path
	for {
		c := filepath.Base(path)
		if _, err := strconv.Atoi(c); err == nil {
			contestID, problemID = c, strings.ToLower(lastDir)
			if result.ContestID == "" {
				result.ContestRootPath = filepath.Dir(path)
			}
			break
		}
		if filepath.Dir(path) == path {
			break
		}
		path, lastDir = filepath.Dir(path), c
	}
	if result.ProblemID != "" {
		problemID = strings.ToLower(result.ProblemID)
	}
	if util.IsUrl(result.ContestID) {
		parsed, err := parseUrl(result.ContestID)
		if err != nil {
			return result, err
		}
		if value, ok := parsed["contestID"]; ok {
			contestID = value
		}
		if value, ok := parsed["problemID"]; ok {
			problemID = strings.ToLower(value)
		}
	} else if _, err := strconv.Atoi(result.ContestID); err == nil {
		contestID = result.ContestID
	}
	result.ContestID = contestID
	result.ProblemID = problemID
	if required.ContestID && contestID == "" {
		return result, errors.New("Unable to find <contest-id>")
	}
	if required.ProblemID && problemID == "" {
		return result, errors.New("Unable to find <problem-id>")
	}
	if required.SubmissionID && result.SubmissionID == "" {
		return result, errors.New("Unable to find <submission-id>")
	}
	if required.Alias && result.Alias == "" {
		return result, errors.New("Unable to find <alias>")
	}
	if required.Filename && result.Filename == "" {
		return result, errors.New("Unable to find <filename>")
	}

	return result, nil
}

func parseUrl(url string) (map[string]string, error) {
	reg := regexp.MustCompile(`/(?P<type>problemset|gym|contest|group)`)
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
		reg_str = `/contest/(?P<contestID>\d+)(/problem/(?P<problemID>[\w\d]+))?`
	case "gym":
		reg_str = `/gym/(?P<contestID>\d+)(/problem/(?P<problemID>[\w\d]+))?`
	case "problemset":
		reg_str = `/problemset/problem/(?P<contestID>\d+)/(?P<problemID>[\w\d]+)?`
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
