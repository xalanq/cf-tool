package cmd

import (
	"os"

	"github.com/xalanq/cf-tool/client"
)

// Clone command
func Clone() (err error) {
	currentPath, err := os.Getwd()
	if err != nil {
		return
	}
	cln := client.Instance
	ac := Args.Accepted
	handle := Args.Handle

	if err = cln.Clone(handle, currentPath, ac); err != nil {
		if err = loginAgain(cln, err); err == nil {
			err = cln.Clone(handle, currentPath, ac)
		}
	}
	return
}
