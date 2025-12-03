package config

import (
	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
	"os"
)

var serveConfig *GlobalConfig

func LoadConfig(configYml string) {
	serveConfig = new(GlobalConfig)
	viper.SetConfigFile(configYml)
	err := viper.ReadInConfig()
	if err != nil {
		println("Config Read failed: " + err.Error())
		os.Exit(1)
	}
	err = viper.Unmarshal(serveConfig)
	if err != nil {
		println("Config Unmarshal failed: " + err.Error())
		os.Exit(1)
	}
	viper.OnConfigChange(func(e fsnotify.Event) {
		println("Config fileHandle changed: ", e.Name)
		_ = viper.ReadInConfig()
		err = viper.Unmarshal(serveConfig)
		if err != nil {
			println("New Config fileHandle Parse Failed: ", e.Name)
			return
		}
	})
	viper.WatchConfig()
}

func GenConfig(configYml string, force bool) error {
	if !FileExist(configYml) || force {
		data, _ := yaml.Marshal(&GlobalConfig{MODE: "dev"})
		err := os.WriteFile(configYml, data, 0644)
		if err != nil {
			return errors.New("Generate file with error: " + err.Error())
		}
		println("config file `config.yaml` generate success in " + configYml)
	} else {
		return errors.New(configYml + " already exist, use -f to Force coverage")
	}
	return nil
}

func GetConfig() *GlobalConfig {
	return serveConfig
}