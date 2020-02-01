package client

import (
	"cf-tool/client/api"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
)

func (c *Client) Status(username string) ([]api.Submission, error) {
	resp, err := c.client.Get(fmt.Sprintf(c.Host+"/api/user.status?handle=%v", username))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	var data map[string]interface{}
	if err = decoder.Decode(&data); err != nil {
		return nil, err
	}

	if status, ok := data["status"].(string); !ok || status != "OK" {
		return nil, fmt.Errorf("Cannot get any submission")
	}

	submissions := data["result"].([]interface{})
	var status []api.Submission

	for _, _submission := range submissions {
		submission := _submission.(map[string]interface{})

		verdict := submission["verdict"].(string)
		contestID := submission["contestId"].(float64)
		submissionID := submission["id"].(float64)
		lang := submission["programmingLanguage"].(string)
		timestamp := submission["creationTimeSeconds"].(float64)
		problemID := strings.ToLower(submission["problem"].(map[string]interface{})["index"].(string))

		status = append(status, *api.NewSubmission(contestID, submissionID, problemID, verdict, lang, timestamp))
	}

	return status, nil
}

func (c *Client) SaveStatus(username, path string) error {
	status, err := c.Status(username)
	if err != nil {
		return err
	}
	finalPath := filepath.Join(path, "status", fmt.Sprintf("%s.json", username))
	b, err := json.Marshal(status)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(finalPath, b, 0644)
}
