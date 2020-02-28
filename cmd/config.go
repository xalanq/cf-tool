package cmd

import (
	"github.com/fatih/color"
	ansi "github.com/k0kubun/go-ansi"
	"github.com/xalanq/cf-tool/client"
	"github.com/xalanq/cf-tool/config"
	"github.com/xalanq/cf-tool/util"
)

// Config command
func Config() error {
	cfg := config.Instance
	cln := client.Instance
	color.Cyan("Configure the tool")
	ansi.Println(`0) login`)
	ansi.Println(`1) add a template`)
	ansi.Println(`2) delete a template`)
	ansi.Println(`3) set default template`)
	ansi.Println(`4) run "cf gen" after "cf parse"`)
	ansi.Println(`5) set host domain`)
	ansi.Println(`6) set proxy`)
	index := util.ChooseIndex(7)
	if index == 0 {
		return cln.ConfigLogin()
	} else if index == 1 {
		return cfg.AddTemplate()
	} else if index == 2 {
		return cfg.RemoveTemplate()
	} else if index == 3 {
		return cfg.SetDefaultTemplate()
	} else if index == 4 {
		return cfg.SetGenAfterParse()
	} else if index == 5 {
		return cfg.SetHost()
	} else if index == 6 {
		return cfg.SetProxy()
	}
	return nil
}
