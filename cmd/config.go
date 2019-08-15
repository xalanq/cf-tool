package cmd

import (
	"github.com/fatih/color"
	ansi "github.com/k0kubun/go-ansi"
	"github.com/xalanq/cf-tool/client"
	"github.com/xalanq/cf-tool/config"
	"github.com/xalanq/cf-tool/util"
)

// Config command
func Config(args map[string]interface{}) error {
	color.Cyan("Configure the tool")
	ansi.Println(`0) username and password`)
	ansi.Println(`1) add a template`)
	ansi.Println(`2) delete a template`)
	ansi.Println(`3) set default template`)
	ansi.Println(`4) run "cf gen" after "cf parse"`)
	ansi.Println(`5) set host domain`)
	index := util.ChooseIndex(6)
	if index == 0 {
		return config.New(config.ConfigPath).Login(config.SessionPath)
	} else if index == 1 {
		return config.New(config.ConfigPath).AddTemplate()
	} else if index == 2 {
		return config.New(config.ConfigPath).RemoveTemplate()
	} else if index == 3 {
		return config.New(config.ConfigPath).SetDefaultTemplate()
	} else if index == 4 {
		return config.New(config.ConfigPath).SetGenAfterParse()
	} else if index == 5 {
		return client.New(config.SessionPath).SetHost()
	}
	return nil
}
