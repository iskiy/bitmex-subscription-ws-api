package config

import (
	"errors"
	"github.com/joho/godotenv"
	"os"
)
type Config struct {
	ApiKey		string
	ApiSecret	string
}

func NewConfig(configSource string) (*Config, error) {
	err := godotenv.Load(configSource)
	if err != nil{
		return nil, err
	}

	apiKey, ok := os.LookupEnv("API_KEY")
	if !ok{
		return nil, errors.New("can`t find API_KEY env variable")
	}

	apiSecret, ok := os.LookupEnv("API_SECRET")
	if !ok{
		return nil, errors.New("can`t find API_SECRET env variable")
	}

	return &Config{ApiKey: apiKey, ApiSecret: apiSecret}, nil
}
