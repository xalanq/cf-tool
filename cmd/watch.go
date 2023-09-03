package cmd

import (
	"github.com/Sahaj-Bamba/cf-tool/client"
)

// Watch command
func Watch() (err error) {
	cln := client.Instance
	info := Args.Info
	n := 10
	if Args.All {
		n = -1
	}
	if _, err = cln.WatchSubmission(info, n, false); err != nil {
		if err = loginAgain(cln, err); err == nil {
			_, err = cln.WatchSubmission(info, n, false)
		}
	}
	return
}
