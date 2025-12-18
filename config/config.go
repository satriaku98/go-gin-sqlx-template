package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	Environment        string `mapstructure:"ENVIRONMENT"`
	ServerPort         string `mapstructure:"SERVER_PORT"`
	DBHost             string `mapstructure:"DB_HOST"`
	DBPort             string `mapstructure:"DB_PORT"`
	DBUser             string `mapstructure:"DB_USER"`
	DBPassword         string `mapstructure:"DB_PASSWORD"`
	DBName             string `mapstructure:"DB_NAME"`
	RedisHost          string `mapstructure:"REDIS_HOST"`
	RedisPort          string `mapstructure:"REDIS_PORT"`
	RedisDB            int    `mapstructure:"REDIS_DB"`
	RedisPassword      string `mapstructure:"REDIS_PASSWORD"`
	TelegramBaseURL    string `mapstructure:"TELEGRAM_BASE_URL"`
	TelegramToken      string `mapstructure:"TELEGRAM_TOKEN"`
	TelegramChatID     string `mapstructure:"TELEGRAM_CHAT_ID"`
	ServiceName        string `mapstructure:"SERVICE_NAME"`
	WorkerName         string `mapstructure:"WORKER_NAME"`
	PubSubEmulatorHost string `mapstructure:"PUBSUB_EMULATOR_HOST"`
	PubSubProjectID    string `mapstructure:"PUBSUB_PROJECT_ID"`
	PubSubCredsFile    string `mapstructure:"PUBSUB_CREDS_FILE"`
}

func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName(".env")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		// It's okay if config file doesn't exist, we might be using env vars
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return
		}
	}

	err = viper.Unmarshal(&config)
	return
}
