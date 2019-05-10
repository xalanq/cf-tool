package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/fatih/color"
	ansi "github.com/k0kubun/go-ansi"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/xalanq/cf-tool/client"
	"github.com/xalanq/cf-tool/util"
)

// AddTemplate add template
func (c *Config) AddTemplate() (err error) {
	color.Cyan("Language list:")
	type kv struct {
		K, V string
	}
	langs := []kv{}
	for k, v := range client.Langs {
		langs = append(langs, kv{k, v})
	}
	sort.Slice(langs, func(i, j int) bool { return langs[i].V < langs[j].V })
	for _, t := range langs {
		fmt.Printf("%5v: %v\n", t.K, t.V)
	}
	color.Cyan("Select a language (e.g. 42): ")
	lang := util.ScanlineTrim()

	color.Cyan("Alias (e.g. cpp): ")
	alias := util.ScanlineTrim()

	color.Cyan(`Template absolute path(e.g. ~/template/io.cpp): `)
	path := ""
	for {
		path = util.ScanlineTrim()
		path, err = homedir.Expand(path)
		if err == nil {
			if _, err := os.Stat(path); err == nil {
				break
			}
		}
		color.Red("%v is invalid. Please input again: ", path)
	}

	color.Cyan("(The suffix of template above will be added by default) Other suffix? (e.g. cxx cc): ")
	tmpSuffix := strings.Fields(util.ScanlineTrim())
	tmpSuffix = append(tmpSuffix, strings.Replace(filepath.Ext(path), ".", "", 1))
	suffixMap := map[string]bool{}
	suffix := []string{}
	for _, s := range tmpSuffix {
		if _, ok := suffixMap[s]; !ok {
			suffixMap[s] = true
			suffix = append(suffix, s)
		}
	}

	color.Cyan("Before script (e.g. g++ $%full%$ -o $%file%$.exe -std=c++11), empty is ok: ")
	beforeScript := util.ScanlineTrim()

	color.Cyan("Script (e.g. ./$%file%$.exe): ")
	script := ""
	for {
		script = util.ScanlineTrim()
		if len(script) > 0 {
			break
		}
		color.Red("script can not be empty. Please input again: ")
	}

	color.Cyan("After script (e.g. rm $%file%$.exe): ")
	afterScript := util.ScanlineTrim()

	c.Template = append(c.Template, CodeTemplate{
		alias, lang, path, suffix,
		beforeScript, script, afterScript,
	})

	if util.YesOrNo("Make it default (y/n)? ") {
		c.Default = len(c.Template) - 1
	}

	return c.save()
}

// RemoveTemplate remove template
func (c *Config) RemoveTemplate() (err error) {
	if len(c.Template) == 0 {
		color.Red("There is no template. Please add one")
		return nil
	}
	for i, template := range c.Template {
		star := " "
		if i == c.Default {
			star = color.New(color.FgGreen).Sprint("*")
		}
		ansi.Printf(`%v%2v: "%v" "%v"`, star, i, template.Alias, template.Path)
		ansi.Println()
	}
	idx := util.ChooseIndex(len(c.Template))
	c.Template = append(c.Template[:idx], c.Template[idx+1:]...)
	if idx == c.Default {
		c.Default = 0
	} else if idx < c.Default {
		c.Default--
	}
	return c.save()
}

// SetDefaultTemplate set default template index
func (c *Config) SetDefaultTemplate() error {
	if len(c.Template) == 0 {
		color.Red("There is no template. Please add one")
		return nil
	}
	for i, template := range c.Template {
		star := " "
		if i == c.Default {
			star = color.New(color.FgGreen).Sprint("*")
		}
		ansi.Printf(`%v%2v: "%v" "%v"`, star, i, template.Alias, template.Path)
		ansi.Println()
	}
	c.Default = util.ChooseIndex(len(c.Template))
	return c.save()
}

// TemplateByAlias return all template which alias equals to alias
func (c *Config) TemplateByAlias(alias string) []CodeTemplate {
	ret := []CodeTemplate{}
	for _, template := range c.Template {
		if template.Alias == alias {
			ret = append(ret, template)
		}
	}
	return ret
}
