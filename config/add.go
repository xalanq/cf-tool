package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/fatih/color"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/xalanq/cf-tool/client"
	"github.com/xalanq/cf-tool/util"
)

// Add template
func (c *Config) Add() (err error) {
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

	color.Cyan("Input alias (e.g. cpp): ")
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

	color.Cyan("Other suffix? (e.g. cxx cc): ")
	suffix := strings.Fields(util.ScanlineTrim())
	suffix = append(suffix, strings.Replace(filepath.Ext(path), ".", "", 1))

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

	color.Cyan("Make it default (y/n)? ")
	for {
		tmp := util.ScanlineTrim()
		if tmp == "y" || tmp == "Y" {
			c.Default = len(c.Template) - 1
			break
		}
		if tmp == "n" || tmp == "N" {
			break
		}
		color.Red("Invalid input. Please input again: ")
	}
	return c.save()
}
