package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/docopt/docopt-go"
	"github.com/xalanq/cf-tool/client"
	"github.com/xalanq/cf-tool/config"
)

// ParsedArgs parsed arguments
type ParsedArgs struct {
	Info        client.Info
	File        string
	Specifier   []string `docopt:"<specifier>"`
	Alias       string   `docopt:"<alias>"`
	LanguageID  string `docopt:"<language-id>"`
	InputFile   string `docopt:"<input-file>"`
	Accepted    bool     `docopt:"ac"`
	All         bool     `docopt:"all"`
	Handle      string   `docopt:"<handle>"`
	Version     string   `docopt:"{version}"`
	Config      bool     `docopt:"config"`
	Submit      bool     `docopt:"submit"`
	List        bool     `docopt:"list"`
	Parse       bool     `docopt:"parse"`
	Gen         bool     `docopt:"gen"`
	Test        bool     `docopt:"test"`
	CustomTest  bool     `docopt:"customtest"`
	Watch       bool     `docopt:"watch"`
	Open        bool     `docopt:"open"`
	Stand       bool     `docopt:"stand"`
	Sid         bool     `docopt:"sid"`
	Race        bool     `docopt:"race"`
	Pull        bool     `docopt:"pull"`
	Clone       bool     `docopt:"clone"`
	Upgrade     bool     `docopt:"upgrade"`
}

// Args global variable
var Args *ParsedArgs

func parseArgs(opts docopt.Opts) error {
	cfg := config.Instance
	cln := client.Instance
	path, err := os.Getwd()
	if err != nil {
		return err
	}
	if file, ok := opts["--file"].(string); ok {
		Args.File = file
	} else if file, ok := opts["<file>"].(string); ok {
		Args.File = file
	}
	if Args.Handle == "" {
		Args.Handle = cln.Handle
	}
	info := client.Info{}
	for _, arg := range Args.Specifier {
		parsed := parseArg(arg)
		if value, ok := parsed["problemType"]; ok {
			if info.ProblemType != "" && info.ProblemType != value {
				return fmt.Errorf("Problem Type conflicts: %v %v", info.ProblemType, value)
			}
			info.ProblemType = value
		}
		if value, ok := parsed["contestID"]; ok {
			if info.ContestID != "" && info.ContestID != value {
				return fmt.Errorf("Contest ID conflicts: %v %v", info.ContestID, value)
			}
			info.ContestID = value
		}
		if value, ok := parsed["groupID"]; ok {
			if info.GroupID != "" && info.GroupID != value {
				return fmt.Errorf("Group ID conflicts: %v %v", info.GroupID, value)
			}
			info.GroupID = value
		}
		if value, ok := parsed["problemID"]; ok {
			if info.ProblemID != "" && info.ProblemID != value {
				return fmt.Errorf("Problem ID conflicts: %v %v", info.ProblemID, value)
			}
			info.ProblemID = value
		}
		if value, ok := parsed["submissionID"]; ok {
			if info.SubmissionID != "" && info.SubmissionID != value {
				return fmt.Errorf("Submission ID conflicts: %v %v", info.SubmissionID, value)
			}
			info.SubmissionID = value
		}
	}
	if info.ProblemType == "" {
		parsed := parsePath(path)
		if value, ok := parsed["problemType"]; ok {
			info.ProblemType = value
		}
		if value, ok := parsed["contestID"]; ok && info.ContestID == "" {
			info.ContestID = value
		}
		if value, ok := parsed["groupID"]; ok && info.GroupID == "" {
			info.GroupID = value
		}
		if value, ok := parsed["problemID"]; ok && info.ProblemID == "" {
			info.ProblemID = value
		}
	}
	if info.ProblemType == "" || info.ProblemType == "contest" {
		if len(info.ContestID) < 6 {
			info.ProblemType = "contest"
		} else {
			info.ProblemType = "gym"
		}
	}
	if info.ProblemType == "acmsguru" {
		if info.ContestID != "99999" && info.ContestID != "" {
			info.ProblemID = info.ContestID
		}
		info.ContestID = "99999"
	}
	root := cfg.FolderName["root"]
	info.RootPath = filepath.Join(path, root)
	for {
		base := filepath.Base(path)
		if base == root {
			info.RootPath = path
			break
		}
		if filepath.Dir(path) == path {
			break
		}
		path = filepath.Dir(path)
	}
	info.RootPath = filepath.Join(info.RootPath, cfg.FolderName[info.ProblemType])
	Args.Info = info
	// util.DebugJSON(Args)
	return nil
}

// ProblemRegStr problem
const ProblemRegStr = `\w+`

// StrictProblemRegStr strict problem
const StrictProblemRegStr = `[a-zA-Z]+\d*`

// ContestRegStr contest
const ContestRegStr = `\d+`

// GroupRegStr group
const GroupRegStr = `\w{10}`

// SubmissionRegStr submission
const SubmissionRegStr = `\d+`

// ArgRegStr for parsing arg
var ArgRegStr = [...]string{
	`^[cC][oO][nN][tT][eE][sS][tT][sS]?$`,
	`^[gG][yY][mM][sS]?$`,
	`^[gG][rR][oO][uU][pP][sS]?$`,
	`^[aA][cC][mM][sS][gG][uU][rR][uU]$`,
	fmt.Sprintf(`/contest/(?P<contestID>%v)(/problem/(?P<problemID>%v))?`, ContestRegStr, ProblemRegStr),
	fmt.Sprintf(`/gym/(?P<contestID>%v)(/problem/(?P<problemID>%v))?`, ContestRegStr, ProblemRegStr),
	fmt.Sprintf(`/problemset/problem/(?P<contestID>%v)/(?P<problemID>%v)`, ContestRegStr, ProblemRegStr),
	fmt.Sprintf(`/group/(?P<groupID>%v)(/contest/(?P<contestID>%v)(/problem/(?P<problemID>%v))?)?`, GroupRegStr, ContestRegStr, ProblemRegStr),
	fmt.Sprintf(`/problemsets/acmsguru/problem/(?P<contestID>%v)/(?P<problemID>%v)`, ContestRegStr, ProblemRegStr),
	fmt.Sprintf(`/problemsets/acmsguru/submission/(?P<contestID>%v)/(?P<submissionID>%v)`, ContestRegStr, SubmissionRegStr),
	fmt.Sprintf(`/submission/(?P<submissionID>%v)`, SubmissionRegStr),
	fmt.Sprintf(`^(?P<contestID>%v)(?P<problemID>%v)$`, ContestRegStr, StrictProblemRegStr),
	fmt.Sprintf(`^(?P<contestID>%v)$`, ContestRegStr),
	fmt.Sprintf(`^(?P<problemID>%v)$`, StrictProblemRegStr),
	fmt.Sprintf(`^(?P<groupID>%v)$`, GroupRegStr),
}

// ArgTypePathRegStr path
var ArgTypePathRegStr = [...]string{
	fmt.Sprintf("%v/%v/((?P<contestID>%v)/((?P<problemID>%v)/)?)?", "%v", "%v", ContestRegStr, ProblemRegStr),
	fmt.Sprintf("%v/%v/((?P<contestID>%v)/((?P<problemID>%v)/)?)?", "%v", "%v", ContestRegStr, ProblemRegStr),
	fmt.Sprintf("%v/%v/((?P<groupID>%v)/((?P<contestID>%v)/((?P<problemID>%v)/)?)?)?", "%v", "%v", GroupRegStr, ContestRegStr, ProblemRegStr),
	fmt.Sprintf("%v/%v/((?P<problemID>%v)/)?", "%v", "%v", ProblemRegStr),
}

// ArgType type
var ArgType = [...]string{
	"contest",
	"gym",
	"group",
	"acmsguru",
	"contest",
	"gym",
	"contest",
	"group",
	"acmsguru",
	"acmsguru",
	"",
	"",
	"",
	"",
	"",
}

func parseArg(arg string) map[string]string {
	output := make(map[string]string)
	for k, regStr := range ArgRegStr {
		reg := regexp.MustCompile(regStr)
		names := reg.SubexpNames()
		for i, val := range reg.FindStringSubmatch(arg) {
			if names[i] != "" && val != "" {
				output[names[i]] = val
			}
			if ArgType[k] != "" {
				output["problemType"] = ArgType[k]
				if k < 4 {
					return output
				}
			}
		}
	}
	return output
}

func parsePath(path string) map[string]string {
	path = filepath.ToSlash(path) + "/"
	output := make(map[string]string)
	cfg := config.Instance
	for k, problemType := range client.ProblemTypes {
		reg := regexp.MustCompile(fmt.Sprintf(ArgTypePathRegStr[k], cfg.FolderName["root"], cfg.FolderName[problemType]))
		names := reg.SubexpNames()
		for i, val := range reg.FindStringSubmatch(path) {
			if names[i] != "" && val != "" {
				output[names[i]] = val
			}
			output["problemType"] = problemType
		}
	}
	if (output["problemType"] != "" && output["problemType"] != "group") ||
		output["groupID"] == output["contestID"] ||
		output["groupID"] == fmt.Sprintf("%v%v", output["contestID"], output["problemID"]) {
		output["groupID"] = ""
	}
	if output["groupID"] != "" && output["problemType"] == "" {
		output["problemType"] = "group"
	}
	return output
}
