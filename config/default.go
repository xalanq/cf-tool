package config

import (
	"github.com/fatih/color"
	ansi "github.com/k0kubun/go-ansi"

	"github.com/xalanq/cf-tool/util"
)

// SetDefault set default template index
func (c *Config) SetDefault() error {
	for i, template := range c.Template {
		star := " "
		if i == c.Default {
			star = color.New(color.FgGreen).Sprint("*")
		}
		ansi.Printf(`%v%2v: "%v" "%v" "%v"`, star, i, template.Alias, template.Path, template.Build)
		ansi.Println()
	}
	c.Default = util.ChooseIndex(len(c.Template))
	return c.save()
}
