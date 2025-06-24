package config

import (
	"flag"
	"github.com/caarlos0/env/v6"
	"github.com/faust8888/shortener/internal/middleware/logger"
	"go.uber.org/zap"
)

// Константы, используемые для флагов командной строки и параметров конфигурации.
const (
	// ServerAddressFlag — флаг для указания адреса сервера (например, "-a").
	ServerAddressFlag = "a"

	// BaseShortURLFlag — флаг для указания базового URL коротких ссылок (например, "-b").
	BaseShortURLFlag = "b"

	// LoggingLevelFlag — флаг для указания уровня логирования (например, "-l").
	LoggingLevelFlag = "l"

	// StorageFilePathFlag — флаг для указания пути к файлу хранения данных (например, "-f").
	StorageFilePathFlag = "f"

	// DataSourceNameFlag — флаг для указания DSN-строки PostgreSQL (например, "-d").
	DataSourceNameFlag = "d"

	// AuthKeyNameFlag — флаг для указания имени ключа аутентификации (например, "-k").
	AuthKeyNameFlag = "k"

	// HashKeyURLQueryParam — имя параметра запроса, содержащего хэш URL (например, "/{hash}").
	HashKeyURLQueryParam = "hashKeyURL"
)

// Config — это структура, представляющая конфигурацию приложения.
// Поля могут заполняться как через флаги командной строки, так и через переменные окружения.
type Config struct {
	// ServerAddress — адрес и порт, на котором будет запущен сервер (например, "localhost:8080").
	ServerAddress string `env:"SERVER_ADDRESS"`

	// BaseShortURL — базовый URL, используемый для формирования полного адреса короткой ссылки.
	BaseShortURL string `env:"BASE_URL"`

	// LoggingLevel — уровень логирования (например, "INFO", "DEBUG").
	LoggingLevel string `env:"LOGGING_LEVEL"`

	// StorageFilePath — путь к файлу, используемому для хранения данных (если используется файловое хранилище).
	StorageFilePath string `env:"FILE_STORAGE_PATH"`

	// DataSourceName — строка подключения к PostgreSQL (DSN).
	DataSourceName string `env:"DATABASE_DSN"`

	// AuthKey — секретный ключ, используемый для генерации токенов аутентификации.
	AuthKey string `env:"AUTH_KEY"`
}

// Create инициализирует и возвращает объект конфигурации, заполняя его значениями из:
// - флагов командной строки,
// - переменных окружения.
//
// Возвращает:
//   - *Config: указатель на готовую конфигурацию приложения.
func Create() *Config {
	var cfg Config

	setFlag(&cfg.ServerAddress, ServerAddressFlag, "localhost:8080", "Address of the server")
	setFlag(&cfg.BaseShortURL, BaseShortURLFlag, "http://localhost:8080", "Base URL for returning short URL")
	setFlag(&cfg.LoggingLevel, LoggingLevelFlag, "INFO", "Level of logging to use")
	setFlag(&cfg.StorageFilePath, StorageFilePathFlag, "./storage.txt", "Path to the storage file")
	setFlag(&cfg.DataSourceName, DataSourceNameFlag, "", "URL to the running PostgreSQL")
	setFlag(&cfg.AuthKey, AuthKeyNameFlag, "dd109d0b86dc6a06584a835538768c6a2ceb588560755c7f7b90c0bf774237c8", "Auth Key for authentication")
	flag.Parse()

	err := env.Parse(&cfg)
	if err != nil {
		logger.Log.Error("Failed to parse config", zap.Error(err))
	}

	return &cfg
}

func setFlag(p *string, flagName string, defaultFlagValue string, description string) {
	if flag.Lookup(flagName) == nil {
		flag.StringVar(p, flagName, defaultFlagValue, description)
	} else {
		*p = flag.Lookup(flagName).Value.String()
	}
}
