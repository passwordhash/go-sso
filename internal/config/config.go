package config

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

const (
	envLocal   = "local"
	envDevelop = "dev"
	envProd    = "prod"
)

type Config struct {
	Env      string        `yaml:"env" env:"ENV" env-required:"true"`
	TokenTTL time.Duration `yaml:"token_ttl" env:"TOKEN_TTL" env-required:"true"`
	GRPC     GRPCConfig    `yaml:"grpc" env-required:"true"`
	Vault    VaultConfig   `yaml:"vault" env-required:"true"`
	PSQL     PSQLConfig    `yaml:"psql" env-required:"true"`
}

type GRPCConfig struct {
	Host    string        `yaml:"host" env:"GRPC_HOST" env-required:"true"`
	Port    int           `yaml:"port" env:"GRPC_PORT" env-required:"true"`
	Timeout time.Duration `yaml:"timeout" env:"TIMEOUT" env-required:"true"`
}

type PSQLConfig struct {
	Port     int             `yaml:"port" env:"POSTGRES_PORT" env-required:"true"`
	Host     string          `yaml:"host" env:"POSTGRES_HOST" env-required:"true"`
	User     string          `yaml:"user" env:"POSTGRES_USER" env-required:"true"`
	Pass     string          `yaml:"pass" env:"POSTGRES_PASSWORD" env-required:"true"`
	DB       string          `yaml:"db" env:"POSTGRES_DB" env-required:"true"`
	Migrator *MigratorConfig `yaml:"migrator" `
}

type VaultConfig struct {
	Addr  string `yaml:"addr" env:"VAULT_ADDR" env-required:"true"`
	Token string `yaml:"token" env:"VAULT_TOKEN" env-required:"true"`
}

type MigratorConfig struct {
	Path  string `yaml:"path" env:"MIGRATIONS_PATH" env-required:"true"`
	Table string `yaml:"table" env:"MIGRATIONS_TABLE" env-required:"true"`
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
		panic("failed to load config: " + err.Error())
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
