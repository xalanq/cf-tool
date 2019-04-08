package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/gorilla/websocket"
	ansi "github.com/k0kubun/go-ansi"
)

// Client codeforces client
type Client struct {
	Jar      *cookiejar.Jar `json:"cookies"`
	Username string         `json:"username"`
	Ftaa     string         `json:"ftaa"`
	Bfaa     string         `json:"bfaa"`
	path     string
}

// New client
func New(path string) *Client {
	jar, _ := cookiejar.New(nil)
	c := &Client{Jar: jar, path: path}
	c.load()
	return c
}

// load from path
func (c *Client) load() (err error) {
	file, err := os.Open(c.path)
	if err != nil {
		return
	}
	defer file.Close()

	bytes, err := ioutil.ReadAll(file)

	if err != nil {
		return err
	}

	return json.Unmarshal(bytes, c)
}

// save file to path
func (c *Client) save() (err error) {
	data, err := json.MarshalIndent(c, "", "  ")
	if err == nil {
		err = ioutil.WriteFile(c.path, data, 0644)
	}
	if err != nil {
		fmt.Printf("Cannot save session to %v\n%v", c.path, err.Error())
	}
	return
}

// genFtaa generate a random one
func genFtaa() string {
	const CHA = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, 18)
	for i := range b {
		b[i] = CHA[rand.Intn(len(CHA))]
	}
	return string(b)
}

// genBfaa generate a bfaa
func genBfaa() string {
	return "f1b3f18c715565b589b7823cda7448ce"
}

// ErrorNotLogged not logged in
var ErrorNotLogged = "Not logged in"

// checkLogin if login return nil
func checkLogin(username string, body []byte) error {
	match, err := regexp.Match(fmt.Sprintf(`handle = "%v"`, username), body)
	if err != nil || !match {
		return errors.New(ErrorNotLogged)
	}
	return nil
}

// findCsrf just find
func findCsrf(body []byte) (string, error) {
	reg, _ := regexp.Compile(`csrf='(.+?)'`)
	tmp := reg.FindSubmatch(body)
	if len(tmp) < 2 {
		return "", errors.New("Cannot find csrf")
	}
	return string(tmp[1]), nil
}

// Login codeforces with username(handler) and password
func (c *Client) Login(username, password string) (err error) {
	jar, _ := cookiejar.New(nil)
	fmt.Printf("Login %v...\n", username)

	client := &http.Client{Jar: jar}

	resp, err := client.Get("https://codeforces.com/enter")
	if err != nil {
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	csrf, err := findCsrf(body)
	if err != nil {
		return
	}

	ftaa := genFtaa()
	bfaa := genBfaa()

	resp, err = client.PostForm("https://codeforces.com/enter", url.Values{
		"csrf_token":    {csrf},
		"action":        {"enter"},
		"ftaa":          {ftaa},
		"bfaa":          {bfaa},
		"handleOrEmail": {username},
		"password":      {password},
		"_tta":          {"176"},
	})
	if err != nil {
		return
	}
	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	err = checkLogin(username, body)
	if err != nil {
		return
	}

	c.Jar = jar
	c.Ftaa = ftaa
	c.Bfaa = bfaa
	c.Username = username
	fmt.Println("Succeed!!")
	return c.save()
}

// SubmitState submit state
type SubmitState struct {
	name   string
	id     uint64
	state  string
	judged uint64
	time   uint64
	memory uint64
	lang   string
	end    bool
}

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
	if judgedTestCount := uint64(d[8].(float64)); judgedTestCount >= s.judged {
		s.judged = judgedTestCount
	}
	return true
}

func (s *SubmitState) display() {
	state := stateToText[s.state]
	if !s.end && s.judged > 0 {
		state = fmt.Sprintf("case %v", s.judged)
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
	fmt.Printf("      #: %v\n", s.id)
	fmt.Printf("   prob: %v\n", s.name)
	fmt.Printf("                                         \n")
	ansi.CursorUp(1)
	fmt.Printf("  state: ")
	if s.end {
		if state == "Accepted" || state == "Pretests passed" {
			color.Green(state + "\n")
		} else if state != "Compilation error" {
			color.Blue(state + "\n")
		} else {
			fmt.Printf("%v\n", state)
		}
	} else {
		fmt.Printf("%v\n", state)
	}
	fmt.Printf("   lang: %v\n", s.lang)
	fmt.Printf("   time: %v ms\n", s.time)
	fmt.Printf(" memory: %v\n", memory)
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

var stateToText = map[string]string{
	"---":    "In queue",
	"30000":  "Denial of judgement",
	"30001":  "Denial of judgement",
	"30010":  "Denial of judgement",
	"30011":  "Denial of judgement",
	"30020":  "Denial of judgement",
	"30021":  "Denial of judgement",
	"30110":  "Denial of judgement",
	"30111":  "Denial of judgement",
	"30120":  "Denial of judgement",
	"30121":  "Denial of judgement",
	"30220":  "Denial of judgement",
	"30221":  "Denial of judgement",
	"31000":  "Pretests and hacks passed",
	"31001":  "Perfect result: 1 points",
	"31010":  "Pretests and hacks passed",
	"31011":  "Perfect result: 1 points",
	"31020":  "Pretests and hacks passed",
	"31021":  "Perfect result: 1 points",
	"31110":  "Pretests and hacks passed",
	"31111":  "Perfect result: 1 points",
	"31120":  "Pretests and hacks passed",
	"31121":  "Perfect result: 1 points",
	"31220":  "Pretests and hacks passed",
	"31221":  "Perfect result: 1 points",
	"32000":  "Partial (hacks)",
	"32001":  "Partial result: 1 points",
	"32010":  "Partial: 0 hacks ouf of 1",
	"32011":  "Partial result: 1 points",
	"32020":  "Partial: 0 hacks ouf of 2",
	"32021":  "Partial result: 1 points",
	"32110":  "Partial: 1 hacks ouf of 1",
	"32111":  "Partial result: 1 points",
	"32120":  "Partial: 1 hacks ouf of 2",
	"32121":  "Partial result: 1 points",
	"32220":  "Partial: 2 hacks ouf of 2",
	"32221":  "Partial result: 1 points",
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
	"34000":  "Runtime error on hack",
	"34001":  "Runtime error on hack",
	"34010":  "Runtime error on hack 1",
	"34011":  "Runtime error on hack 1",
	"34020":  "Runtime error on hack 2",
	"34021":  "Runtime error on hack 2",
	"34110":  "Runtime error on hack 1",
	"34111":  "Runtime error on hack 1",
	"34120":  "Runtime error on hack 2",
	"34121":  "Runtime error on hack 2",
	"34220":  "Runtime error on hack 2",
	"34221":  "Runtime error on hack 2",
	"35000":  "Wrong answer on hack",
	"35001":  "Wrong answer on hack",
	"35010":  "Wrong answer on hack 1",
	"35011":  "Wrong answer on hack 1",
	"35020":  "Wrong answer on hack 2",
	"35021":  "Wrong answer on hack 2",
	"35110":  "Wrong answer on hack 1",
	"35111":  "Wrong answer on hack 1",
	"35120":  "Wrong answer on hack 2",
	"35121":  "Wrong answer on hack 2",
	"35220":  "Wrong answer on hack 2",
	"35221":  "Wrong answer on hack 2",
	"36000":  "Presentation error on hack",
	"36001":  "Presentation error on hack",
	"36010":  "Presentation error on hack 1",
	"36011":  "Presentation error on hack 1",
	"36020":  "Presentation error on hack 2",
	"36021":  "Presentation error on hack 2",
	"36110":  "Presentation error on hack 1",
	"36111":  "Presentation error on hack 1",
	"36120":  "Presentation error on hack 2",
	"36121":  "Presentation error on hack 2",
	"36220":  "Presentation error on hack 2",
	"36221":  "Presentation error on hack 2",
	"37000":  "Time limit exceeded on hack",
	"37001":  "Time limit exceeded on hack",
	"37010":  "Time limit exceeded on hack 1",
	"37011":  "Time limit exceeded on hack 1",
	"37020":  "Time limit exceeded on hack 2",
	"37021":  "Time limit exceeded on hack 2",
	"37110":  "Time limit exceeded on hack 1",
	"37111":  "Time limit exceeded on hack 1",
	"37120":  "Time limit exceeded on hack 2",
	"37121":  "Time limit exceeded on hack 2",
	"37220":  "Time limit exceeded on hack 2",
	"37221":  "Time limit exceeded on hack 2",
	"38000":  "Memory limit exceeded on hack",
	"38001":  "Memory limit exceeded on hack",
	"38010":  "Memory limit exceeded on hack 1",
	"38011":  "Memory limit exceeded on hack 1",
	"38020":  "Memory limit exceeded on hack 2",
	"38021":  "Memory limit exceeded on hack 2",
	"38110":  "Memory limit exceeded on hack 1",
	"38111":  "Memory limit exceeded on hack 1",
	"38120":  "Memory limit exceeded on hack 2",
	"38121":  "Memory limit exceeded on hack 2",
	"38220":  "Memory limit exceeded on hack 2",
	"38221":  "Memory limit exceeded on hack 2",
	"39000":  "Idleness limit exceeded on hack",
	"39001":  "Idleness limit exceeded on hack",
	"39010":  "Idleness limit exceeded on hack 1",
	"39011":  "Idleness limit exceeded on hack 1",
	"39020":  "Idleness limit exceeded on hack 2",
	"39021":  "Idleness limit exceeded on hack 2",
	"39110":  "Idleness limit exceeded on hack 1",
	"39111":  "Idleness limit exceeded on hack 1",
	"39120":  "Idleness limit exceeded on hack 2",
	"39121":  "Idleness limit exceeded on hack 2",
	"39220":  "Idleness limit exceeded on hack 2",
	"39221":  "Idleness limit exceeded on hack 2",
	"310000": "Security violated on hack",
	"310001": "Security violated on hack",
	"310010": "Security violated on hack 1",
	"310011": "Security violated on hack 1",
	"310020": "Security violated on hack 2",
	"310021": "Security violated on hack 2",
	"310110": "Security violated on hack 1",
	"310111": "Security violated on hack 1",
	"310120": "Security violated on hack 2",
	"310121": "Security violated on hack 2",
	"310220": "Security violated on hack 2",
	"310221": "Security violated on hack 2",
	"311000": "Judgement crashed on hack",
	"311001": "Judgement crashed on hack",
	"311010": "Judgement crashed on hack",
	"311011": "Judgement crashed on hack",
	"311020": "Judgement crashed on hack",
	"311021": "Judgement crashed on hack",
	"311110": "Judgement crashed on hack",
	"311111": "Judgement crashed on hack",
	"311120": "Judgement crashed on hack",
	"311121": "Judgement crashed on hack",
	"311220": "Judgement crashed on hack",
	"311221": "Judgement crashed on hack",
	"312000": "Input preparation failed on hack",
	"312001": "Input preparation failed on hack",
	"312010": "Input preparation failed on hack 1",
	"312011": "Input preparation failed on hack 1",
	"312020": "Input preparation failed on hack 2",
	"312021": "Input preparation failed on hack 2",
	"312110": "Input preparation failed on hack 1",
	"312111": "Input preparation failed on hack 1",
	"312120": "Input preparation failed on hack 2",
	"312121": "Input preparation failed on hack 2",
	"312220": "Input preparation failed on hack 2",
	"312221": "Input preparation failed on hack 2",
	"313000": "Hacked",
	"313001": "Hacked",
	"313010": "Hacked",
	"313011": "Hacked",
	"313020": "Hacked",
	"313021": "Hacked",
	"313110": "Hacked",
	"313111": "Hacked",
	"313120": "Hacked",
	"313121": "Hacked",
	"313220": "Hacked",
	"313221": "Hacked",
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
	"315000": "Running on hack",
	"315001": "Running on hack",
	"315010": "Running on hack 1",
	"315011": "Running on hack 1",
	"315020": "Running on hack 2",
	"315021": "Running on hack 2",
	"315110": "Running on hack 1",
	"315111": "Running on hack 1",
	"315120": "Running on hack 2",
	"315121": "Running on hack 2",
	"315220": "Running on hack 2",
	"315221": "Running on hack 2",
	"316000": "Rejected on hack",
	"316001": "Rejected on hack",
	"316010": "Rejected on hack 1",
	"316011": "Rejected on hack 1",
	"316020": "Rejected on hack 2",
	"316021": "Rejected on hack 2",
	"316110": "Rejected on hack 1",
	"316111": "Rejected on hack 1",
	"316120": "Rejected on hack 2",
	"316121": "Rejected on hack 2",
	"316220": "Rejected on hack 2",
	"316221": "Rejected on hack 2",
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
	"3-000":  "In queue",
	"3-001":  "In queue",
	"3-010":  "In queue",
	"3-011":  "In queue",
	"3-020":  "In queue",
	"3-021":  "In queue",
	"3-110":  "In queue",
	"3-111":  "In queue",
	"3-120":  "In queue",
	"3-121":  "In queue",
	"3-220":  "In queue",
	"3-221":  "In queue",
	"10000":  "Judgement Failed",
	"10001":  "Judgement Failed",
	"10010":  "Judgement Failed",
	"10011":  "Judgement Failed",
	"10020":  "Judgement Failed",
	"10021":  "Judgement Failed",
	"10110":  "Judgement Failed",
	"10111":  "Judgement Failed",
	"10120":  "Judgement Failed",
	"10121":  "Judgement Failed",
	"10220":  "Judgement Failed",
	"10221":  "Judgement Failed",
	"11000":  "Pretests passed",
	"11001":  "Perfect result: 1 points",
	"11010":  "Pretests passed",
	"11011":  "Perfect result: 1 points",
	"11020":  "Pretests passed",
	"11021":  "Perfect result: 1 points",
	"11110":  "Pretests passed",
	"11111":  "Perfect result: 1 points",
	"11120":  "Pretests passed",
	"11121":  "Perfect result: 1 points",
	"11220":  "Pretests passed",
	"11221":  "Perfect result: 1 points",
	"12000":  "Partial (pretests)",
	"12001":  "Partial result: 1 points",
	"12010":  "Partial: 0 pretests out of 1",
	"12011":  "Partial result: 1 points",
	"12020":  "Partial: 0 pretests out of 2",
	"12021":  "Partial result: 1 points",
	"12110":  "Partial: 1 pretests out of 1",
	"12111":  "Partial result: 1 points",
	"12120":  "Partial: 1 pretests out of 2",
	"12121":  "Partial result: 1 points",
	"12220":  "Partial: 2 pretests out of 2",
	"12221":  "Partial result: 1 points",
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
	"14000":  "Runtime error on pretest",
	"14001":  "Runtime error on pretest",
	"14010":  "Runtime error on pretest 1",
	"14011":  "Runtime error on pretest 1",
	"14020":  "Runtime error on pretest 2",
	"14021":  "Runtime error on pretest 2",
	"14110":  "Runtime error on pretest 1",
	"14111":  "Runtime error on pretest 1",
	"14120":  "Runtime error on pretest 2",
	"14121":  "Runtime error on pretest 2",
	"14220":  "Runtime error on pretest 2",
	"14221":  "Runtime error on pretest 2",
	"15000":  "Wrong answer on pretest",
	"15001":  "Wrong answer on pretest",
	"15010":  "Wrong answer on pretest 1",
	"15011":  "Wrong answer on pretest 1",
	"15020":  "Wrong answer on pretest 2",
	"15021":  "Wrong answer on pretest 2",
	"15110":  "Wrong answer on pretest 1",
	"15111":  "Wrong answer on pretest 1",
	"15120":  "Wrong answer on pretest 2",
	"15121":  "Wrong answer on pretest 2",
	"15220":  "Wrong answer on pretest 2",
	"15221":  "Wrong answer on pretest 2",
	"16000":  "Presentation error on pretest",
	"16001":  "Presentation error on pretest",
	"16010":  "Presentation error on pretest 1",
	"16011":  "Presentation error on pretest 1",
	"16020":  "Presentation error on pretest 2",
	"16021":  "Presentation error on pretest 2",
	"16110":  "Presentation error on pretest 1",
	"16111":  "Presentation error on pretest 1",
	"16120":  "Presentation error on pretest 2",
	"16121":  "Presentation error on pretest 2",
	"16220":  "Presentation error on pretest 2",
	"16221":  "Presentation error on pretest 2",
	"17000":  "Time limit exceeded on pretest",
	"17001":  "Time limit exceeded on pretest",
	"17010":  "Time limit exceeded on pretest 1",
	"17011":  "Time limit exceeded on pretest 1",
	"17020":  "Time limit exceeded on pretest 2",
	"17021":  "Time limit exceeded on pretest 2",
	"17110":  "Time limit exceeded on pretest 1",
	"17111":  "Time limit exceeded on pretest 1",
	"17120":  "Time limit exceeded on pretest 2",
	"17121":  "Time limit exceeded on pretest 2",
	"17220":  "Time limit exceeded on pretest 2",
	"17221":  "Time limit exceeded on pretest 2",
	"18000":  "Memory limit exceeded on pretest",
	"18001":  "Memory limit exceeded on pretest",
	"18010":  "Memory limit exceeded on pretest 1",
	"18011":  "Memory limit exceeded on pretest 1",
	"18020":  "Memory limit exceeded on pretest 2",
	"18021":  "Memory limit exceeded on pretest 2",
	"18110":  "Memory limit exceeded on pretest 1",
	"18111":  "Memory limit exceeded on pretest 1",
	"18120":  "Memory limit exceeded on pretest 2",
	"18121":  "Memory limit exceeded on pretest 2",
	"18220":  "Memory limit exceeded on pretest 2",
	"18221":  "Memory limit exceeded on pretest 2",
	"19000":  "Idleness limit exceeded on pretest",
	"19001":  "Idleness limit exceeded on pretest",
	"19010":  "Idleness limit exceeded on pretest 1",
	"19011":  "Idleness limit exceeded on pretest 1",
	"19020":  "Idleness limit exceeded on pretest 2",
	"19021":  "Idleness limit exceeded on pretest 2",
	"19110":  "Idleness limit exceeded on pretest 1",
	"19111":  "Idleness limit exceeded on pretest 1",
	"19120":  "Idleness limit exceeded on pretest 2",
	"19121":  "Idleness limit exceeded on pretest 2",
	"19220":  "Idleness limit exceeded on pretest 2",
	"19221":  "Idleness limit exceeded on pretest 2",
	"110000": "Security violated on pretest",
	"110001": "Security violated on pretest",
	"110010": "Security violated on pretest 1",
	"110011": "Security violated on pretest 1",
	"110020": "Security violated on pretest 2",
	"110021": "Security violated on pretest 2",
	"110110": "Security violated on pretest 1",
	"110111": "Security violated on pretest 1",
	"110120": "Security violated on pretest 2",
	"110121": "Security violated on pretest 2",
	"110220": "Security violated on pretest 2",
	"110221": "Security violated on pretest 2",
	"111000": "Denial of judgement",
	"111001": "Denial of judgement",
	"111010": "Denial of judgement",
	"111011": "Denial of judgement",
	"111020": "Denial of judgement",
	"111021": "Denial of judgement",
	"111110": "Denial of judgement",
	"111111": "Denial of judgement",
	"111120": "Denial of judgement",
	"111121": "Denial of judgement",
	"111220": "Denial of judgement",
	"111221": "Denial of judgement",
	"112000": "Input preparation failed on pretest",
	"112001": "Input preparation failed on pretest",
	"112010": "Input preparation failed on pretest 1",
	"112011": "Input preparation failed on pretest 1",
	"112020": "Input preparation failed on pretest 2",
	"112021": "Input preparation failed on pretest 2",
	"112110": "Input preparation failed on pretest 1",
	"112111": "Input preparation failed on pretest 1",
	"112120": "Input preparation failed on pretest 2",
	"112121": "Input preparation failed on pretest 2",
	"112220": "Input preparation failed on pretest 2",
	"112221": "Input preparation failed on pretest 2",
	"113000": "Hacked",
	"113001": "Hacked",
	"113010": "Hacked",
	"113011": "Hacked",
	"113020": "Hacked",
	"113021": "Hacked",
	"113110": "Hacked",
	"113111": "Hacked",
	"113120": "Hacked",
	"113121": "Hacked",
	"113220": "Hacked",
	"113221": "Hacked",
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
	"115000": "Running on pretest",
	"115001": "Running on pretest",
	"115010": "Running on pretest 1",
	"115011": "Running on pretest 1",
	"115020": "Running on pretest 2",
	"115021": "Running on pretest 2",
	"115110": "Running on pretest 1",
	"115111": "Running on pretest 1",
	"115120": "Running on pretest 2",
	"115121": "Running on pretest 2",
	"115220": "Running on pretest 2",
	"115221": "Running on pretest 2",
	"116000": "Rejected on pretest",
	"116001": "Rejected on pretest",
	"116010": "Rejected on pretest 1",
	"116011": "Rejected on pretest 1",
	"116020": "Rejected on pretest 2",
	"116021": "Rejected on pretest 2",
	"116110": "Rejected on pretest 1",
	"116111": "Rejected on pretest 1",
	"116120": "Rejected on pretest 2",
	"116121": "Rejected on pretest 2",
	"116220": "Rejected on pretest 2",
	"116221": "Rejected on pretest 2",
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
	"1-000":  "In queue",
	"1-001":  "In queue",
	"1-010":  "In queue",
	"1-011":  "In queue",
	"1-020":  "In queue",
	"1-021":  "In queue",
	"1-110":  "In queue",
	"1-111":  "In queue",
	"1-120":  "In queue",
	"1-121":  "In queue",
	"1-220":  "In queue",
	"1-221":  "In queue",
	"20000":  "Judgement failed",
	"20001":  "Judgement failed",
	"20010":  "Judgement failed",
	"20011":  "Judgement failed",
	"20020":  "Judgement failed",
	"20021":  "Judgement failed",
	"20110":  "Judgement failed",
	"20111":  "Judgement failed",
	"20120":  "Judgement failed",
	"20121":  "Judgement failed",
	"20220":  "Judgement failed",
	"20221":  "Judgement failed",
	"21000":  "Accepted",
	"21001":  "Perfect result: 1 points",
	"21010":  "Accepted",
	"21011":  "Perfect result: 1 points",
	"21020":  "Accepted",
	"21021":  "Perfect result: 1 points",
	"21110":  "Accepted",
	"21111":  "Perfect result: 1 points",
	"21120":  "Accepted",
	"21121":  "Perfect result: 1 points",
	"21220":  "Accepted",
	"21221":  "Perfect result: 1 points",
	"22000":  "Partial",
	"22001":  "Partial result: 1 points",
	"22010":  "Partial: 0 tests out of 1",
	"22011":  "Partial result: 1 points",
	"22020":  "Partial: 0 tests out of 2",
	"22021":  "Partial result: 1 points",
	"22110":  "Partial: 1 tests out of 1",
	"22111":  "Partial result: 1 points",
	"22120":  "Partial: 1 tests out of 2",
	"22121":  "Partial result: 1 points",
	"22220":  "Partial: 2 tests out of 2",
	"22221":  "Partial result: 1 points",
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
	"24000":  "Runtime error",
	"24001":  "Runtime error",
	"24010":  "Runtime error on test 1",
	"24011":  "Runtime error on test 1",
	"24020":  "Runtime error on test 2",
	"24021":  "Runtime error on test 2",
	"24110":  "Runtime error on test 1",
	"24111":  "Runtime error on test 1",
	"24120":  "Runtime error on test 2",
	"24121":  "Runtime error on test 2",
	"24220":  "Runtime error on test 2",
	"24221":  "Runtime error on test 2",
	"25000":  "Wrong answer",
	"25001":  "Wrong answer",
	"25010":  "Wrong answer on test 1",
	"25011":  "Wrong answer on test 1",
	"25020":  "Wrong answer on test 2",
	"25021":  "Wrong answer on test 2",
	"25110":  "Wrong answer on test 1",
	"25111":  "Wrong answer on test 1",
	"25120":  "Wrong answer on test 2",
	"25121":  "Wrong answer on test 2",
	"25220":  "Wrong answer on test 2",
	"25221":  "Wrong answer on test 2",
	"26000":  "Presentation error",
	"26001":  "Presentation error",
	"26010":  "Presentation error on test 1",
	"26011":  "Presentation error on test 1",
	"26020":  "Presentation error on test 2",
	"26021":  "Presentation error on test 2",
	"26110":  "Presentation error on test 1",
	"26111":  "Presentation error on test 1",
	"26120":  "Presentation error on test 2",
	"26121":  "Presentation error on test 2",
	"26220":  "Presentation error on test 2",
	"26221":  "Presentation error on test 2",
	"27000":  "Time limit exceeded",
	"27001":  "Time limit exceeded",
	"27010":  "Time limit exceeded on test 1",
	"27011":  "Time limit exceeded on test 1",
	"27020":  "Time limit exceeded on test 2",
	"27021":  "Time limit exceeded on test 2",
	"27110":  "Time limit exceeded on test 1",
	"27111":  "Time limit exceeded on test 1",
	"27120":  "Time limit exceeded on test 2",
	"27121":  "Time limit exceeded on test 2",
	"27220":  "Time limit exceeded on test 2",
	"27221":  "Time limit exceeded on test 2",
	"28000":  "Memory limit exceeded",
	"28001":  "Memory limit exceeded",
	"28010":  "Memory limit exceeded on test 1",
	"28011":  "Memory limit exceeded on test 1",
	"28020":  "Memory limit exceeded on test 2",
	"28021":  "Memory limit exceeded on test 2",
	"28110":  "Memory limit exceeded on test 1",
	"28111":  "Memory limit exceeded on test 1",
	"28120":  "Memory limit exceeded on test 2",
	"28121":  "Memory limit exceeded on test 2",
	"28220":  "Memory limit exceeded on test 2",
	"28221":  "Memory limit exceeded on test 2",
	"29000":  "Idleness limit exceeded",
	"29001":  "Idleness limit exceeded",
	"29010":  "Idleness limit exceeded on test 1",
	"29011":  "Idleness limit exceeded on test 1",
	"29020":  "Idleness limit exceeded on test 2",
	"29021":  "Idleness limit exceeded on test 2",
	"29110":  "Idleness limit exceeded on test 1",
	"29111":  "Idleness limit exceeded on test 1",
	"29120":  "Idleness limit exceeded on test 2",
	"29121":  "Idleness limit exceeded on test 2",
	"29220":  "Idleness limit exceeded on test 2",
	"29221":  "Idleness limit exceeded on test 2",
	"210000": "Security violated",
	"210001": "Security violated",
	"210010": "Security violated on test 1",
	"210011": "Security violated on test 1",
	"210020": "Security violated on test 2",
	"210021": "Security violated on test 2",
	"210110": "Security violated on test 1",
	"210111": "Security violated on test 1",
	"210120": "Security violated on test 2",
	"210121": "Security violated on test 2",
	"210220": "Security violated on test 2",
	"210221": "Security violated on test 2",
	"211000": "Denial of judgement",
	"211001": "Denial of judgement",
	"211010": "Denial of judgement",
	"211011": "Denial of judgement",
	"211020": "Denial of judgement",
	"211021": "Denial of judgement",
	"211110": "Denial of judgement",
	"211111": "Denial of judgement",
	"211120": "Denial of judgement",
	"211121": "Denial of judgement",
	"211220": "Denial of judgement",
	"211221": "Denial of judgement",
	"212000": "Input preparation failed",
	"212001": "Input preparation failed",
	"212010": "Input preparation failed on test 1",
	"212011": "Input preparation failed on test 1",
	"212020": "Input preparation failed on test 2",
	"212021": "Input preparation failed on test 2",
	"212110": "Input preparation failed on test 1",
	"212111": "Input preparation failed on test 1",
	"212120": "Input preparation failed on test 2",
	"212121": "Input preparation failed on test 2",
	"212220": "Input preparation failed on test 2",
	"212221": "Input preparation failed on test 2",
	"213000": "Hacked",
	"213001": "Hacked",
	"213010": "Hacked",
	"213011": "Hacked",
	"213020": "Hacked",
	"213021": "Hacked",
	"213110": "Hacked",
	"213111": "Hacked",
	"213120": "Hacked",
	"213121": "Hacked",
	"213220": "Hacked",
	"213221": "Hacked",
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
	"215000": "Running",
	"215001": "Running",
	"215010": "Running on test 1",
	"215011": "Running on test 1",
	"215020": "Running on test 2",
	"215021": "Running on test 2",
	"215110": "Running on test 1",
	"215111": "Running on test 1",
	"215120": "Running on test 2",
	"215121": "Running on test 2",
	"215220": "Running on test 2",
	"215221": "Running on test 2",
	"216000": "Rejected",
	"216001": "Rejected",
	"216010": "Rejected on test 1",
	"216011": "Rejected on test 1",
	"216020": "Rejected on test 2",
	"216021": "Rejected on test 2",
	"216110": "Rejected on test 1",
	"216111": "Rejected on test 1",
	"216120": "Rejected on test 2",
	"216121": "Rejected on test 2",
	"216220": "Rejected on test 2",
	"216221": "Rejected on test 2",
	"217000": "Pending judgement",
	"217001": "Pending judgement",
	"217010": "Pending judgement of 1 test",
	"217011": "Pending judgement of 1 test",
	"217020": "Pending judgement of 2 tests",
	"217021": "Pending judgement of 2 tests",
	"217110": "Pending judgement of 1 test",
	"217111": "Pending judgement of 1 test",
	"217120": "Pending judgement of 2 tests",
	"217121": "Pending judgement of 2 tests",
	"217220": "Pending judgement of 2 tests",
	"217221": "Pending judgement of 2 tests",
	"2-000":  "In queue",
	"2-001":  "In queue",
	"2-010":  "In queue",
	"2-011":  "In queue",
	"2-020":  "In queue",
	"2-021":  "In queue",
	"2-110":  "In queue",
	"2-111":  "In queue",
	"2-120":  "In queue",
	"2-121":  "In queue",
	"2-220":  "In queue",
	"2-221":  "In queue",
}
