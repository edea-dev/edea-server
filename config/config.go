package config

import (
	"os"

	"github.com/kelseyhightower/envconfig"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

var Cfg Config

// Config holds all the keys which should be available in the conig.yml or as
// environment variables
type Config struct {
	Server struct {
		Port string `yaml:"port" envconfig:"SERVER_PORT"`
		Host string `yaml:"host" envconfig:"SERVER_HOST"`
	} `yaml:"server"`
	Database struct {
		Username string `yaml:"user" envconfig:"DB_USERNAME"`
		Password string `yaml:"pass" envconfig:"DB_PASSWORD"`
	} `yaml:"database"`
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
		Auth0 struct {
			Use           bool   `yaml:"use" envconfig:"USE_AUTH0"`
			ProviderURL   string `yaml:"provider_url" envconfig:"AUTH0_PROVIDER_URL"`
			ClientID      string `yaml:"client_id" envconfig:"AUTH0_CLIENT_ID"`
			ClientSecret  string `yaml:"client_secret" envconfig:"AUTH0_CLIENT_SECRET"`
			RedirectURL   string `yaml:"redirect_url" envconfig:"AUTH0_REDIRECT_URL"`
			SessionSecret string `yaml:"session_secret" envconfig:"AUTH0_SESSION_SECRET"`
		} `yaml:"auth0"`
		JWKS string `yaml:"jwks" encvonfig:"AUTH_JWKS"`
	} `yaml:"auth"`
}

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
