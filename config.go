package main
import (
	"os"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Kratos struct {
		PublicURL string `yaml:"public_url"`
		AdminURL  string `yaml:"admin_url"`
		UIURL     string `yaml:"ui_url"`
	} `yaml:"kratos"`

	App struct {
		URL string `yaml:"url"`
	} `yaml:"app"`
}

var AppConfig Config

func (c *Config) KratosLoginURL() string {
	return c.Kratos.UIURL + "/login"
}

func loadConfig() {
	b, err := os.ReadFile("config.yaml")
	if err != nil {
		panic(err)
	}
	if err := yaml.Unmarshal(b, &AppConfig); err != nil {
		panic(err)
	}
}