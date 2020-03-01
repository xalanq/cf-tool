package client

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/fatih/color"
	"github.com/xalanq/cf-tool/util"
)

// Register for a contest
func (c *Client) Register(contestID string) error {
	URL := fmt.Sprintf("%v/contestRegistration/%v", c.host, contestID)
	resp, err := c.client.Get(URL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if _, err = findHandle(body); err != nil {
		return err
	}
	if msg := findCodeforcesMessage(body); msg != "" {
		return errors.New(msg)
	}
	doc, err := goquery.NewDocumentFromReader(ioutil.NopCloser(bytes.NewReader(body)))
	if err != nil {
		return err
	}
	color.HiCyan(findTitle(doc))
	if !agreesToTerms(doc) {
		return errors.New("You cannot participate without agreeing to the terms")
	}
	formData, err := getFormData(doc, c)
	if err != nil {
		return err
	}
	resp, err = c.client.PostForm(URL, formData)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	fmt.Println(findCodeforcesMessage(body))
	return nil
}

func findTitle(doc *goquery.Document) string {
	title := doc.Find("h2").Text()
	return cleanText(title)
}

func agreesToTerms(doc *goquery.Document) bool {
	label := cleanText(doc.Find("label[for=registrationTerms]").Text())
	terms := cleanText(doc.Find("#registrationTerms").Text())
	terms = strings.ReplaceAll(terms, "\n*", "\n    *")
	color.Green(label)
	fmt.Println(terms)
	return util.YesOrNo("Do you argree to the terms? (y/n)")
}

func getFormData(doc *goquery.Document, c *Client) (url.Values, error) {
	form := doc.Find(".contestRegistration").First()
	data := url.Values{}
	var err error
	form.Find("input").Each(func(i int, input *goquery.Selection) {
		key, k := input.Attr("name")
		value, v := input.Attr("value")
		if key != "" {
			if !k || !v {
				err = errors.New("Unable to get form data")
				return
			}
			data.Set(key, value)
		}
	})
	if data.Get("_tta") == "" {
		tta, err := getTta(c)
		if err != nil {
			return data, nil
		}
		data.Set("_tta", tta)
	}
	return data, err
}
