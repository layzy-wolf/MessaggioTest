package config

import (
	"flag"
	"github.com/ilyakaznacheev/cleanenv"
	"log"
	"os"
)

var (
	confPath string
	conf     Cfg
)

type Cfg struct {
	Receiver receiverCfg `yaml:"receiver"`
	Broker   brokerCfg   `yaml:"broker"`
	DB       dbCfg       `yaml:"db"`
}

type receiverCfg struct {
	Socket string `yaml:"socket" socket-default:":8080"`
}

type brokerCfg struct {
	Socket      string `yaml:"socket" socket-default:"localhost:9092"`
	CommonTopic string `yaml:"topic" topic-default:"messages"`
	CommonGroup string `yaml:"group" group-default:"messages-group"`
}

type dbCfg struct {
	Host     string `yaml:"host" socket-default:"localhost"`
	Port     int    `yaml:"port" port-default:"5432"`
	DB       string `yaml:"db" db-default:"default"`
	User     string `yaml:"user" user-default:"default"`
	Password string `yaml:"password" password-default:"default"`
}

func init() {
	flag.StringVar(&confPath, "config", "", "path to config")
	flag.Parse()

	if confPath == "" {
		confPath = os.Getenv("CHAT_CONFIG_PATH")
	}

	if confPath == "" {
		confPath = "./config/base.yaml"
	}
}

func Load() Cfg {
	if _, err := os.Stat(confPath); os.IsNotExist(err) {
		log.Panicf("config path is empty: %s", confPath)
	}

	if err := cleanenv.ReadConfig(confPath, &conf); err != nil {
		log.Panicf("unexpected err: %v", err)
	}

	return conf
}
