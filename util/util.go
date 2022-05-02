package util

import (
	"bufio"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"regexp"
	"strings"
	"crypto/aes"
	"crypto/cipher"

	"github.com/fatih/color"
)

// CHA map
const CHA = "abcdefghijklmnopqrstuvwxyz0123456789"

// RandString n is the length. a-z 0-9
func RandString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = CHA[rand.Intn(len(CHA))]
	}
	return string(b)
}

// Scanline scan line
func Scanline() string {
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		return scanner.Text()
	}
	fmt.Println("\nInterrupted.")
	os.Exit(1)
	return ""
}

// ScanlineTrim scan line and trim
func ScanlineTrim() string {
	return strings.TrimSpace(Scanline())
}

// ChooseIndex return valid index in [0, maxLen)
func ChooseIndex(maxLen int) int {
	color.Cyan("Please choose one (index): ")
	for {
		index := ScanlineTrim()
		i, err := strconv.Atoi(index)
		if err == nil && i >= 0 && i < maxLen {
			return i
		}
		color.Red("Invalid index! Please try again: ")
	}
}

// YesOrNo must choose one
func YesOrNo(note string) bool {
	color.Cyan(note)
	for {
		tmp := ScanlineTrim()
		if tmp == "y" || tmp == "Y" {
			return true
		}
		if tmp == "n" || tmp == "N" {
			return false
		}
		color.Red("Invalid input. Please input again: ")
	}
}

// GetBody read body
func SetRCPC(client *http.Client, body []byte, URL string) ([]byte, error) {
	reg := regexp.MustCompile(`toNumbers\("(.+?)"\)`)
	res := reg.FindAllStringSubmatch(string(body), -1)
	text, _ := hex.DecodeString( res[2][1] )
	key, _ := hex.DecodeString( res[0][1] )
	iv, _ := hex.DecodeString( res[1][1] )

	block, _ := aes.NewCipher(key)
	mode := cipher.NewCBCDecrypter(block, iv)

	mode.CryptBlocks([]byte(text), []byte(text))

	var cookies []*http.Cookie
	cookie := &http.Cookie{
		Name:   "RCPC",
		Value:  hex.EncodeToString(text),
		Path:   "/",
		Domain: ".codeforces.com",
	}
	cookies = append(cookies, cookie)
	u, _ := url.Parse("https://codeforces.com/")
	client.Jar.SetCookies(u, cookies)

	reg = regexp.MustCompile(`href="(.+?)"`)
	link := reg.FindSubmatch(body)[1]
	return GetBody( client, string(link) )
}

// GetBody read body
func GetBody(client *http.Client, URL string) ([]byte, error) {
	resp, err := client.Get(URL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body);
	reg := regexp.MustCompile(`Redirecting...`)
	is_redirected := ( len( reg.FindSubmatch(body) ) > 0 );

	if is_redirected {
		return SetRCPC(client, body, URL)
	}
	return body, err
}

// PostBody read post body
func PostBody(client *http.Client, URL string, data url.Values) ([]byte, error) {
	resp, err := client.PostForm(URL, data)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

// GetJSONBody read json body
func GetJSONBody(client *http.Client, url string) (map[string]interface{}, error) {
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	var data map[string]interface{}
	if err = decoder.Decode(&data); err != nil {
		return nil, err
	}
	return data, nil
}

// DebugSave write data to temperory file
func DebugSave(data interface{}) {
	f, err := os.OpenFile("./tmp/body", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	if data, ok := data.([]byte); ok {
		if _, err := f.Write(data); err != nil {
			log.Fatal(err)
		}
	} else {
		if _, err := f.Write([]byte(fmt.Sprintf("%v\n\n", data))); err != nil {
			log.Fatal(err)
		}
	}
	if err := f.Close(); err != nil {
		log.Fatal(err)
	}
}

// DebugJSON debug
func DebugJSON(data interface{}) {
	text, _ := json.MarshalIndent(data, "", "  ")
	fmt.Println(string(text))
}

// IsURL returns true if a given string is an url
func IsURL(str string) bool {
	if _, err := url.ParseRequestURI(str); err == nil {
		return true
	}
	return false
}
