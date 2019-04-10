package config

import homedir "github.com/mitchellh/go-homedir"

// ConfigPath path to config data
var ConfigPath = "~/.cfconfig"

// SessionPath path to config data
var SessionPath = "~/.cfsession"

// Init unwrap homedir of config path and session path
func Init() {
	ConfigPath, _ = homedir.Expand(ConfigPath)
	SessionPath, _ = homedir.Expand(SessionPath)
}
