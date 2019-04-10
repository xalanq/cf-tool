package cmd

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/xalanq/cf-tool/config"
	"github.com/xalanq/cf-tool/util"
)

// Gen command
func Gen(args map[string]interface{}) error {
	currentPath, err := os.Getwd()
	if err != nil {
		return err
	}
	savePath := filepath.Join(currentPath, filepath.Base(currentPath))
	path := ""
	cfg := config.New(config.ConfigPath)
	if len(cfg.Template) == 0 {
		return errors.New("You have to add at least one code template by `cf config add`")
	}
	if alias, ok := args["<alias>"].(string); ok {
		templates := cfg.Alias(alias)
		if len(templates) < 1 {
			return fmt.Errorf("Cannot find any template with alias %v", alias)
		} else if len(templates) == 1 {
			path = templates[0].Path
		} else {
			fmt.Printf("There are multiple templates with alias %v\n", alias)
			for i, template := range templates {
				fmt.Printf(`%3v: "%v"`, i, template.Path)
				fmt.Println()
			}
			i := util.ChooseIndex(len(templates))
			path = templates[i].Path
		}
	} else {
		if cfg.Default < 0 || cfg.Default >= len(cfg.Template) {
			return fmt.Errorf("Invalid default value %v in config file", cfg.Default)
		}
		path = cfg.Template[cfg.Default].Path
	}
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	now := time.Now()
	source := string(b)
	source = strings.ReplaceAll(source, "$%U%$", cfg.Username)
	source = strings.ReplaceAll(source, "$%Y%$", fmt.Sprintf("%v", now.Year()))
	source = strings.ReplaceAll(source, "$%M%$", fmt.Sprintf("%02v", int(now.Month())))
	source = strings.ReplaceAll(source, "$%D%$", fmt.Sprintf("%02v", now.Day()))
	source = strings.ReplaceAll(source, "$%h%$", fmt.Sprintf("%02v", now.Hour()))
	source = strings.ReplaceAll(source, "$%m%$", fmt.Sprintf("%02v", now.Minute()))
	source = strings.ReplaceAll(source, "$%s%$", fmt.Sprintf("%02v", now.Second()))
	ext := filepath.Ext(path)
	tmpPath := savePath + ext
	_, err = os.Stat(tmpPath)
	for i := 1; err == nil; i++ {
		nxtPath := fmt.Sprintf("%v%v%v", savePath, i, ext)
		fmt.Printf("%v is existed. Rename to %v\n", filepath.Base(tmpPath), filepath.Base(nxtPath))
		tmpPath = nxtPath
		_, err = os.Stat(tmpPath)
	}
	savePath = tmpPath
	err = ioutil.WriteFile(savePath, []byte(source), 0644)
	if err == nil {
		color.Green("Generated! See %v", filepath.Base(savePath))
	}
	return err
}
