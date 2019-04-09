package config

// Alias return all template which alias equals to alias
func (c *Config) Alias(alias string) []CodeTemplate {
	ret := []CodeTemplate{}
	for _, template := range c.Template {
		if template.Alias == alias {
			ret = append(ret, template)
		}
	}
	return ret
}
