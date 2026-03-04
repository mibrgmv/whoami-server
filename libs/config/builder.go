package config

import "strings"

type Builder struct {
	opts *Options
}

func NewBuilder() *Builder {
	return &Builder{
		opts: DefaultOptions(),
	}
}

func (b *Builder) WithConfigName(name string) *Builder {
	b.opts.ConfigName = name
	return b
}

func (b *Builder) WithConfigType(configType string) *Builder {
	b.opts.ConfigType = configType
	return b
}

func (b *Builder) WithConfigPaths(paths ...string) *Builder {
	b.opts.ConfigPaths = paths
	return b
}

func (b *Builder) WithEnvFiles(files ...string) *Builder {
	b.opts.EnvFiles = files
	return b
}

func (b *Builder) WithEnvPrefix(prefix string) *Builder {
	b.opts.EnvPrefix = prefix
	return b
}

func (b *Builder) WithEnvKeyReplacer(replacer *strings.Replacer) *Builder {
	b.opts.EnvKeyReplacer = replacer
	return b
}

func (b *Builder) IgnoreConfigNotFound() *Builder {
	b.opts.IgnoreNotFound = true
	return b
}

func (b *Builder) Silent() *Builder {
	b.opts.Silent = true
	return b
}

func (b *Builder) Load(cfg interface{}) error {
	return LoadWithOptions(cfg, b.opts)
}
