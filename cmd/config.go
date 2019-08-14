package cmd

import (
	"github.com/fatih/color"
	ansi "github.com/k0kubun/go-ansi"
	"github.com/xalanq/cf-tool/config"
	"github.com/xalanq/cf-tool/util"
)

// Config command
func Config(args map[string]interface{}) error {
	color.Cyan("Configure the tool")
	cfg := config.New(config.ConfigPath)
	ansi.Println("0) username and password")
	ansi.Println("1) add a template")
	ansi.Println("2) delete a template")
	ansi.Println("3) set default template")
	index := util.ChooseIndex(4)
	if index == 0 {
		return cfg.Login(config.SessionPath)
	} else if index == 1 {
		return cfg.AddTemplate()
	} else if index == 2 {
		return cfg.RemoveTemplate()
	} else if index == 3 {
		return cfg.SetDefaultTemplate()
	}
	return nil
}
