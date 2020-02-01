package cmd

import (
	"cf-tool/client"
	"os"
)

// Status command
func Status(args map[string]interface{}) error {
	currentPath, err := os.Getwd()
	if err != nil {
		return err
	}
	cln := client.New(currentPath)
	username := args["<username>"].(string)

	return cln.SaveStatus(username, currentPath)
}
