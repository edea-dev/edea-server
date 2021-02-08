package config

import (
	"os"

	"github.com/kelseyhightower/envconfig"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

// Cfg global config state
var Cfg Config

// Config holds all the keys which should be available in the conig.yml or as
// environment variables
type Config struct {
	Server struct {
		Port string `yaml:"port" envconfig:"SERVER_PORT"`
		Host string `yaml:"host" envconfig:"SERVER_HOST"`
	} `yaml:"server"`
	Dev bool   `yaml:"dev" envconfig:"IS_DEV"`
	DSN string `yaml:"dsn" envconfig:"DB_DSN"`
	JWT struct {
		PublicKey string `yaml:"publickey" envconfig:"JWT_PUBLIC_KEY"`
	} `yaml:"jwt"`
	API struct {
		GitHubToken string `yaml:"githubtoken" envconfig:"GITHUB_API_TOKEN"`
	} `yaml:"api"`
	Cache struct {
		Repo struct {
			Base string `yaml:"base" envconfig:"REPO_CACHE_BASE"`
		} `yaml:"repo"`
	} `yaml:"cache"`
	Auth struct {
		Kratos struct {
			Use  bool   `yaml:"use" envconfig:"USE_KRATOS"`
			Host string `yaml:"host" envconfig:"KRATOS_HOST"`
		} `yaml:"kratos"`
		OIDC struct {
			ProviderURL   string `yaml:"provider_url" envconfig:"AUTH_PROVIDER_URL"`
			ClientID      string `yaml:"client_id" envconfig:"AUTH_CLIENT_ID"`
			ClientSecret  string `yaml:"client_secret" envconfig:"AUTH_CLIENT_SECRET"`
			RedirectURL   string `yaml:"redirect_url" envconfig:"AUTH_REDIRECT_URL"`
			SessionSecret string `yaml:"session_secret" envconfig:"AUTH_SESSION_SECRET"`
		} `yaml:"oidc"`
		UseMock bool `yaml:"use_mock" envconfig:"USE_AUTH_MOCK"`
	} `yaml:"auth"`
}

// ReadConfig reads the configuration yaml file and overrides it with any set environment variables
func ReadConfig() {
	readFile(&Cfg)
	readEnv(&Cfg)
	log.Printf("%+v", Cfg)
}

func processError(err error) {
	log.Printf("%v", err)
	os.Exit(2)
}

func readFile(cfg *Config) {
	f, err := os.Open("config.yml")
	if err != nil {
		processError(err)
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(cfg)
	if err != nil {
		processError(err)
	}
}

func readEnv(cfg *Config) {
	err := envconfig.Process("", cfg)
	if err != nil {
		processError(err)
	}
}
