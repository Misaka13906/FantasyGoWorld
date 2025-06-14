package config

import (
	"log"

	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
)

type Config struct {
	HttpPort     string `env:"HTTP_PORT"`
	JwtSecretKey string `env:"JWT_SECRET_KEY"`
	MysqlDb      struct {
		Name string `env:"NAME"`
		User string `env:"USER"`
		Pswd string `env:"PSWD"`
		Host string `env:"HOST"`
		Port string `env:"PORT"`
	} `envPrefix:"DB_"`
	RedisDb struct {
		Host string `env:"HOST"`
		Port string `env:"PORT"`
		Pswd string `env:"PSWD"`
		Db   int    `env:"DB"`
		// Redis 连接配置
		// PoolSize           int `yaml:"pool_size" envDefault:"10"`            // Redis连接池大小
		// MaxRetries         int `yaml:"max_retries" envDefault:"3"`           // Redis最大重试次数
		// MinIdleConns       int `yaml:"min_idle_conns" envDefault:"2"`        // Redis最小空闲连接数
		// IdleTimeout        int `yaml:"idle_timeout" envDefault:"300"`        // Redis空闲连接超时时间（秒）
		// IdleCheckFrequency int `yaml:"idle_check_frequency" envDefault:"60"` // Redis空闲连接检查频率（秒）
	} `envPrefix:"REDIS_"`
}

var (
	Configs Config
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Fatalln("Error loading .env file:", err)
	}

	if err := env.Parse(&Configs); err != nil {
		log.Fatalf("Error parsing .env file into Config struct: %+v\\n", err)
	}

	log.Println("Configuration loaded successfully.")

	if Configs.MysqlDb.Name == "" || Configs.MysqlDb.User == "" || Configs.MysqlDb.Pswd == "" || Configs.MysqlDb.Host == "" || Configs.MysqlDb.Port == "" {
		log.Fatalln("Database configuration is incomplete. Please check your .env file.")
	}
}
