package main
import (
	"os"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Kratos struct {
		PublicURL string `yaml:"public_url"`
		LoginURL  string `yaml:"login_url"`
	} `yaml:"kratos"`

	App struct {
		AfterLoginRedirect string `yaml:"after_login_redirect"`
	} `yaml:"app"`
}

var AppConfig Config

func loadConfig() {
	b, err := os.ReadFile("conf.yaml")
	if err != nil {
		panic(err)
	}
	if err := yaml.Unmarshal(b, &AppConfig); err != nil {
		panic(err)
	}
}