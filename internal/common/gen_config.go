package common

type GenerateConfig struct {
	BuildConfig buildConfig
	TaskFile    taskFile
	Env         env
}

type env struct {
	GiteeAccessToken  string `env:"GITEE_ACCESS_TOKEN"`
	GithubAccessToken string `env:"GITHUB_ACCESS_TOKEN"`
	R2PublicBaseUrl   string `yaml:"R2_PUBLIC_BASE_URL"`
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

func DefaultGenerateConfig() *GenerateConfig {
	gc := GenerateConfig{
		BuildConfig: buildConfig{
			Info: info{
				CompanyName:       "wtl1989",
				ProductName:       "Music assist",
				ProductIdentifier: "com.wtl1989.musicassist",
				Description:       "Music Assist",
				Copyright:         "Copyright © 2026 wtl1989",
				Comments:          "Music Assist",
				Version:           "1.0.0",
			},
		},
		TaskFile: taskFile{
			Vars: vars{
				AppName: "music-assist",
			},
		},
		Env: env{
			R2PublicBaseUrl:   "",
			GiteeAccessToken:  "",
			GithubAccessToken: "",
		},
	}
	return &gc
}

func (s *GenerateConfig) ExeName() string {
	return s.TaskFile.Vars.AppName
}

func (s *GenerateConfig) ProductName() string {
	return s.BuildConfig.Info.ProductName
}

func (s *GenerateConfig) Version() string {
	return s.BuildConfig.Info.Version
}
