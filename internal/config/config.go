package config

import (
    "flag"
    "github.com/ilyakaznacheev/cleanenv"
    "os"
    "time"
)

const (
    envLocal   = "local"
    envDevelop = "dev"
    envProd    = "prod"
)

type Config struct {
    Env      string        `yaml:"env" env:"ENV" required:"true"`
    TokenTTL time.Duration `yaml:"token_ttl" env:"TOKEN_TTL" required:"true"`
    GRPC     GRPCConfig    `yaml:"grpc" required:"true"`
}

type GRPCConfig struct {
    Port    int           `yaml:"port" env:"GRPC_PORT" required:"true"`
    Timeout time.Duration `yaml:"timeout" env:"TIMEOUT" required:"true"`
}

func MustLoad() *Config {
    path := fetchConfigPath()
    if path == "" {
        panic("config path is not set")
    }

    if _, err := os.Stat(path); os.IsNotExist(err) {
        panic("config file does not exist: " + path)
    }

    var cfg Config

    if err := cleanenv.ReadConfig(path, &cfg); err != nil {
        panic("failed to read config file: " + err.Error())
    }

    return &cfg
}

// fetchConfigPath получает путь к конфигурационному файлу из флага или переменной окружения.
// Если путь не указан, возвращает пустую строку.
func fetchConfigPath() string {
    var res string

    flag.StringVar(&res, "config", "", "path to config file")
    flag.Parse()

    if res == "" {
        res = os.Getenv("CONFIG_PATH")
    }

    return res
}
