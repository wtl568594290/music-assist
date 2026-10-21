package main

import (
	_ "embed"
	"musicassist/tools"
)

//go:embed build/config.yml
var buildConfigYml []byte

//go:embed TaskFile.yml
var taskFileYml []byte

//go:embed README.md
var readme string

var staticConfig *tools.StaticConfig

func init() {
	staticConfig = tools.NewStaticConfig(buildConfigYml, taskFileYml)
}
