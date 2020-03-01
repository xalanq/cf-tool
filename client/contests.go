package client

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// ContestInfo contests information
type ContestInfo struct {
	ID           string
	Name         string
	Start        string
	Length       string
	State        string
	Registration string
	Registered   bool
}

func getRegistrationStatus(cell *goquery.Selection) (string, bool) {
	participants := cell.Find(".contestParticipantCountLinkMargin")
	participantCount := participants.Text()
	participants.Remove()
	if cell.Find(".welldone").Length() != 0 {
		text := cleanText(cell.Text())
		return text + "\n" + participantCount, true
	}
	countdown := cell.Find(".countdown")
	text := ""
	if strings.Contains(cell.Text(), "»") {
		text = "Registering"
		parent := countdown.Parent()
		countdownText := countdown.Text()
		countdown.Remove()
		note := cleanText(parent.Text())
		if strings.Contains(cell.Text(), "*") {
			countdownText += "*"
		}
		return fmt.Sprintf("%s %s\n%s %s", text, participantCount, note, countdownText), false
	}
	return cleanText(cell.Text()), false
}

func getState(cell *goquery.Selection) string {
	cell.Find("a").Remove()
	countdown := cell.Find(".countdown").Text()
	cell.Find(".countdown").Remove()
	text := cleanText(cell.Text())

	return text + "\n" + countdown
}

func findContests(body io.ReadCloser, utcOffset string) ([]ContestInfo, error) {
	doc, err := goquery.NewDocumentFromReader(body)
	if err != nil {
		return nil, err
	}
	table := doc.Find(".datatable").First().Find("tbody")
	rows := table.Find("tr").Slice(1, goquery.ToEnd)
	contests := []ContestInfo{}
	rows.Each(func(i int, row *goquery.Selection) {
		contest := ContestInfo{}
		contest.ID, _ = row.Attr("data-contestid")
		row.Find("td").Each(func(j int, cell *goquery.Selection) {
			switch j {
			case 0:
				cell.Find("a").Remove()
				name := cleanText(cell.Text())
				contest.Name = strings.Replace(name, " (", "\n(", 1)
			case 2:
				contest.Start = parseWhen(cell.Find(".format-time").Text(), utcOffset)
			case 3:
				duration := cleanText(cell.Text())
				if strings.Count(duration, ":") == 2 {
					duration = strings.TrimSuffix(duration, ":00")
				}
				contest.Length = duration
			case 4:
				contest.State = getState(cell)
			case 5:
				contest.Registration, contest.Registered = getRegistrationStatus(cell)
			}
		})
		contests = append(contests, contest)
	})
	if err != nil {
		return nil, err
	}
	return contests, nil
}

// StatisContest get upcoming contests
func (c *Client) GetContests() (contests []ContestInfo, err error) {
	URL := c.host + "/contests?complete=true"
	resp, err := c.client.Get(URL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Status code error: %d %s", resp.StatusCode, resp.Status)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if _, err = findHandle(body); err != nil {
		return
	}
	utcOffset, err := findCfOffset(body)
	if err != nil {
		return nil, err
	}
	return findContests(ioutil.NopCloser(bytes.NewReader(body)), utcOffset)
}
