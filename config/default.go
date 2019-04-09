package config

import (
	"fmt"

	"github.com/xalanq/cf-tool/util"
)

// SetDefault set default template index
func (c *Config) SetDefault() error {
	for i, template := range c.Template {
		star := " "
		if i == c.Default {
			star = "*"
		}
		fmt.Printf("%v%2v: (%v) %v\n", star, i, template.Alias, template.Path)
	}
	c.Default = util.ChooseIndex(len(c.Template))
	return c.save()
}
