package cmd

import (
	"os"

	"github.com/xalanq/cf-tool/client"
	"github.com/xalanq/cf-tool/config"
)

// Clone command
func Clone(args map[string]interface{}) error {
	currentPath, err := os.Getwd()
	if err != nil {
		return err
	}
	cfg := config.Instance
	cln := client.Instance
	ac := args["ac"].(bool)
	handle := args["<handle>"].(string)

	if err = cln.Clone(handle, currentPath, ac); err != nil {
		if err = loginAgain(cfg, cln, err); err == nil {
			err = cln.Clone(handle, currentPath, ac)
		}
	}
	return err
}
