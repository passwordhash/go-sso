package config

import (
    "flag"
    "fmt"
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
    PSQL     PSQLConfig    `yaml:"psql" required:"true"`
}

type GRPCConfig struct {
    Host    string        `yaml:"host" env:"GRPC_HOST" required:"true"`
    Port    int           `yaml:"port" env:"GRPC_PORT" required:"true"`
    Timeout time.Duration `yaml:"timeout" env:"TIMEOUT" required:"true"`
}

type PSQLConfig struct {
    Port int    `yaml:"port" env:"POSTGRES_PORT" required:"true"`
    Host string `yaml:"host" env:"POSTGRES_HOST" required:"true"`
    User string `yaml:"user" env:"POSTGRES_USER" required:"true"`
    Pass string `yaml:"pass" env:"POSTGRES_PASSWORD" required:"true"`
    DB   string `yaml:"db" env:"POSTGRES_DB" required:"true"`
}

func (c *PSQLConfig) DSN() string {
    return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
        c.User,
        c.Pass,
        c.Host,
        c.Port,
        c.DB,
    )
}

func MustLoad() *Config {
    path := fetchConfigPath()
    if path == "" {
        panic("config path is not set")
    }

    return MustLoadByPath(path)
}

func MustLoadByPath(configPath string) *Config {
    if _, err := os.Stat(configPath); os.IsNotExist(err) {
        panic("config file does not exist: " + configPath)
    }

    var cfg Config

    if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
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
