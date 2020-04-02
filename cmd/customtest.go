package cmd

import (
	//"github.com/docopt/docopt-go"
	"io/ioutil"
	"strconv"
	"os"

	"github.com/xalanq/cf-tool/client"
	"github.com/xalanq/cf-tool/config"
)

// CustomTest command
func CustomTest() (err error) {
	input := ""
	if Args.InputFile != "" {
		file, err := os.Open(Args.InputFile)
		if err != nil { return err }
		defer file.Close()

		bytes, err := ioutil.ReadAll(file)
		if err != nil { return err }

		input = string(bytes)
	}

	langId := 0
	if Args.LanguageID == "" {
		cfg := config.Instance
		_, index, err := getOneCode(Args.File, cfg.Template)
		if err != nil { return err }
		langId, _ = strconv.Atoi(cfg.Template[index].Lang)
	} else {
		langId, err = strconv.Atoi(Args.LanguageID)
		if err != nil { return err }
	}

	file, err := os.Open(Args.File)
	if err != nil { return err }
	defer file.Close()

	bytes, err := ioutil.ReadAll(file)
	if err != nil { return err }

	source := string(bytes)

	cln := client.Instance
	if err = cln.CustomTest(langId, source, input); err != nil {
		if err = loginAgain(cln, err); err == nil {
			err = cln.CustomTest(langId, source, input)
		}
	}

	return
}
