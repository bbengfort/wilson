package config

const Prefix = "wilson"

type Config struct{}

func New() (*Config, error) {
	return &Config{}, nil
}
