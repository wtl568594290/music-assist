package tools

import (
	"os"

	"github.com/joho/godotenv"
)

type env struct {
	GiteeAccessToken  string `env:"GITEE_ACCESS_TOKEN"`
	GithubAccessToken string `env:"GITHUB_ACCESS_TOKEN"`
}

var Env = env{
	GiteeAccessToken:  "",
	GithubAccessToken: "",
}

func init() {
	if err := godotenv.Load(); err == nil {
		Env.GiteeAccessToken = os.Getenv("GITEE_ACCESS_TOKEN")
		Env.GithubAccessToken = os.Getenv("GITHUB_ACCESS_TOKEN")
	}
}
