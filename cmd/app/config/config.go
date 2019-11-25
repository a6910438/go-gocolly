package config

import (
	"fmt"
	"github.com/jinzhu/configor"
)

type (
	LogConfig struct {
		//日志
		Path  string `yaml:"path" env:"LOG_PATH"`
		File  string `yaml:"file" env:"LOG_FILE"`
		Level int    `yaml:"level" env:"LOG_LEVEL"`
	}

	DbConfig struct {
		//数据库
		Host     string `yaml:"host" env:"DB_HOST"`
		Port     int    `yaml:"port" env:"DB_PORT"`
		User     string `yaml:"user" env:"DB_USER"`
		Password string `yaml:"password" env:"DB_PASSWORD"`
		Dbname   string `yaml:"dbname" env:"DB_DBNAME"`
	}

	Configuration struct {
		Log          LogConfig `yaml:"log"`
		Db           DbConfig  `yaml:"db"`
		TokenviewUrl string    `yaml:"tokenview_url" env:"TOKENVIEW_URL"`
		BiWorldUrl   string    `yaml:"bi_shijie_url" env:"BISHIJIE_URL"`
		BBtcUrl      string    `yaml:"b_btc_url" env:"8BTC_URL"`
	}
)

var Cfg Configuration

func Init(filePath string) error {
	fmt.Println(filePath)
	err := configor.Load(&Cfg, filePath)
	fmt.Printf("config: %#v\n", Cfg)
	return err
}
