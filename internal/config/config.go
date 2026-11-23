package config

import (
	"errors"
	"musicassist/internal/common"
	"os"
	"path/filepath"

	"gopkg.in/ini.v1"
)

// MetronomeConf 独立节拍器窗口的配置，对应 ini 中的 [metronome] 分区
type MetronomeConf struct {
	Width  int
	Height int
	PosX   int
	PosY   int
}

type Conf struct {
	Height   int
	Width    int
	PosX     int
	PosY     int
	Sources  []string
	SortType common.SortType

	Metronome MetronomeConf `ini:"metronome"`
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
	appDir := filepath.Join(userDir, ProjectName)
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
		Height:   768,
		Width:    1024,
		PosX:     -1,
		PosY:     -1,
		SortType: common.SortByName,

		Metronome: MetronomeConf{
			Width:  1024,
			Height: 768,
			PosX:   -1,
			PosY:   -1,
		},
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
	// ini.ReflectFrom 对空切片是空操作，不会覆盖文件里的旧值，
	// 因此列表为空时先显式把该 key 的值清空，其余字段仍由 ReflectFrom 写入
	if len(config.Sources) == 0 {
		iniFile.Section("").Key("Sources").SetValue("")
	}
	if err := iniFile.ReflectFrom(config); err != nil {
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

func GetMetronomeSize() (int, int) {
	return config.Metronome.Width, config.Metronome.Height
}

func SetMetronomeSize(width, height int) {
	config.Metronome.Width = width
	config.Metronome.Height = height
}

func GetMetronomePos() (int, int) {
	return config.Metronome.PosX, config.Metronome.PosY
}

func SetMetronomePos(posX, posY int) {
	config.Metronome.PosX = posX
	config.Metronome.PosY = posY
}

func GetSources() []string {
	return config.Sources
}

func SetSources(sources []string) {
	config.Sources = sources
}

func GetSortType() common.SortType {
	return config.SortType
}

func SetSortType(sortType common.SortType) {
	config.SortType = sortType
}
