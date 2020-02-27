package cmd

import (
	"github.com/xalanq/cf-tool/client"
)

// Register command
func Register() error {
	return client.Instance.Register(Args.Info.ContestID)
}
