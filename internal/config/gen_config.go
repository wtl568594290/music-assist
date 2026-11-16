package config

import (
	_ "embed"
	"encoding/json"
	"musicassist/internal/common"
)

//go:generate go run ../../cmd/genconfig/main.go

//go:embed gen_config.json
var genConfigJson []byte

// GenConfig is the generated config.
var GenConfig *common.GenerateConfig = &common.GenerateConfig{}

const (
	ProjectName = "music-assist"
)

func init() {
	if err := json.Unmarshal(genConfigJson, GenConfig); err != nil {
		panic(err)
	}
}
