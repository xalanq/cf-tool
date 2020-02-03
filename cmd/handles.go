package cmd

import (
	"cf-tool/client"
	"os"
)

// Handles command
func Handles(args map[string]interface{}) error {
	currentPath, err := os.Getwd()
	if err != nil {
		return err
	}
	cln := client.New(currentPath)

	return cln.SaveHandles(currentPath)
}
