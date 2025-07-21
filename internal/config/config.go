package config

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env    string       `yaml:"env" env-default:"prod"`
	Server ServerConfig `yaml:"server"`
	DB     DBConfig     `yaml:"db"`
	Redis  RedisConfig  `yaml:"redis"`
	Minio  MinioConfig  `yaml:"minio"`
}

type ServerConfig struct {
	Addr        string        `yaml:"addr" env-default:"0.0.0.0:50070"`
	Timeout     time.Duration `yaml:"timeout" env-default:"4s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"60s"`
}

type DBConfig struct {
	Host           string `env:"DB_HOST" yaml:"host" env-default:"localhost"`
	Port           int    `env:"DB_PORT" yaml:"port" env-default:"5222"`
	User           string `env:"DB_USER" yaml:"user" env-default:"postgres"`
	Password       string `env:"DB_PASSWORD" yaml:"password" env-required:"true"`
	Name           string `env:"DB_NAME" yaml:"name" env-default:"chat"`
	MinPools       int    `env:"DB_MIN_POOLS" yaml:"min_pools" env-default:"3"`
	MaxPools       int    `env:"DB_MAX_POOLS" yaml:"max_pools" env-default:"5"`
	MigrationsPath string `env:"MIGRATIONS_PATH" yaml:"migrations_path" env-default:"./migrations"`
}

type RedisConfig struct {
	Host       string        `env:"REDIS_HOST" yaml:"host" env-default:"localhost"`
	Port       int           `env:"REDIS_PORT" yaml:"port" env-default:"6379"`
	User       string        `env:"REDIS_USER" yaml:"user" env-default:"root"`
	Password   string        `env:"REDIS_PASSWORD" yaml:"password" env-default:"root"`
	DB         int           `env:"REDIS_DB" yaml:"db" env-default:"0"`
	Expiration time.Duration `env:"REDIS_EXPIRATION" yaml:"expire_time" env-default:"24h"`
}

type MinioConfig struct {
	Endpoint   string `env:"MINIO_ENDPOINT" yaml:"endpoint" env-default:"localhost:9000"`
	Port       int    `env:"MINIO_PORT" yaml:"port" env-default:"9000"`
	User       string `env:"MINIO_ROOT_USER" yaml:"user" env-default:"minio"`
	Password   string `env:"MINIO_ROOT_PASSWORD" yaml:"password" env-required:"true"`
	BucketName string `env:"MINIO_BUCKET_NAME" yaml:"bucket_name" env-default:"users"`
	IsUseSsl   bool   `env:"MINIO_USE_SSL" yaml:"is_use_ssl" env-default:"false"`
}

func MustLoad() *Config {
	path := fetchPath()
	cfg, err := Load(path)
	if err != nil {
		panic(err)
	}
	return cfg
}

func Load(path string) (*Config, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file does not exist: %s", path)
	}

	cfg := &Config{}

	if err := cleanenv.ReadConfig(path, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return cfg, nil
}

func fetchPath() string {
	var path string

	flag.StringVar(&path, "config", "", "path to config file")
	flag.Parse()

	if path == "" {
		path = os.Getenv("CONFIG_PATH")
	}

	return path
}
