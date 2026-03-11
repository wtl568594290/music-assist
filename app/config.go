package app

import (
	"errors"
	"log"
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
	SortType    SortType
}

const (
	CONFIG_NAME = "config.ini"
)

var Config *Conf
var iniFile *ini.File
var configPath = ""

func init() {
	userDir, err := os.UserConfigDir()
	if err != nil {
		panic(err)
	}
	appDir := filepath.Join(userDir, "music-assist")
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
	Config = &Conf{
		Height:      768,
		Width:       1024,
		PosX:        -1,
		PosY:        -1,
		LatestIndex: -1,
		SortType:    SortByName,
	}
	err = iniFile.MapTo(Config)
	if err != nil {
		panic(err)
	}
}

func (c *Conf) Save() error {
	if configPath == "" {
		return errors.New("config path is empty")
	}
	err := iniFile.ReflectFrom(c)
	if err != nil {
		return err
	}
	return iniFile.SaveTo(configPath)
}

func (a *app) GetConfigs() Conf {
	log.Printf("GetConfigs: %v", Config)
	viewAbleConfig := Conf{
		LatestIndex: Config.LatestIndex,
		SortType:    Config.SortType,
	}
	return viewAbleConfig
}
