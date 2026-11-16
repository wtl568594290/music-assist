package main

import (
	"encoding/json"
	"musicassist/internal/common"
	"os"

	"github.com/joho/godotenv"
	"go.yaml.in/yaml/v4"
)

const (
	// 此路径是相对于执行go generate的目录
	rootPath = "../../"

	envPath            = rootPath + ".env"
	buildConfigYmlPath = rootPath + "build/config.yml"
	taskFileYmlPath    = rootPath + "TaskFile.yml"
	outPath            = rootPath + "internal/config/gen_config.json"
)

func main() {
	err := Generate()
	if err != nil {
		panic(err)
	}
}

func Generate() error {
	gc := common.DefaultGenerateConfig()
	if err := godotenv.Load(envPath); err == nil {
		r2PublicBaseUrl := os.Getenv("R2_PUBLIC_BASE_URL")
		if r2PublicBaseUrl != "" {
			gc.Env.R2PublicBaseUrl = r2PublicBaseUrl
		}
		giteeAccessToken := os.Getenv("GITEE_ACCESS_TOKEN")
		if giteeAccessToken != "" {
			gc.Env.GiteeAccessToken = giteeAccessToken
		}
		githubAccessToken := os.Getenv("GITHUB_ACCESS_TOKEN")
		if githubAccessToken != "" {
			gc.Env.GithubAccessToken = githubAccessToken
		}
	} else {
		return err
	}

	buildConfigYml, err := os.ReadFile(buildConfigYmlPath)
	if err != nil {
		return err
	}
	if err = yaml.Load(buildConfigYml, &gc.BuildConfig); err != nil {
		return err
	}
	taskFileYml, err := os.ReadFile(taskFileYmlPath)
	if err != nil {
		return err
	}
	if err = yaml.Load(taskFileYml, &gc.TaskFile); err != nil {
		return err
	}
	jsonStr, err := json.Marshal(gc)
	if err != nil {
		return err
	}

	// 生成json文件
	return os.WriteFile(outPath, jsonStr, 0644)
}
