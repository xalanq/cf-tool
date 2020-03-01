package client

import (
	"errors"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func cleanText(s string) string {
	s = strings.Trim(s, " \n")
	space := regexp.MustCompile(`\ +`)
	s = space.ReplaceAllString(s, " ")
	return s
}

func findCodeforcesMessage(body []byte) string {
	str := `\n\s{8}Codeforces\.showMessage\("(.+)"\);\s{8}Codeforces\.reformatTimes\(\)`
	reg := regexp.MustCompile(str)
	tmp := reg.FindStringSubmatch(string(body))
	if tmp != nil {
		return strings.ReplaceAll(tmp[1], "<br/>", "\n")
	}
	return ""
}

func getCsrfToken(doc *goquery.Document) string {
	token, _ := doc.Find("meta[name='X-Csrf-Token']").Attr("content")
	if len(token) == 32 {
		return token
	}
	token, _ = doc.Find("span.csrf-token").Attr("data-csrf")
	if len(token) == 32 {
		return token
	}
	return ""
}

func getTta(c *Client) (string, error) {
	cookie, err := c.Jar.GetEntry("codeforces.com", "/", "39ce7")
	if err != nil {
		return "", errors.New("Unable to get required cookie")
	}
	return decodeTta(cookie.Value), nil
}

func decodeTta(cookie string) string {
	var result int
	for i := 0; i < len(cookie); i++ {
		result = (result + (i+1)*(i+2)*int(cookie[i])) % 1009
		if i%3 == 0 {
			result++
		}
		if i%2 == 0 {
			result *= 2
		}
		if i > 0 {
			result -= int(cookie[i/2]) / 2 * (result % 5)
		}
		for result < 0 {
			result += 1009
		}
		for result >= 1009 {
			result -= 1009
		}
	}
	return strconv.Itoa(result)
}
