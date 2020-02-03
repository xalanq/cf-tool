package config

import (
	"cf-tool/util"
)

// SetGenAfterParse set it yes or no
func (c *Config) SetGenAfterParse() (err error) {
	c.GenAfterParse = util.YesOrNo(`Run "cf gen" after "cf parse" (y/n)? `)
	return c.save()
}
