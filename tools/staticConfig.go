package tools

import (
	"log"

	"go.yaml.in/yaml/v4"
)

type StaticConfig struct {
	BuildConfig buildConfig
	TaskFile    taskFile
}

type buildConfig struct {
	Info info
}

type info struct {
	CompanyName       string `yaml:"companyName"`
	ProductName       string `yaml:"productName"`
	ProductIdentifier string `yaml:"productIdentifier"`
	Description       string `yaml:"description"`
	Copyright         string `yaml:"copyright"`
	Comments          string `yaml:"comments"`
	Version           string `yaml:"version"`
}

type taskFile struct {
	Vars vars `yaml:"vars"`
}
type vars struct {
	AppName string `yaml:"APP_NAME"`
}

const (
	ProjectName = "music-assist"
)

func NewStaticConfig(buildConfigYml []byte, taskFileYml []byte) *StaticConfig {
	bc := buildConfig{
		Info: info{
			CompanyName:       "wtl1989",
			ProductName:       "Music assist",
			ProductIdentifier: "com.wtl1989.musicassist",
			Description:       "Music Assist",
			Copyright:         "Copyright © 2026 wtl1989",
			Comments:          "Music Assist",
			Version:           "1.0.0",
		},
	}
	if err := yaml.Load(buildConfigYml, &bc); err != nil {
		log.Printf("Error: %v", err)
	}
	tf := taskFile{
		Vars: vars{
			AppName: ProjectName,
		},
	}
	if err := yaml.Load(taskFileYml, &tf); err != nil {
		log.Printf("Error: %v", err)
	}
	return &StaticConfig{
		BuildConfig: bc,
		TaskFile:    tf,
	}
}

func (s *StaticConfig) ExeName() string {
	return s.TaskFile.Vars.AppName
}

func (s *StaticConfig) ProductName() string {
	return s.BuildConfig.Info.ProductName
}

func (s *StaticConfig) Version() string {
	return s.BuildConfig.Info.Version
}
