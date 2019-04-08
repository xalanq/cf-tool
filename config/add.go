package config

import (
	"fmt"
	"os"
	"sort"
	"strings"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/xalanq/codeforces/client"
)

// Add template
func (c *Config) Add() (err error) {
	fmt.Println("Language list:")
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
	fmt.Print("Select a language(e.g. 42): ")
	var lang string
	fmt.Scanln(&lang)

	fmt.Print(`Template absolute path(e.g. ~/template/io.cpp): `)
	var path string
	for {
		fmt.Scanln(&path)
		path, err := homedir.Expand(path)
		if err == nil {
			if _, err := os.Stat(path); err == nil {
				break
			}
		}
		fmt.Printf("%v is invalid. Please input again: ", path)
	}

	fmt.Print("Match suffix(e.g. cpp cxx): ")
	var sf string
	fmt.Scanln(&sf)
	suffix := strings.Fields(sf)

	c.Template = append(c.Template, CodeTemplate{lang, path, suffix})
	return c.save()
}
