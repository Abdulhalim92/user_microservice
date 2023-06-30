package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"sync"
	"user/internal/logging"
)

type Config struct {
	BrokerCfg BrokerCfg `yaml:"broker"`
	DbCfg     DbCfg     `yaml:"db"`
	RedisCfg  RedisCfg  `yaml:"redis"`
}

type BrokerCfg struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}

type DbCfg struct {
	Driver   string `yaml:"driver"`
	Host     string `yaml:"host" env-default:"120.0.0.1"`
	Port     string `yaml:"port" env-default:"5432"`
	DbName   string `yaml:"db_name"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Sslmode  string `yaml:"sslmode"`
}

type RedisCfg struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}

var (
	instance *Config
	once     sync.Once
)

func GetConfig() *Config {
	once.Do(func() {
		logger := logging.GetLogger()
		logger.Info("read application configuration")
		instance = &Config{}
		if err := cleanenv.ReadConfig("./config/config.yml", instance); err != nil {
			help, _ := cleanenv.GetDescription(instance, nil)
			logger.Info(help)
			logger.Fatal(err)
		}
	})
	return instance
}
