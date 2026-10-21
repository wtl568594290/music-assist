package config

import (
	"errors"
	"musicassist/tools"
	"os"
	"path/filepath"

	"gopkg.in/ini.v1"
)

type Conf struct {
	Height      int
	Width       int
	PosX        int
	PosY        int
	Sources     []string
	LatestIndex float64
	SortType    tools.SortType
	RemoteUrl   string
}

const (
	CONFIG_NAME = "config.ini"
)

var config *Conf
var iniFile *ini.File
var configPath = ""

func init() {
	userDir, err := os.UserConfigDir()
	if err != nil {
		panic(err)
	}
	appDir := filepath.Join(userDir, tools.ProjectName)
	if _, err := os.Stat(appDir); os.IsNotExist(err) {
		os.Mkdir(appDir, 0755)
	}
	configPath = filepath.Join(appDir, CONFIG_NAME)
	// 如果文件不存在，创建一个空文件
	_, err = os.Stat(configPath)
	if os.IsNotExist(err) {
		os.Create(configPath)
	}
	iniFile, err = ini.Load(configPath)
	if err != nil {
		panic(err)
	}
	config = &Conf{
		Height:      768,
		Width:       1024,
		PosX:        -1,
		PosY:        -1,
		LatestIndex: -1,
		SortType:    tools.SortByName,
		RemoteUrl:   "https://render-test-zar6.onrender.com",
	}
	err = iniFile.MapTo(config)
	if err != nil {
		panic(err)
	}
}

func Save() error {
	if configPath == "" {
		return errors.New("config path is empty")
	}
	err := iniFile.ReflectFrom(config)
	if err != nil {
		return err
	}
	return iniFile.SaveTo(configPath)
}

func GetSize() (int, int) {
	return config.Width, config.Height
}

func SetSize(width, height int) {
	config.Width = width
	config.Height = height
}

func GetPos() (int, int) {
	return config.PosX, config.PosY
}

func SetPos(posX, posY int) {
	config.PosX = posX
	config.PosY = posY
}

func GetSources() []string {
	return config.Sources
}

func SetSources(sources []string) {
	config.Sources = sources
}

func GetLatestIndex() float64 {
	return config.LatestIndex
}

func SetLatestIndex(index float64) {
	config.LatestIndex = index
}

func GetSortType() tools.SortType {
	return config.SortType
}

func SetSortType(sortType tools.SortType) {
	config.SortType = sortType
}

func GetRemoteUrl() string {
	return config.RemoteUrl
}

func SetRemoteUrl(remoteUrl string) {
	config.RemoteUrl = remoteUrl
}

func ViewableConfig() Conf {
	viewAbleConfig := Conf{
		LatestIndex: config.LatestIndex,
		SortType:    config.SortType,
		RemoteUrl:   config.RemoteUrl,
	}
	return viewAbleConfig
}
