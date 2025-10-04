package config

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct{
	Env string `yaml:"env" env-default:"local"`
	StoragePath string `yaml:"storage_path" env-required:"true"`
	TokenTTL time.Duration `yaml:"token_ttl" env-required:"true"`
	GRPC `yaml:"grpc"`
}

type GRPC struct{
	Port int `yaml:"port"`
	Timeout time.Duration `yaml:"timeout"`
}


func MustLoad() *Config {
	path := FetchConfigPath()
	fmt.Println()
	if path == "" {
		panic("config path is empty")
	}

	return MustLoadByPath(path)
}


// fetchConfigPath fetches config path from command line flag or environment variable
// Priority: flag > env > default.
// Default value is empty string.

func FetchConfigPath() string {
	var res string
	// Первым аргументом мы передаем переменную в которую будет записанно значение -
	// флага. Вторым аргументом имя флага (--config="path/to/config.yaml"), значение по -
	// умолчанию будет пустое, подсказка для командной строки что это путь до config файла

	flag.StringVar(&res, "config", "", "path to config file")
	// Parse parses the command-line flags from os.Args[1:]
	flag.Parse()

	// Если после выполнения парсинга путь пустой мы прочитаем переменную окружения
	if res == "" {
		res = os.Getenv("CONFIG_PATH")
	}
	return res
}


func MustLoadByPath(configPath string) *Config {

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		panic("config path does no exist:" + configPath)
	}
	
	// Проверям существует ли файл
	if _, err := os.Stat(configPath); os.IsNotExist(err){
		panic("config file does not exist: " + configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath,  &cfg); err != nil{
		panic("failed to read config: " + err.Error())
	}
	return &cfg
}