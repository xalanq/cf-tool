package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/gorilla/websocket"
	ansi "github.com/k0kubun/go-ansi"
)

func (s *SubmitState) update(text interface{}) bool {
	d := text.([]interface{})
	if len(d) < 17 {
		return false
	}
	id := uint64(d[1].(float64))
	if id != s.id {
		return false
	}
	if waits := d[6].(string); !(waits == "null" || waits == "TESTING" || waits == "SUBMITTED") {
		s.end = true
		timeConsumed := uint64(d[9].(float64))
		memoryConsumed := uint64(d[10].(float64))
		s.time = timeConsumed
		s.memory = memoryConsumed
	}
	s.state = d[12].(string)
	if tmp, ok := d[8].(float64); ok {
		if judgedTestCount := uint64(tmp); judgedTestCount >= s.judged {
			s.judged = judgedTestCount
		}
	}
	if tmp, ok := d[7].(float64); ok {
		if passedTestCount := uint64(tmp); passedTestCount >= s.passed {
			s.passed = passedTestCount
		}
	}
	if tmp, ok := d[5].(float64); ok {
		if points := uint64(tmp); points >= s.points {
			s.points = points
		}
	}
	return true
}

func (s *SubmitState) display() {
	state := stateToText[s.state]
	state = strings.ReplaceAll(state, "${f-points}", fmt.Sprintf("%v", s.points))
	state = strings.ReplaceAll(state, "${f-passed}", fmt.Sprintf("%v", s.passed))
	state = strings.ReplaceAll(state, "${f-judged}", fmt.Sprintf("%v", s.judged))
	for k, v := range colorMap {
		tmp := strings.ReplaceAll(state, k, "")
		if tmp != state {
			state = color.New(v).Sprint(tmp)
		}
	}
	memory := fmt.Sprintf("%v B", s.memory)
	if s.memory > 1024 {
		memory = fmt.Sprintf("%.2f KB", float64(s.memory)/1024.0)
		if s.memory > 1024*1024 {
			memory = fmt.Sprintf("%.2f MB", float64(s.memory)/1024.0/1024.0)
		}
	}
	if s.state != "---" {
		ansi.CursorUp(6)
	}
	ansi.Printf("      #: %v\n", s.id)
	ansi.Printf("   prob: %v\n", s.name)
	ansi.Printf("                                               \n")
	ansi.CursorUp(1)
	ansi.Printf("  state: %v\n", state)
	ansi.Printf("   lang: %v\n", s.lang)
	ansi.Printf("   time: %v ms\n", s.time)
	ansi.Printf(" memory: %v\n", memory)
}

// findSubmission just find
func findSubmission(body []byte) ([]byte, error) {
	reg, _ := regexp.Compile(`<tr data-submission[\s\S]+?</tr>`)
	tmp := reg.Find(body)
	if tmp == nil {
		return nil, errors.New("Cannot find submission")
	}
	return tmp, nil
}

// findSubmitID just find
func findSubmitID(body []byte) (string, error) {
	reg, _ := regexp.Compile(`submission/(\d+?)"`)
	tmp := reg.FindSubmatch(body)
	if tmp == nil {
		return "", errors.New("Cannot find submitID")
	}
	return string(tmp[1]), nil
}

// findSubmitName just find
func findSubmitName(body []byte) (string, error) {
	reg, _ := regexp.Compile(`<a href="/contest[\s\S]*?">([\s\S]*?)</a>`)
	tmp := reg.FindSubmatch(body)
	if tmp == nil {
		return "", errors.New("Cannot find submit name")
	}
	return strings.TrimSpace(string(tmp[1])), nil
}

// findSubmitName just find
func findSubmitLang(body []byte) (string, error) {
	reg, _ := regexp.Compile(`<td>([\s\S]*?)</td>`)
	tmp := reg.FindSubmatch(body)
	if tmp == nil {
		return "", errors.New("Cannot find submit lang")
	}
	return strings.TrimSpace(string(tmp[1])), nil
}

// findChannel websocket channel
func findChannel(body []byte) []string {
	reg, _ := regexp.Compile(`name="cc" content="(.+?)"[\s\S]*name="pc" content="(.+?)"`)
	// reg, _ := regexp.Compile(`name="cc" content="(.+?)"`)
	tmp := reg.FindSubmatch(body)
	var ret []string
	for i := 1; i < len(tmp); i++ {
		ret = append(ret, "s_"+string(tmp[i]))
	}
	return ret
}

// SubmitContest submit problem in contest (and block util pending)
func (c *Client) SubmitContest(contestID, probID, langID, source string) (err error) {
	fmt.Printf("Try to submit %v %v %v\n", contestID, probID, Langs[langID])
	color.Output = ansi.NewAnsiStdout()
	submitURL := fmt.Sprintf("https://codeforces.com/contest/%v/submit", contestID)

	client := &http.Client{Jar: c.Jar}
	resp, err := client.Get(submitURL)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	err = checkLogin(c.Username, body)
	if err != nil {
		return
	}

	fmt.Printf("Current user: %v\n", c.Username)

	csrf, err := findCsrf(body)
	if err != nil {
		return
	}

	resp, err = client.PostForm(fmt.Sprintf("%v?csrf=%v", submitURL, csrf), url.Values{
		"csrf_token":            {csrf},
		"ftaa":                  {c.Ftaa},
		"bfaa":                  {c.Bfaa},
		"action":                {"submitSolutionFormSubmitted"},
		"submittedProblemIndex": {probID},
		"programTypeId":         {langID},
		"source":                {source},
		"tabSize":               {"4"},
		"_tta":                  {"594"},
	})
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	submission, err := findSubmission(body)
	if err != nil {
		return
	}
	submitID, err := findSubmitID(submission)
	if err != nil {
		return
	}
	submitName, err := findSubmitName(submission)
	if err != nil {
		return
	}
	submitLang, err := findSubmitLang(submission)
	if err != nil {
		return
	}
	fmt.Println("Submitted")
	channels := findChannel(body)
	tm := time.Now().UTC().Format("20060102150405")
	url := fmt.Sprintf(`wss://pubsub.codeforces.com/ws/%v?_=%v&tag=&time=&eventid=`, strings.Join(channels[:], "/"), tm)
	ws, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return
	}

	var state SubmitState
	state.id, err = strconv.ParseUint(submitID, 10, 64)
	state.lang = submitLang
	state.name = submitName
	state.state = "---"
	if err != nil {
		return
	}
	state.display()

	done := make(chan struct{})
	go func() {
		defer close(done)
		for !state.end {
			func() {
				defer func() {
					recover()
				}()
				_, bytes, err := ws.ReadMessage()
				if err != nil {
					return
				}
				var recv map[string]interface{}
				err = json.Unmarshal(bytes, &recv)
				if err != nil {
					return
				}
				c := recv["channel"].(string)
				if c != channels[0] {
					return
				}
				var sub map[string]interface{}
				text := recv["text"].(string)
				err = json.Unmarshal([]byte(text), &sub)
				if err != nil {
					return
				}
				t := sub["t"].(string)
				if t != "s" {
					return
				}
				if state.update(sub["d"]) {
					state.display()
				}
			}()
		}
	}()

	<-done

	return
}

func findLangBlock(body []byte) ([]byte, error) {
	reg, _ := regexp.Compile(`name="programTypeId".+?</select>`)
	tmp := reg.Find(body)
	if tmp == nil {
		return nil, errors.New("Cannot find language selection")
	}
	return tmp, nil
}

func findLang(body []byte) (map[string]string, error) {
	reg, _ := regexp.Compile(`value="(.+?)"[\s\S]*?>([\s\S]+?)<`)
	tmp := reg.FindAllSubmatch(body, -1)
	if tmp == nil {
		return nil, errors.New("Cannot find any language")
	}
	ret := make(map[string]string)
	for i := 0; i < len(tmp); i++ {
		ret[string(tmp[i][1])] = string(tmp[i][2])
	}
	return ret, nil
}

// GetLangList get language list from url (require login)
func (c *Client) GetLangList(url string) (langs map[string]string, err error) {
	resp, err := http.Get(url)
	if err != nil {
		return
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	fmt.Println(string(body))
	block, err := findLangBlock(body)
	if err != nil {
		return
	}

	return findLang(block)
}

// Langs generated by
// ^[\s\S]*?value="(.+?)"[\s\S]*?>([\s\S]+?)<[\s\S]*?$
//     "\1": "\2",
var Langs = map[string]string{
	"43": "GNU GCC C11 5.1.0",
	"52": "Clang++17 Diagnostics",
	"42": "GNU G++11 5.1.0",
	"50": "GNU G++14 6.4.0",
	"54": "GNU G++17 7.3.0",
	"2":  "Microsoft Visual C++ 2010",
	"59": "Microsoft Visual C++ 2017",
	"9":  "C# Mono 5.18",
	"28": "D DMD32 v2.083.1",
	"32": "Go 1.11.4",
	"12": "Haskell GHC 8.6.3",
	"36": "Java 1.8.0_162",
	"48": "Kotlin 1.3.10",
	"19": "OCaml 4.02.1",
	"3":  "Delphi 7",
	"4":  "Free Pascal 3.0.2",
	"51": "PascalABC.NET 3.4.2",
	"13": "Perl 5.20.1",
	"6":  "PHP 7.2.13",
	"7":  "Python 2.7.15",
	"31": "Python 3.7.2",
	"40": "PyPy 2.7 (6.0.0)",
	"41": "PyPy 3.5 (6.0.0)",
	"8":  "Ruby 2.0.0p645",
	"49": "Rust 1.31.1",
	"20": "Scala 2.12.8",
	"34": "JavaScript V8 4.8.0",
	"55": "Node.js 9.4.0",
}

var colorMap = map[string]color.Attribute{
	"${c-waiting}":  color.FgWhite,
	"${c-failed}":   color.FgRed,
	"${c-accepted}": color.FgGreen,
	"${c-rejected}": color.FgBlue,
}

// <span\sclass=(\\["'])?([\S^>]+?)\1?>[\d]+?<\\/span>
// ${\2}
// <span\sclass=\\["']([\S^>]+?)\\["']>(.*?)<\\/span>
// ${\1}\2
// preparedVerdictFormats\[(\S+?)\]\s=\s(.+?);
//     \1: \2,
// verdict-format-
// f-
// verdict-
// c-
var stateToText = map[string]string{
	"---":    "${c-waiting}In queue",
	"30000":  "${c-failed}Denial of judgement",
	"30001":  "${c-failed}Denial of judgement",
	"30010":  "${c-failed}Denial of judgement",
	"30011":  "${c-failed}Denial of judgement",
	"30020":  "${c-failed}Denial of judgement",
	"30021":  "${c-failed}Denial of judgement",
	"30110":  "${c-failed}Denial of judgement",
	"30111":  "${c-failed}Denial of judgement",
	"30120":  "${c-failed}Denial of judgement",
	"30121":  "${c-failed}Denial of judgement",
	"30220":  "${c-failed}Denial of judgement",
	"30221":  "${c-failed}Denial of judgement",
	"31000":  "${c-accepted}Pretests and hacks passed",
	"31001":  "${c-accepted}Perfect result: ${f-points} points",
	"31010":  "${c-accepted}Pretests and hacks passed",
	"31011":  "${c-accepted}Perfect result: ${f-points} points",
	"31020":  "${c-accepted}Pretests and hacks passed",
	"31021":  "${c-accepted}Perfect result: ${f-points} points",
	"31110":  "${c-accepted}Pretests and hacks passed",
	"31111":  "${c-accepted}Perfect result: ${f-points} points",
	"31120":  "${c-accepted}Pretests and hacks passed",
	"31121":  "${c-accepted}Perfect result: ${f-points} points",
	"31220":  "${c-accepted}Pretests and hacks passed",
	"31221":  "${c-accepted}Perfect result: ${f-points} points",
	"32000":  "Partial (hacks)",
	"32001":  "Partial result: ${f-points} points",
	"32010":  "Partial: ${f-passed} hacks ouf of ${f-judged}",
	"32011":  "Partial result: ${f-points} points",
	"32020":  "Partial: ${f-passed} hacks ouf of ${f-judged}",
	"32021":  "Partial result: ${f-points} points",
	"32110":  "Partial: ${f-passed} hacks ouf of ${f-judged}",
	"32111":  "Partial result: ${f-points} points",
	"32120":  "Partial: ${f-passed} hacks ouf of ${f-judged}",
	"32121":  "Partial result: ${f-points} points",
	"32220":  "Partial: ${f-passed} hacks ouf of ${f-judged}",
	"32221":  "Partial result: ${f-points} points",
	"33000":  "Compilation error",
	"33001":  "Compilation error",
	"33010":  "Compilation error",
	"33011":  "Compilation error",
	"33020":  "Compilation error",
	"33021":  "Compilation error",
	"33110":  "Compilation error",
	"33111":  "Compilation error",
	"33120":  "Compilation error",
	"33121":  "Compilation error",
	"33220":  "Compilation error",
	"33221":  "Compilation error",
	"34000":  "${c-rejected}Runtime error on hack",
	"34001":  "${c-rejected}Runtime error on hack",
	"34010":  "${c-rejected}Runtime error on hack ${f-judged}",
	"34011":  "${c-rejected}Runtime error on hack ${f-judged}",
	"34020":  "${c-rejected}Runtime error on hack ${f-judged}",
	"34021":  "${c-rejected}Runtime error on hack ${f-judged}",
	"34110":  "${c-rejected}Runtime error on hack ${f-judged}",
	"34111":  "${c-rejected}Runtime error on hack ${f-judged}",
	"34120":  "${c-rejected}Runtime error on hack ${f-judged}",
	"34121":  "${c-rejected}Runtime error on hack ${f-judged}",
	"34220":  "${c-rejected}Runtime error on hack ${f-judged}",
	"34221":  "${c-rejected}Runtime error on hack ${f-judged}",
	"35000":  "${c-rejected}Wrong answer on hack",
	"35001":  "${c-rejected}Wrong answer on hack",
	"35010":  "${c-rejected}Wrong answer on hack ${f-judged}",
	"35011":  "${c-rejected}Wrong answer on hack ${f-judged}",
	"35020":  "${c-rejected}Wrong answer on hack ${f-judged}",
	"35021":  "${c-rejected}Wrong answer on hack ${f-judged}",
	"35110":  "${c-rejected}Wrong answer on hack ${f-judged}",
	"35111":  "${c-rejected}Wrong answer on hack ${f-judged}",
	"35120":  "${c-rejected}Wrong answer on hack ${f-judged}",
	"35121":  "${c-rejected}Wrong answer on hack ${f-judged}",
	"35220":  "${c-rejected}Wrong answer on hack ${f-judged}",
	"35221":  "${c-rejected}Wrong answer on hack ${f-judged}",
	"36000":  "${c-rejected}Presentation error on hack",
	"36001":  "${c-rejected}Presentation error on hack",
	"36010":  "${c-rejected}Presentation error on hack ${f-judged}",
	"36011":  "${c-rejected}Presentation error on hack ${f-judged}",
	"36020":  "${c-rejected}Presentation error on hack ${f-judged}",
	"36021":  "${c-rejected}Presentation error on hack ${f-judged}",
	"36110":  "${c-rejected}Presentation error on hack ${f-judged}",
	"36111":  "${c-rejected}Presentation error on hack ${f-judged}",
	"36120":  "${c-rejected}Presentation error on hack ${f-judged}",
	"36121":  "${c-rejected}Presentation error on hack ${f-judged}",
	"36220":  "${c-rejected}Presentation error on hack ${f-judged}",
	"36221":  "${c-rejected}Presentation error on hack ${f-judged}",
	"37000":  "${c-rejected}Time limit exceeded on hack",
	"37001":  "${c-rejected}Time limit exceeded on hack",
	"37010":  "${c-rejected}Time limit exceeded on hack ${f-judged}",
	"37011":  "${c-rejected}Time limit exceeded on hack ${f-judged}",
	"37020":  "${c-rejected}Time limit exceeded on hack ${f-judged}",
	"37021":  "${c-rejected}Time limit exceeded on hack ${f-judged}",
	"37110":  "${c-rejected}Time limit exceeded on hack ${f-judged}",
	"37111":  "${c-rejected}Time limit exceeded on hack ${f-judged}",
	"37120":  "${c-rejected}Time limit exceeded on hack ${f-judged}",
	"37121":  "${c-rejected}Time limit exceeded on hack ${f-judged}",
	"37220":  "${c-rejected}Time limit exceeded on hack ${f-judged}",
	"37221":  "${c-rejected}Time limit exceeded on hack ${f-judged}",
	"38000":  "${c-rejected}Memory limit exceeded on hack",
	"38001":  "${c-rejected}Memory limit exceeded on hack",
	"38010":  "${c-rejected}Memory limit exceeded on hack ${f-judged}",
	"38011":  "${c-rejected}Memory limit exceeded on hack ${f-judged}",
	"38020":  "${c-rejected}Memory limit exceeded on hack ${f-judged}",
	"38021":  "${c-rejected}Memory limit exceeded on hack ${f-judged}",
	"38110":  "${c-rejected}Memory limit exceeded on hack ${f-judged}",
	"38111":  "${c-rejected}Memory limit exceeded on hack ${f-judged}",
	"38120":  "${c-rejected}Memory limit exceeded on hack ${f-judged}",
	"38121":  "${c-rejected}Memory limit exceeded on hack ${f-judged}",
	"38220":  "${c-rejected}Memory limit exceeded on hack ${f-judged}",
	"38221":  "${c-rejected}Memory limit exceeded on hack ${f-judged}",
	"39000":  "${c-rejected}Idleness limit exceeded on hack",
	"39001":  "${c-rejected}Idleness limit exceeded on hack",
	"39010":  "${c-rejected}Idleness limit exceeded on hack ${f-judged}",
	"39011":  "${c-rejected}Idleness limit exceeded on hack ${f-judged}",
	"39020":  "${c-rejected}Idleness limit exceeded on hack ${f-judged}",
	"39021":  "${c-rejected}Idleness limit exceeded on hack ${f-judged}",
	"39110":  "${c-rejected}Idleness limit exceeded on hack ${f-judged}",
	"39111":  "${c-rejected}Idleness limit exceeded on hack ${f-judged}",
	"39120":  "${c-rejected}Idleness limit exceeded on hack ${f-judged}",
	"39121":  "${c-rejected}Idleness limit exceeded on hack ${f-judged}",
	"39220":  "${c-rejected}Idleness limit exceeded on hack ${f-judged}",
	"39221":  "${c-rejected}Idleness limit exceeded on hack ${f-judged}",
	"310000": "${c-rejected}Security violated on hack",
	"310001": "${c-rejected}Security violated on hack",
	"310010": "${c-rejected}Security violated on hack ${f-judged}",
	"310011": "${c-rejected}Security violated on hack ${f-judged}",
	"310020": "${c-rejected}Security violated on hack ${f-judged}",
	"310021": "${c-rejected}Security violated on hack ${f-judged}",
	"310110": "${c-rejected}Security violated on hack ${f-judged}",
	"310111": "${c-rejected}Security violated on hack ${f-judged}",
	"310120": "${c-rejected}Security violated on hack ${f-judged}",
	"310121": "${c-rejected}Security violated on hack ${f-judged}",
	"310220": "${c-rejected}Security violated on hack ${f-judged}",
	"310221": "${c-rejected}Security violated on hack ${f-judged}",
	"311000": "${c-failed}Judgement crashed on hack",
	"311001": "${c-failed}Judgement crashed on hack",
	"311010": "${c-failed}Judgement crashed on hack",
	"311011": "${c-failed}Judgement crashed on hack",
	"311020": "${c-failed}Judgement crashed on hack",
	"311021": "${c-failed}Judgement crashed on hack",
	"311110": "${c-failed}Judgement crashed on hack",
	"311111": "${c-failed}Judgement crashed on hack",
	"311120": "${c-failed}Judgement crashed on hack",
	"311121": "${c-failed}Judgement crashed on hack",
	"311220": "${c-failed}Judgement crashed on hack",
	"311221": "${c-failed}Judgement crashed on hack",
	"312000": "${c-failed}Input preparation failed on hack",
	"312001": "${c-failed}Input preparation failed on hack",
	"312010": "${c-failed}Input preparation failed on hack ${f-judged}",
	"312011": "${c-failed}Input preparation failed on hack ${f-judged}",
	"312020": "${c-failed}Input preparation failed on hack ${f-judged}",
	"312021": "${c-failed}Input preparation failed on hack ${f-judged}",
	"312110": "${c-failed}Input preparation failed on hack ${f-judged}",
	"312111": "${c-failed}Input preparation failed on hack ${f-judged}",
	"312120": "${c-failed}Input preparation failed on hack ${f-judged}",
	"312121": "${c-failed}Input preparation failed on hack ${f-judged}",
	"312220": "${c-failed}Input preparation failed on hack ${f-judged}",
	"312221": "${c-failed}Input preparation failed on hack ${f-judged}",
	"313000": "${c-failed}Hacked",
	"313001": "${c-failed}Hacked",
	"313010": "${c-failed}Hacked",
	"313011": "${c-failed}Hacked",
	"313020": "${c-failed}Hacked",
	"313021": "${c-failed}Hacked",
	"313110": "${c-failed}Hacked",
	"313111": "${c-failed}Hacked",
	"313120": "${c-failed}Hacked",
	"313121": "${c-failed}Hacked",
	"313220": "${c-failed}Hacked",
	"313221": "${c-failed}Hacked",
	"314000": "Skipped",
	"314001": "Skipped",
	"314010": "Skipped",
	"314011": "Skipped",
	"314020": "Skipped",
	"314021": "Skipped",
	"314110": "Skipped",
	"314111": "Skipped",
	"314120": "Skipped",
	"314121": "Skipped",
	"314220": "Skipped",
	"314221": "Skipped",
	"315000": "${c-waiting}Running on hack",
	"315001": "${c-waiting}Running on hack",
	"315010": "${c-waiting}Running on hack ${f-judged}",
	"315011": "${c-waiting}Running on hack ${f-judged}",
	"315020": "${c-waiting}Running on hack ${f-judged}",
	"315021": "${c-waiting}Running on hack ${f-judged}",
	"315110": "${c-waiting}Running on hack ${f-judged}",
	"315111": "${c-waiting}Running on hack ${f-judged}",
	"315120": "${c-waiting}Running on hack ${f-judged}",
	"315121": "${c-waiting}Running on hack ${f-judged}",
	"315220": "${c-waiting}Running on hack ${f-judged}",
	"315221": "${c-waiting}Running on hack ${f-judged}",
	"316000": "${c-rejected}Rejected on hack",
	"316001": "${c-rejected}Rejected on hack",
	"316010": "${c-rejected}Rejected on hack ${f-judged}",
	"316011": "${c-rejected}Rejected on hack ${f-judged}",
	"316020": "${c-rejected}Rejected on hack ${f-judged}",
	"316021": "${c-rejected}Rejected on hack ${f-judged}",
	"316110": "${c-rejected}Rejected on hack ${f-judged}",
	"316111": "${c-rejected}Rejected on hack ${f-judged}",
	"316120": "${c-rejected}Rejected on hack ${f-judged}",
	"316121": "${c-rejected}Rejected on hack ${f-judged}",
	"316220": "${c-rejected}Rejected on hack ${f-judged}",
	"316221": "${c-rejected}Rejected on hack ${f-judged}",
	"317000": "StatusForChallenge::submitted",
	"317001": "StatusForChallenge::submitted",
	"317010": "StatusForChallenge::submitted 1 tests / =1",
	"317011": "StatusForChallenge::submitted 1 tests / =1",
	"317020": "StatusForChallenge::submitted 2 tests / 2-4",
	"317021": "StatusForChallenge::submitted 2 tests / 2-4",
	"317110": "StatusForChallenge::submitted 1 tests / =1",
	"317111": "StatusForChallenge::submitted 1 tests / =1",
	"317120": "StatusForChallenge::submitted 2 tests / 2-4",
	"317121": "StatusForChallenge::submitted 2 tests / 2-4",
	"317220": "StatusForChallenge::submitted 2 tests / 2-4",
	"317221": "StatusForChallenge::submitted 2 tests / 2-4",
	"3-000":  "${c-waiting}In queue",
	"3-001":  "${c-waiting}In queue",
	"3-010":  "${c-waiting}In queue",
	"3-011":  "${c-waiting}In queue",
	"3-020":  "${c-waiting}In queue",
	"3-021":  "${c-waiting}In queue",
	"3-110":  "${c-waiting}In queue",
	"3-111":  "${c-waiting}In queue",
	"3-120":  "${c-waiting}In queue",
	"3-121":  "${c-waiting}In queue",
	"3-220":  "${c-waiting}In queue",
	"3-221":  "${c-waiting}In queue",
	"10000":  "${c-failed}Judgement Failed",
	"10001":  "${c-failed}Judgement Failed",
	"10010":  "${c-failed}Judgement Failed",
	"10011":  "${c-failed}Judgement Failed",
	"10020":  "${c-failed}Judgement Failed",
	"10021":  "${c-failed}Judgement Failed",
	"10110":  "${c-failed}Judgement Failed",
	"10111":  "${c-failed}Judgement Failed",
	"10120":  "${c-failed}Judgement Failed",
	"10121":  "${c-failed}Judgement Failed",
	"10220":  "${c-failed}Judgement Failed",
	"10221":  "${c-failed}Judgement Failed",
	"11000":  "${c-accepted}Pretests passed",
	"11001":  "${c-accepted}Perfect result: ${f-points} points",
	"11010":  "${c-accepted}Pretests passed",
	"11011":  "${c-accepted}Perfect result: ${f-points} points",
	"11020":  "${c-accepted}Pretests passed",
	"11021":  "${c-accepted}Perfect result: ${f-points} points",
	"11110":  "${c-accepted}Pretests passed",
	"11111":  "${c-accepted}Perfect result: ${f-points} points",
	"11120":  "${c-accepted}Pretests passed",
	"11121":  "${c-accepted}Perfect result: ${f-points} points",
	"11220":  "${c-accepted}Pretests passed",
	"11221":  "${c-accepted}Perfect result: ${f-points} points",
	"12000":  "Partial (pretests)",
	"12001":  "Partial result: ${f-points} points",
	"12010":  "Partial: ${f-passed} pretests out of ${f-judged}",
	"12011":  "Partial result: ${f-points} points",
	"12020":  "Partial: ${f-passed} pretests out of ${f-judged}",
	"12021":  "Partial result: ${f-points} points",
	"12110":  "Partial: ${f-passed} pretests out of ${f-judged}",
	"12111":  "Partial result: ${f-points} points",
	"12120":  "Partial: ${f-passed} pretests out of ${f-judged}",
	"12121":  "Partial result: ${f-points} points",
	"12220":  "Partial: ${f-passed} pretests out of ${f-judged}",
	"12221":  "Partial result: ${f-points} points",
	"13000":  "Compilation error",
	"13001":  "Compilation error",
	"13010":  "Compilation error",
	"13011":  "Compilation error",
	"13020":  "Compilation error",
	"13021":  "Compilation error",
	"13110":  "Compilation error",
	"13111":  "Compilation error",
	"13120":  "Compilation error",
	"13121":  "Compilation error",
	"13220":  "Compilation error",
	"13221":  "Compilation error",
	"14000":  "${c-rejected}Runtime error on pretest",
	"14001":  "${c-rejected}Runtime error on pretest",
	"14010":  "${c-rejected}Runtime error on pretest ${f-judged}",
	"14011":  "${c-rejected}Runtime error on pretest ${f-judged}",
	"14020":  "${c-rejected}Runtime error on pretest ${f-judged}",
	"14021":  "${c-rejected}Runtime error on pretest ${f-judged}",
	"14110":  "${c-rejected}Runtime error on pretest ${f-judged}",
	"14111":  "${c-rejected}Runtime error on pretest ${f-judged}",
	"14120":  "${c-rejected}Runtime error on pretest ${f-judged}",
	"14121":  "${c-rejected}Runtime error on pretest ${f-judged}",
	"14220":  "${c-rejected}Runtime error on pretest ${f-judged}",
	"14221":  "${c-rejected}Runtime error on pretest ${f-judged}",
	"15000":  "${c-rejected}Wrong answer on pretest",
	"15001":  "${c-rejected}Wrong answer on pretest",
	"15010":  "${c-rejected}Wrong answer on pretest ${f-judged}",
	"15011":  "${c-rejected}Wrong answer on pretest ${f-judged}",
	"15020":  "${c-rejected}Wrong answer on pretest ${f-judged}",
	"15021":  "${c-rejected}Wrong answer on pretest ${f-judged}",
	"15110":  "${c-rejected}Wrong answer on pretest ${f-judged}",
	"15111":  "${c-rejected}Wrong answer on pretest ${f-judged}",
	"15120":  "${c-rejected}Wrong answer on pretest ${f-judged}",
	"15121":  "${c-rejected}Wrong answer on pretest ${f-judged}",
	"15220":  "${c-rejected}Wrong answer on pretest ${f-judged}",
	"15221":  "${c-rejected}Wrong answer on pretest ${f-judged}",
	"16000":  "${c-rejected}Presentation error on pretest",
	"16001":  "${c-rejected}Presentation error on pretest",
	"16010":  "${c-rejected}Presentation error on pretest ${f-judged}",
	"16011":  "${c-rejected}Presentation error on pretest ${f-judged}",
	"16020":  "${c-rejected}Presentation error on pretest ${f-judged}",
	"16021":  "${c-rejected}Presentation error on pretest ${f-judged}",
	"16110":  "${c-rejected}Presentation error on pretest ${f-judged}",
	"16111":  "${c-rejected}Presentation error on pretest ${f-judged}",
	"16120":  "${c-rejected}Presentation error on pretest ${f-judged}",
	"16121":  "${c-rejected}Presentation error on pretest ${f-judged}",
	"16220":  "${c-rejected}Presentation error on pretest ${f-judged}",
	"16221":  "${c-rejected}Presentation error on pretest ${f-judged}",
	"17000":  "${c-rejected}Time limit exceeded on pretest",
	"17001":  "${c-rejected}Time limit exceeded on pretest",
	"17010":  "${c-rejected}Time limit exceeded on pretest ${f-judged}",
	"17011":  "${c-rejected}Time limit exceeded on pretest ${f-judged}",
	"17020":  "${c-rejected}Time limit exceeded on pretest ${f-judged}",
	"17021":  "${c-rejected}Time limit exceeded on pretest ${f-judged}",
	"17110":  "${c-rejected}Time limit exceeded on pretest ${f-judged}",
	"17111":  "${c-rejected}Time limit exceeded on pretest ${f-judged}",
	"17120":  "${c-rejected}Time limit exceeded on pretest ${f-judged}",
	"17121":  "${c-rejected}Time limit exceeded on pretest ${f-judged}",
	"17220":  "${c-rejected}Time limit exceeded on pretest ${f-judged}",
	"17221":  "${c-rejected}Time limit exceeded on pretest ${f-judged}",
	"18000":  "${c-rejected}Memory limit exceeded on pretest",
	"18001":  "${c-rejected}Memory limit exceeded on pretest",
	"18010":  "${c-rejected}Memory limit exceeded on pretest ${f-judged}",
	"18011":  "${c-rejected}Memory limit exceeded on pretest ${f-judged}",
	"18020":  "${c-rejected}Memory limit exceeded on pretest ${f-judged}",
	"18021":  "${c-rejected}Memory limit exceeded on pretest ${f-judged}",
	"18110":  "${c-rejected}Memory limit exceeded on pretest ${f-judged}",
	"18111":  "${c-rejected}Memory limit exceeded on pretest ${f-judged}",
	"18120":  "${c-rejected}Memory limit exceeded on pretest ${f-judged}",
	"18121":  "${c-rejected}Memory limit exceeded on pretest ${f-judged}",
	"18220":  "${c-rejected}Memory limit exceeded on pretest ${f-judged}",
	"18221":  "${c-rejected}Memory limit exceeded on pretest ${f-judged}",
	"19000":  "${c-rejected}Idleness limit exceeded on pretest",
	"19001":  "${c-rejected}Idleness limit exceeded on pretest",
	"19010":  "${c-rejected}Idleness limit exceeded on pretest ${f-judged}",
	"19011":  "${c-rejected}Idleness limit exceeded on pretest ${f-judged}",
	"19020":  "${c-rejected}Idleness limit exceeded on pretest ${f-judged}",
	"19021":  "${c-rejected}Idleness limit exceeded on pretest ${f-judged}",
	"19110":  "${c-rejected}Idleness limit exceeded on pretest ${f-judged}",
	"19111":  "${c-rejected}Idleness limit exceeded on pretest ${f-judged}",
	"19120":  "${c-rejected}Idleness limit exceeded on pretest ${f-judged}",
	"19121":  "${c-rejected}Idleness limit exceeded on pretest ${f-judged}",
	"19220":  "${c-rejected}Idleness limit exceeded on pretest ${f-judged}",
	"19221":  "${c-rejected}Idleness limit exceeded on pretest ${f-judged}",
	"110000": "${c-rejected}Security violated on pretest",
	"110001": "${c-rejected}Security violated on pretest",
	"110010": "${c-rejected}Security violated on pretest ${f-judged}",
	"110011": "${c-rejected}Security violated on pretest ${f-judged}",
	"110020": "${c-rejected}Security violated on pretest ${f-judged}",
	"110021": "${c-rejected}Security violated on pretest ${f-judged}",
	"110110": "${c-rejected}Security violated on pretest ${f-judged}",
	"110111": "${c-rejected}Security violated on pretest ${f-judged}",
	"110120": "${c-rejected}Security violated on pretest ${f-judged}",
	"110121": "${c-rejected}Security violated on pretest ${f-judged}",
	"110220": "${c-rejected}Security violated on pretest ${f-judged}",
	"110221": "${c-rejected}Security violated on pretest ${f-judged}",
	"111000": "${c-failed}Denial of judgement",
	"111001": "${c-failed}Denial of judgement",
	"111010": "${c-failed}Denial of judgement",
	"111011": "${c-failed}Denial of judgement",
	"111020": "${c-failed}Denial of judgement",
	"111021": "${c-failed}Denial of judgement",
	"111110": "${c-failed}Denial of judgement",
	"111111": "${c-failed}Denial of judgement",
	"111120": "${c-failed}Denial of judgement",
	"111121": "${c-failed}Denial of judgement",
	"111220": "${c-failed}Denial of judgement",
	"111221": "${c-failed}Denial of judgement",
	"112000": "${c-failed}Input preparation failed on pretest",
	"112001": "${c-failed}Input preparation failed on pretest",
	"112010": "${c-failed}Input preparation failed on pretest ${f-judged}",
	"112011": "${c-failed}Input preparation failed on pretest ${f-judged}",
	"112020": "${c-failed}Input preparation failed on pretest ${f-judged}",
	"112021": "${c-failed}Input preparation failed on pretest ${f-judged}",
	"112110": "${c-failed}Input preparation failed on pretest ${f-judged}",
	"112111": "${c-failed}Input preparation failed on pretest ${f-judged}",
	"112120": "${c-failed}Input preparation failed on pretest ${f-judged}",
	"112121": "${c-failed}Input preparation failed on pretest ${f-judged}",
	"112220": "${c-failed}Input preparation failed on pretest ${f-judged}",
	"112221": "${c-failed}Input preparation failed on pretest ${f-judged}",
	"113000": "${c-failed}Hacked",
	"113001": "${c-failed}Hacked",
	"113010": "${c-failed}Hacked",
	"113011": "${c-failed}Hacked",
	"113020": "${c-failed}Hacked",
	"113021": "${c-failed}Hacked",
	"113110": "${c-failed}Hacked",
	"113111": "${c-failed}Hacked",
	"113120": "${c-failed}Hacked",
	"113121": "${c-failed}Hacked",
	"113220": "${c-failed}Hacked",
	"113221": "${c-failed}Hacked",
	"114000": "Skipped",
	"114001": "Skipped",
	"114010": "Skipped",
	"114011": "Skipped",
	"114020": "Skipped",
	"114021": "Skipped",
	"114110": "Skipped",
	"114111": "Skipped",
	"114120": "Skipped",
	"114121": "Skipped",
	"114220": "Skipped",
	"114221": "Skipped",
	"115000": "${c-waiting}Running on pretest",
	"115001": "${c-waiting}Running on pretest",
	"115010": "${c-waiting}Running on pretest ${f-judged}",
	"115011": "${c-waiting}Running on pretest ${f-judged}",
	"115020": "${c-waiting}Running on pretest ${f-judged}",
	"115021": "${c-waiting}Running on pretest ${f-judged}",
	"115110": "${c-waiting}Running on pretest ${f-judged}",
	"115111": "${c-waiting}Running on pretest ${f-judged}",
	"115120": "${c-waiting}Running on pretest ${f-judged}",
	"115121": "${c-waiting}Running on pretest ${f-judged}",
	"115220": "${c-waiting}Running on pretest ${f-judged}",
	"115221": "${c-waiting}Running on pretest ${f-judged}",
	"116000": "${c-rejected}Rejected on pretest",
	"116001": "${c-rejected}Rejected on pretest",
	"116010": "${c-rejected}Rejected on pretest ${f-judged}",
	"116011": "${c-rejected}Rejected on pretest ${f-judged}",
	"116020": "${c-rejected}Rejected on pretest ${f-judged}",
	"116021": "${c-rejected}Rejected on pretest ${f-judged}",
	"116110": "${c-rejected}Rejected on pretest ${f-judged}",
	"116111": "${c-rejected}Rejected on pretest ${f-judged}",
	"116120": "${c-rejected}Rejected on pretest ${f-judged}",
	"116121": "${c-rejected}Rejected on pretest ${f-judged}",
	"116220": "${c-rejected}Rejected on pretest ${f-judged}",
	"116221": "${c-rejected}Rejected on pretest ${f-judged}",
	"117000": "StatusForPretest::submitted",
	"117001": "StatusForPretest::submitted",
	"117010": "StatusForPretest::submitted 1 tests / =1",
	"117011": "StatusForPretest::submitted 1 tests / =1",
	"117020": "StatusForPretest::submitted 2 tests / 2-4",
	"117021": "StatusForPretest::submitted 2 tests / 2-4",
	"117110": "StatusForPretest::submitted 1 tests / =1",
	"117111": "StatusForPretest::submitted 1 tests / =1",
	"117120": "StatusForPretest::submitted 2 tests / 2-4",
	"117121": "StatusForPretest::submitted 2 tests / 2-4",
	"117220": "StatusForPretest::submitted 2 tests / 2-4",
	"117221": "StatusForPretest::submitted 2 tests / 2-4",
	"1-000":  "${c-waiting}In queue",
	"1-001":  "${c-waiting}In queue",
	"1-010":  "${c-waiting}In queue",
	"1-011":  "${c-waiting}In queue",
	"1-020":  "${c-waiting}In queue",
	"1-021":  "${c-waiting}In queue",
	"1-110":  "${c-waiting}In queue",
	"1-111":  "${c-waiting}In queue",
	"1-120":  "${c-waiting}In queue",
	"1-121":  "${c-waiting}In queue",
	"1-220":  "${c-waiting}In queue",
	"1-221":  "${c-waiting}In queue",
	"20000":  "${c-failed}Judgement failed",
	"20001":  "${c-failed}Judgement failed",
	"20010":  "${c-failed}Judgement failed",
	"20011":  "${c-failed}Judgement failed",
	"20020":  "${c-failed}Judgement failed",
	"20021":  "${c-failed}Judgement failed",
	"20110":  "${c-failed}Judgement failed",
	"20111":  "${c-failed}Judgement failed",
	"20120":  "${c-failed}Judgement failed",
	"20121":  "${c-failed}Judgement failed",
	"20220":  "${c-failed}Judgement failed",
	"20221":  "${c-failed}Judgement failed",
	"21000":  "${c-accepted}Accepted",
	"21001":  "${c-accepted}Perfect result: ${f-points} points",
	"21010":  "${c-accepted}Accepted",
	"21011":  "${c-accepted}Perfect result: ${f-points} points",
	"21020":  "${c-accepted}Accepted",
	"21021":  "${c-accepted}Perfect result: ${f-points} points",
	"21110":  "${c-accepted}Accepted",
	"21111":  "${c-accepted}Perfect result: ${f-points} points",
	"21120":  "${c-accepted}Accepted",
	"21121":  "${c-accepted}Perfect result: ${f-points} points",
	"21220":  "${c-accepted}Accepted",
	"21221":  "${c-accepted}Perfect result: ${f-points} points",
	"22000":  "Partial",
	"22001":  "Partial result: ${f-points} points",
	"22010":  "Partial: ${f-passed} tests out of ${f-judged}",
	"22011":  "Partial result: ${f-points} points",
	"22020":  "Partial: ${f-passed} tests out of ${f-judged}",
	"22021":  "Partial result: ${f-points} points",
	"22110":  "Partial: ${f-passed} tests out of ${f-judged}",
	"22111":  "Partial result: ${f-points} points",
	"22120":  "Partial: ${f-passed} tests out of ${f-judged}",
	"22121":  "Partial result: ${f-points} points",
	"22220":  "Partial: ${f-passed} tests out of ${f-judged}",
	"22221":  "Partial result: ${f-points} points",
	"23000":  "Compilation error",
	"23001":  "Compilation error",
	"23010":  "Compilation error",
	"23011":  "Compilation error",
	"23020":  "Compilation error",
	"23021":  "Compilation error",
	"23110":  "Compilation error",
	"23111":  "Compilation error",
	"23120":  "Compilation error",
	"23121":  "Compilation error",
	"23220":  "Compilation error",
	"23221":  "Compilation error",
	"24000":  "${c-rejected}Runtime error",
	"24001":  "${c-rejected}Runtime error",
	"24010":  "${c-rejected}Runtime error on test ${f-judged}",
	"24011":  "${c-rejected}Runtime error on test ${f-judged}",
	"24020":  "${c-rejected}Runtime error on test ${f-judged}",
	"24021":  "${c-rejected}Runtime error on test ${f-judged}",
	"24110":  "${c-rejected}Runtime error on test ${f-judged}",
	"24111":  "${c-rejected}Runtime error on test ${f-judged}",
	"24120":  "${c-rejected}Runtime error on test ${f-judged}",
	"24121":  "${c-rejected}Runtime error on test ${f-judged}",
	"24220":  "${c-rejected}Runtime error on test ${f-judged}",
	"24221":  "${c-rejected}Runtime error on test ${f-judged}",
	"25000":  "${c-rejected}Wrong answer",
	"25001":  "${c-rejected}Wrong answer",
	"25010":  "${c-rejected}Wrong answer on test ${f-judged}",
	"25011":  "${c-rejected}Wrong answer on test ${f-judged}",
	"25020":  "${c-rejected}Wrong answer on test ${f-judged}",
	"25021":  "${c-rejected}Wrong answer on test ${f-judged}",
	"25110":  "${c-rejected}Wrong answer on test ${f-judged}",
	"25111":  "${c-rejected}Wrong answer on test ${f-judged}",
	"25120":  "${c-rejected}Wrong answer on test ${f-judged}",
	"25121":  "${c-rejected}Wrong answer on test ${f-judged}",
	"25220":  "${c-rejected}Wrong answer on test ${f-judged}",
	"25221":  "${c-rejected}Wrong answer on test ${f-judged}",
	"26000":  "${c-rejected}Presentation error",
	"26001":  "${c-rejected}Presentation error",
	"26010":  "${c-rejected}Presentation error on test ${f-judged}",
	"26011":  "${c-rejected}Presentation error on test ${f-judged}",
	"26020":  "${c-rejected}Presentation error on test ${f-judged}",
	"26021":  "${c-rejected}Presentation error on test ${f-judged}",
	"26110":  "${c-rejected}Presentation error on test ${f-judged}",
	"26111":  "${c-rejected}Presentation error on test ${f-judged}",
	"26120":  "${c-rejected}Presentation error on test ${f-judged}",
	"26121":  "${c-rejected}Presentation error on test ${f-judged}",
	"26220":  "${c-rejected}Presentation error on test ${f-judged}",
	"26221":  "${c-rejected}Presentation error on test ${f-judged}",
	"27000":  "${c-rejected}Time limit exceeded",
	"27001":  "${c-rejected}Time limit exceeded",
	"27010":  "${c-rejected}Time limit exceeded on test ${f-judged}",
	"27011":  "${c-rejected}Time limit exceeded on test ${f-judged}",
	"27020":  "${c-rejected}Time limit exceeded on test ${f-judged}",
	"27021":  "${c-rejected}Time limit exceeded on test ${f-judged}",
	"27110":  "${c-rejected}Time limit exceeded on test ${f-judged}",
	"27111":  "${c-rejected}Time limit exceeded on test ${f-judged}",
	"27120":  "${c-rejected}Time limit exceeded on test ${f-judged}",
	"27121":  "${c-rejected}Time limit exceeded on test ${f-judged}",
	"27220":  "${c-rejected}Time limit exceeded on test ${f-judged}",
	"27221":  "${c-rejected}Time limit exceeded on test ${f-judged}",
	"28000":  "${c-rejected}Memory limit exceeded",
	"28001":  "${c-rejected}Memory limit exceeded",
	"28010":  "${c-rejected}Memory limit exceeded on test ${f-judged}",
	"28011":  "${c-rejected}Memory limit exceeded on test ${f-judged}",
	"28020":  "${c-rejected}Memory limit exceeded on test ${f-judged}",
	"28021":  "${c-rejected}Memory limit exceeded on test ${f-judged}",
	"28110":  "${c-rejected}Memory limit exceeded on test ${f-judged}",
	"28111":  "${c-rejected}Memory limit exceeded on test ${f-judged}",
	"28120":  "${c-rejected}Memory limit exceeded on test ${f-judged}",
	"28121":  "${c-rejected}Memory limit exceeded on test ${f-judged}",
	"28220":  "${c-rejected}Memory limit exceeded on test ${f-judged}",
	"28221":  "${c-rejected}Memory limit exceeded on test ${f-judged}",
	"29000":  "${c-rejected}Idleness limit exceeded",
	"29001":  "${c-rejected}Idleness limit exceeded",
	"29010":  "${c-rejected}Idleness limit exceeded on test ${f-judged}",
	"29011":  "${c-rejected}Idleness limit exceeded on test ${f-judged}",
	"29020":  "${c-rejected}Idleness limit exceeded on test ${f-judged}",
	"29021":  "${c-rejected}Idleness limit exceeded on test ${f-judged}",
	"29110":  "${c-rejected}Idleness limit exceeded on test ${f-judged}",
	"29111":  "${c-rejected}Idleness limit exceeded on test ${f-judged}",
	"29120":  "${c-rejected}Idleness limit exceeded on test ${f-judged}",
	"29121":  "${c-rejected}Idleness limit exceeded on test ${f-judged}",
	"29220":  "${c-rejected}Idleness limit exceeded on test ${f-judged}",
	"29221":  "${c-rejected}Idleness limit exceeded on test ${f-judged}",
	"210000": "${c-rejected}Security violated",
	"210001": "${c-rejected}Security violated",
	"210010": "${c-rejected}Security violated on test ${f-judged}",
	"210011": "${c-rejected}Security violated on test ${f-judged}",
	"210020": "${c-rejected}Security violated on test ${f-judged}",
	"210021": "${c-rejected}Security violated on test ${f-judged}",
	"210110": "${c-rejected}Security violated on test ${f-judged}",
	"210111": "${c-rejected}Security violated on test ${f-judged}",
	"210120": "${c-rejected}Security violated on test ${f-judged}",
	"210121": "${c-rejected}Security violated on test ${f-judged}",
	"210220": "${c-rejected}Security violated on test ${f-judged}",
	"210221": "${c-rejected}Security violated on test ${f-judged}",
	"211000": "${c-failed}Denial of judgement",
	"211001": "${c-failed}Denial of judgement",
	"211010": "${c-failed}Denial of judgement",
	"211011": "${c-failed}Denial of judgement",
	"211020": "${c-failed}Denial of judgement",
	"211021": "${c-failed}Denial of judgement",
	"211110": "${c-failed}Denial of judgement",
	"211111": "${c-failed}Denial of judgement",
	"211120": "${c-failed}Denial of judgement",
	"211121": "${c-failed}Denial of judgement",
	"211220": "${c-failed}Denial of judgement",
	"211221": "${c-failed}Denial of judgement",
	"212000": "${c-failed}Input preparation failed",
	"212001": "${c-failed}Input preparation failed",
	"212010": "${c-failed}Input preparation failed on test ${f-judged}",
	"212011": "${c-failed}Input preparation failed on test ${f-judged}",
	"212020": "${c-failed}Input preparation failed on test ${f-judged}",
	"212021": "${c-failed}Input preparation failed on test ${f-judged}",
	"212110": "${c-failed}Input preparation failed on test ${f-judged}",
	"212111": "${c-failed}Input preparation failed on test ${f-judged}",
	"212120": "${c-failed}Input preparation failed on test ${f-judged}",
	"212121": "${c-failed}Input preparation failed on test ${f-judged}",
	"212220": "${c-failed}Input preparation failed on test ${f-judged}",
	"212221": "${c-failed}Input preparation failed on test ${f-judged}",
	"213000": "${c-failed}Hacked",
	"213001": "${c-failed}Hacked",
	"213010": "${c-failed}Hacked",
	"213011": "${c-failed}Hacked",
	"213020": "${c-failed}Hacked",
	"213021": "${c-failed}Hacked",
	"213110": "${c-failed}Hacked",
	"213111": "${c-failed}Hacked",
	"213120": "${c-failed}Hacked",
	"213121": "${c-failed}Hacked",
	"213220": "${c-failed}Hacked",
	"213221": "${c-failed}Hacked",
	"214000": "Skipped",
	"214001": "Skipped",
	"214010": "Skipped",
	"214011": "Skipped",
	"214020": "Skipped",
	"214021": "Skipped",
	"214110": "Skipped",
	"214111": "Skipped",
	"214120": "Skipped",
	"214121": "Skipped",
	"214220": "Skipped",
	"214221": "Skipped",
	"215000": "${c-waiting}Running",
	"215001": "${c-waiting}Running",
	"215010": "${c-waiting}Running on test ${f-judged}",
	"215011": "${c-waiting}Running on test ${f-judged}",
	"215020": "${c-waiting}Running on test ${f-judged}",
	"215021": "${c-waiting}Running on test ${f-judged}",
	"215110": "${c-waiting}Running on test ${f-judged}",
	"215111": "${c-waiting}Running on test ${f-judged}",
	"215120": "${c-waiting}Running on test ${f-judged}",
	"215121": "${c-waiting}Running on test ${f-judged}",
	"215220": "${c-waiting}Running on test ${f-judged}",
	"215221": "${c-waiting}Running on test ${f-judged}",
	"216000": "${c-rejected}Rejected",
	"216001": "${c-rejected}Rejected",
	"216010": "${c-rejected}Rejected on test ${f-judged}",
	"216011": "${c-rejected}Rejected on test ${f-judged}",
	"216020": "${c-rejected}Rejected on test ${f-judged}",
	"216021": "${c-rejected}Rejected on test ${f-judged}",
	"216110": "${c-rejected}Rejected on test ${f-judged}",
	"216111": "${c-rejected}Rejected on test ${f-judged}",
	"216120": "${c-rejected}Rejected on test ${f-judged}",
	"216121": "${c-rejected}Rejected on test ${f-judged}",
	"216220": "${c-rejected}Rejected on test ${f-judged}",
	"216221": "${c-rejected}Rejected on test ${f-judged}",
	"217000": "Pending judgement",
	"217001": "Pending judgement",
	"217010": "Pending judgement of ${f-judged} test",
	"217011": "Pending judgement of ${f-judged} test",
	"217020": "Pending judgement of ${f-judged} tests",
	"217021": "Pending judgement of ${f-judged} tests",
	"217110": "Pending judgement of ${f-judged} test",
	"217111": "Pending judgement of ${f-judged} test",
	"217120": "Pending judgement of ${f-judged} tests",
	"217121": "Pending judgement of ${f-judged} tests",
	"217220": "Pending judgement of ${f-judged} tests",
	"217221": "Pending judgement of ${f-judged} tests",
	"2-000":  "${c-waiting}In queue",
	"2-001":  "${c-waiting}In queue",
	"2-010":  "${c-waiting}In queue",
	"2-011":  "${c-waiting}In queue",
	"2-020":  "${c-waiting}In queue",
	"2-021":  "${c-waiting}In queue",
	"2-110":  "${c-waiting}In queue",
	"2-111":  "${c-waiting}In queue",
	"2-120":  "${c-waiting}In queue",
	"2-121":  "${c-waiting}In queue",
	"2-220":  "${c-waiting}In queue",
	"2-221":  "${c-waiting}In queue",
}
