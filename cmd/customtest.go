package cmd

import (
	//"github.com/docopt/docopt-go"
	"io/ioutil"
	"strconv"
	"os"

	"github.com/xalanq/cf-tool/client"
)

// CustomTest command
func CustomTest() error {
	input := ""
	if Args.InputFile != "" {
		file, err := os.Open(Args.InputFile)
		if err != nil { return err }
		defer file.Close()

		bytes, err := ioutil.ReadAll(file)
		if err != nil { return err }

		input = string(bytes)
	}

	langId, err := strconv.Atoi(Args.LanguageID)
	if err != nil { return err }

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

	return nil
}
