package conf

import (
	"encoding/json"

	"github.com/kelseyhightower/envconfig"
	"github.com/spf13/viper"
	"github.com/starine/aim/logger"
)

// Config Config
type Config struct {
	Listen    string `default:":8100"`
	ConsulURL string `default:"localhost:8500"`
	LogLevel  string `default:"INFO"`
}

func (c Config) String() string {
	bts, _ := json.Marshal(c)
	return string(bts)
}

// Init InitConfig
func Init(file string) (*Config, error) {
	viper.SetConfigFile(file)
	viper.AddConfigPath(".")
	viper.AddConfigPath("/etc/conf")

	var config Config

	err := envconfig.Process("aim", &config)
	if err != nil {
		return nil, err
	}

	if err := viper.ReadInConfig(); err != nil {
		logger.Warn(err)
	} else {
		if err := viper.Unmarshal(&config); err != nil {
			return nil, err
		}
	}

	return &config, nil
}
