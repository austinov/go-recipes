package config

import (
	"flag"
	"io/ioutil"
	"log"
	"sync"

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
		Token string `yaml:"token"`
	}

	DBConfig struct {
		Network  string `yaml:"network"`
		Address  string `yaml:"address"`
		Password string `yaml:"password"`
	}

	Config struct {
		Bot BotConfig `yaml:"bot"`
		Db  DBConfig  `yaml:"db"`
	}
)

func init() {
	flag.StringVar(&cfgPath, "config", defaultCfgPath, "application's configuration file")
}

var once sync.Once

func GetConfig() Config {
	once.Do(func() {
		log.Printf("Application started with config file '%s'", cfgPath)
		data, err := ioutil.ReadFile(cfgPath)
		if err != nil {
			log.Fatal(err)
		}
		err = yaml.Unmarshal([]byte(data), &cfg)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Parsed configuration: %#v", cfg)
	})
	return cfg
}
