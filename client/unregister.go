package client

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// Unregister from a contest
func (c *Client) Unregister(contestID string) error {
	resp, err := getRegistrantsPage(c, contestID, 1)
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
	formData, err := getUnregisterFormData(doc, c, contestID)
	if err != nil {
		return err
	}
	URL := fmt.Sprintf("%v/data/contestRegistration/%v", c.host, contestID)
	resp, err = c.client.PostForm(URL, formData)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if strings.Contains(string(body), `{"success":"true"}`) {
		fmt.Println("Succesfully unregistered from the contest")
	} else {
		return errors.New("Can't unregister. Possible reason: you made at least one action in the contest")
	}
	return err
}

func getUnregisterFormData(doc *goquery.Document, c *Client, contestID string) (url.Values, error) {
	pageCount := getPageCount(doc)
	for page := 1; page <= pageCount; page++ {
		if page != 1 {
			resp, err := getRegistrantsPage(c, contestID, page)
			if err != nil {
				return nil, err
			}
			defer resp.Body.Close()
			doc, err = goquery.NewDocumentFromReader(resp.Body)
			if err != nil {
				return nil, err
			}
		}
		user := doc.Find(".deleteParty").First()
		participantID, ok := user.Attr("participantid")
		if !ok {
			continue
		}
		data := url.Values{}
		token := getCsrfToken(doc)
		data.Add("participantId", participantID)
		data.Add("action", "deleteParty")
		data.Add("csrf_token", token)
		return data, nil
	}
	return nil, errors.New("You are not registered in this contest")
}

func getRegistrantsPage(c *Client, contestID string, page int) (r *http.Response, err error) {
	URL := fmt.Sprintf("%v/contestRegistrants/%v/friends/true/page/%d", c.host, contestID, page)
	resp, err := c.client.Get(URL)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func getPageCount(doc *goquery.Document) int {
	count, ok := doc.Find(".page-index").Last().Attr("pageindex")
	if ok {
		c, _ := strconv.Atoi(count)
		return c
	}
	return 1
}
