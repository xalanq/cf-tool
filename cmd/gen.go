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

func parseTemplate(source string, cfg *config.Config) string {
	now := time.Now()
	source = strings.ReplaceAll(source, "$%U%$", cfg.Username)
	source = strings.ReplaceAll(source, "$%Y%$", fmt.Sprintf("%v", now.Year()))
	source = strings.ReplaceAll(source, "$%M%$", fmt.Sprintf("%02v", int(now.Month())))
	source = strings.ReplaceAll(source, "$%D%$", fmt.Sprintf("%02v", now.Day()))
	source = strings.ReplaceAll(source, "$%h%$", fmt.Sprintf("%02v", now.Hour()))
	source = strings.ReplaceAll(source, "$%m%$", fmt.Sprintf("%02v", now.Minute()))
	source = strings.ReplaceAll(source, "$%s%$", fmt.Sprintf("%02v", now.Second()))
	return source
}

func readTemplateSource(path string, cfg *config.Config) (source string, err error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return
	}
	source = parseTemplate(string(b), cfg)
	return
}

func gen(source, currentPath, ext string) error {
	path := filepath.Join(currentPath, filepath.Base(currentPath))

	savePath := path + ext
	i := 1
	for _, err := os.Stat(savePath); err == nil; _, err = os.Stat(savePath) {
		tmpPath := fmt.Sprintf("%v%v%v", path, i, ext)
		fmt.Printf("%v exists. Rename to %v\n", filepath.Base(savePath), filepath.Base(tmpPath))
		savePath = tmpPath
		i++
	}

	err := ioutil.WriteFile(savePath, []byte(source), 0644)
	if err == nil {
		color.Green("Generated! See %v", filepath.Base(savePath))
	}
	return err
}

// Gen command
func Gen(args map[string]interface{}) error {
	cfg := config.New(config.ConfigPath)
	if len(cfg.Template) == 0 {
		return errors.New("You have to add at least one code template by `cf config`")
	}

	var path string

	if alias, ok := args["<alias>"].(string); ok {
		templates := cfg.TemplateByAlias(alias)
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
		path = cfg.Template[cfg.Default].Path
	}

	source, err := readTemplateSource(path, cfg)
	if err != nil {
		return err
	}

	currentPath, err := os.Getwd()
	if err != nil {
		return err
	}

	ext := filepath.Ext(path)
	gen(source, currentPath, ext)
	return err
}
