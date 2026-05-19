package main
import (
	"os"
	"log"
)

type Config struct {
	Kratos struct {
		APIURL     string // サーバー間通信用 (例: http://kratos:4433)
		BrowserURL string // ブラウザリダイレクト用 (例: http://localhost:4433)
		UIURL      string
	}

	App struct {
		URL string
	}
}

var AppConfig Config

func (c *Config) KratosLoginURL() string {
	return c.Kratos.UIURL + "/login"
}

func mustGetenv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("environment variable %s is required", key)
	}
	return v
}

func loadConfig() {
	AppConfig = Config{}

	AppConfig.Kratos.APIURL     = mustGetenv("KRATOS_API_URL")
	AppConfig.Kratos.BrowserURL = mustGetenv("KRATOS_BROWSER_URL")
	AppConfig.Kratos.UIURL      = mustGetenv("KRATOS_UI_URL")
	AppConfig.App.URL           = mustGetenv("APP_URL")
}