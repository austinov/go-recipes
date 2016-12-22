package config

import (
	"errors"
	"flag"
	"io/ioutil"
	"log"
	"sync"
	"time"

	"gopkg.in/yaml.v2"
)

const (
	defaultCfgPath = `./bot.yaml`
)

var (
	cfgPath string
	cfg     = Config{}
)

type (
	BotConfig struct {
		Token       string `yaml:"token"`
		NumHandlers int    `yaml:"num-handlers"`
		NumSenders  int    `yaml:"num-senders"`
	}

	DBConfig struct {
		Type             string `yaml:"type"`
		ConnectionString string `yaml:"connection-string"`
	}

	CMetalConfig struct {
		BaseURL    string        `yaml:"base-url"`
		Frequency  time.Duration `yaml:"frequency"`
		NumLoaders int           `yaml:"num-loaders"`
		NumSavers  int           `yaml:"num-savers"`
	}

	Config struct {
		Bot    BotConfig    `yaml:"bot"`
		DB     DBConfig     `yaml:"db"`
		CMetal CMetalConfig `yaml:"cmetal"`
	}
)

func init() {
	flag.StringVar(&cfgPath, "config", defaultCfgPath, "application's configuration file")
	flag.Parse()
}

var once sync.Once

func GetConfig() Config {
	once.Do(func() {
		log.Printf("Application started with config file '%s'\n", cfgPath)
		data, err := ioutil.ReadFile(cfgPath)
		if err != nil {
			log.Fatal(err)
		}
		err = yaml.Unmarshal([]byte(data), &cfg)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Parsed configuration: %#v\n", cfg)
	})
	return cfg
}

func (c Config) Verify() error {
	if err := c.Bot.Verify(); err != nil {
		return err
	}
	if err := c.DB.Verify(); err != nil {
		return err
	}
	if err := c.CMetal.Verify(); err != nil {
		return err
	}
	return nil
}

func (c BotConfig) Verify() error {
	if c.Token == "" {
		return errors.New("Bot token is empty")
	}
	return nil
}

func (c DBConfig) Verify() error {
	if c.Type == "" {
		return errors.New("Db type is empty")
	}
	return nil
}

func (c CMetalConfig) Verify() error {
	if c.BaseURL == "" {
		return errors.New("Concert-metal base url is empty")
	}
	return nil
}
