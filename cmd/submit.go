package cmd

import (
	"io/ioutil"

	"github.com/xalanq/cf-tool/client"
	"github.com/xalanq/cf-tool/config"
)

// Submit command
func Submit() (err error) {
	cln := client.Instance
	cfg := config.Instance
	info := Args.Info
	filename, index, err := getOneCode(Args.File, cfg.Template)
	if err != nil {
		return
	}

	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return
	}
	source := string(bytes)

	lang := cfg.Template[index].Lang
	if err = cln.Submit(info, lang, source); err != nil {
		if err = loginAgain(cln, err); err == nil {
			err = cln.Submit(info, lang, source)
		}
	}
	return
}
