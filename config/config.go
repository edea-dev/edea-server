package config

// SPDX-License-Identifier: EUPL-1.2

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
	Tools struct {
		Merge   string `yaml:"merge" envconfig:"MERGE_TOOL"`
		PlotPCB string `yaml:"plotpcb" envconfig:"PLOTPCB"`
	} `yaml:"tools"`
	Cache struct {
		Repo struct {
			Base string `yaml:"base" envconfig:"REPO_CACHE_BASE"`
		} `yaml:"repo"`
		Book struct {
			Base string `yaml:"base" envconfig:"BOOK_CACHE_BASE"` // mdbook destination folder
		} `yaml:"book"`
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
			LogoutURL     string `yaml:"logout_url" envconfig:"AUTH_LOGOUT_URL"`
			PostLogoutURL string `yaml:"post_logout_url" envconfig:"AUTH_POST_LOGOUT_URL"`
		} `yaml:"oidc"`
		UseMock bool `yaml:"use_mock" envconfig:"USE_AUTH_MOCK"`
	} `yaml:"auth"`
	Search struct {
		Host   string `yaml:"host" envconfig:"SEARCH_HOST"`
		Index  string `yaml:"index" envconfig:"SEARCH_INDEX"`
		APIKey string `yaml:"api_key" envconfig:"SEARCH_API_KEY"`
	} `yaml:"search"`
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
		f, err = os.Open("/etc/edead.yml")
		if err != nil {
			processError(err)
		}
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
