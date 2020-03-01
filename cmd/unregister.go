package cmd

import (
	"github.com/xalanq/cf-tool/client"
)

// Unregister command
func Unregister() error {
	return client.Instance.Unregister(Args.Info.ContestID)
}
