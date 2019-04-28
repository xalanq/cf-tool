package cmd

import "github.com/xalanq/cf-tool/config"

// Config command
func Config(args map[string]interface{}) error {
	cfg := config.New(config.ConfigPath)
	if args["login"].(bool) {
		return cfg.Login(config.SessionPath)
	} else if args["add"].(bool) {
		return cfg.AddTemplate()
	} else if args["del"].(bool) {
		return cfg.RemoveTemplate()
	} else if args["default"].(bool) {
		return cfg.SetDefaultTemplate()
	}
	return nil
}
