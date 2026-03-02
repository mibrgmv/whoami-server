package config

import "strings"

type Options struct {
	ConfigName     string
	ConfigType     string
	ConfigPaths    []string
	EnvFiles       []string
	EnvPrefix      string
	EnvKeyReplacer *strings.Replacer
	IgnoreNotFound bool
	Silent         bool
}

func DefaultOptions() *Options {
	return &Options{
		EnvKeyReplacer: strings.NewReplacer(".", "_"),
		IgnoreNotFound: false,
		Silent:         false,
	}
}
