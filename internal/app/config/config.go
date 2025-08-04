// Package config предоставляет управление конфигурацией для приложения.
// Он загружает настройки из командных флагов, переменных окружения и JSON-файла,
// обеспечивая правильный приоритет: флаги > переменные окружения > JSON-файл > значения по умолчанию.
package config

import (
	"encoding/json"
	"flag"
	"io"
	"os"
	"sync"

	"github.com/caarlos0/env/v6"
	"github.com/faust8888/shortener/internal/middleware/logger"
	"go.uber.org/zap"
)

// Константы для флагов командной строки.
const (
	// ServerAddressFlag - флаг для адреса сервера (-a).
	ServerAddressFlag = "a"
	// BaseShortURLFlag - флаг для базового URL сокращенных ссылок (-b).
	BaseShortURLFlag = "b"
	// LoggingLevelFlag - флаг для уровня логирования (-l).
	LoggingLevelFlag = "l"
	// StorageFilePathFlag - флаг для пути к файлу хранения (-f).
	StorageFilePathFlag = "f"
	// DataSourceNameFlag - флаг для строки подключения к БД (-d).
	DataSourceNameFlag = "d"
	// AuthKeyNameFlag - флаг для ключа аутентификации (-k).
	AuthKeyNameFlag = "k"
	// EnableTLSOnServerFlag - флаг для включения HTTPS (-s).
	EnableTLSOnServerFlag = "s"
	// ConfigFileFlag - флаг для пути к файлу конфигурации (-c).
	ConfigFileFlag = "c"
	// TrustedSubnetFlag - флаг для указания доверенной подсети (CIDR) для административного доступа (-t).
	TrustedSubnetFlag = "t"
	// ConfigFileFlagAlias - псевдоним флага для пути к файлу конфигурации (-config).
	ConfigFileFlagAlias = "config"
	// HashKeyURLQueryParam - имя параметра URL, содержащего хэш.
	HashKeyURLQueryParam = "hashKeyURL"
)

// Config хранит все настройки конфигурации приложения.
type Config struct {
	// ServerAddress - сетевой адрес и порт для запуска сервера (флаг -a, env SERVER_ADDRESS).
	ServerAddress string `env:"SERVER_ADDRESS" json:"server_address"`
	// BaseShortURL - базовый URL для формирования сокращенных ссылок (флаг -b, env BASE_URL).
	BaseShortURL string `env:"BASE_URL" json:"base_url"`
	// LoggingLevel - уровень логирования (флаг -l, env LOGGING_LEVEL).
	LoggingLevel string `env:"LOGGING_LEVEL"`
	// StorageFilePath - путь к файлу для хранения данных, если не используется БД (флаг -f, env FILE_STORAGE_PATH).
	StorageFilePath string `env:"FILE_STORAGE_PATH" json:"file_storage_path"`
	// DataSourceName - строка подключения к базе данных PostgreSQL (флаг -d, env DATABASE_DSN).
	DataSourceName string `env:"DATABASE_DSN" json:"database_dsn"`
	// AuthKey - секретный ключ для подписи токенов аутентификации (флаг -k, env AUTH_KEY).
	AuthKey string `env:"AUTH_KEY"`
	// EnableHTTPS - флаг, включающий HTTPS на сервере (флаг -s, env ENABLE_HTTPS).
	EnableHTTPS bool `env:"ENABLE_HTTPS" json:"enable_https"`
	// TrustedSubnet - CIDR-подсеть, в пределах которой разрешён доступ к административным функциям (флаг -t, env TRUSTED_SUBNET).
	TrustedSubnet string `env:"TRUSTED_SUBNET" json:"trusted_subnet"`
}

// JSONConfig - это вспомогательная структура для разбора конфигурации из JSON-файла.
// Использование указателей позволяет отличить отсутствующее в JSON поле от поля с нулевым значением
// (например, пустой строки или false).
type JSONConfig struct {
	ServerAddress   *string `json:"server_address"`
	BaseShortURL    *string `json:"base_url"`
	StorageFilePath *string `json:"file_storage_path"`
	DataSourceName  *string `json:"database_dsn"`
	EnableHTTPS     *bool   `json:"enable_https"`
	TrustedSubnet   *string `json:"trusted_subnet"`
}

var (
	// cfg - глобальный синглтон-экземпляр конфигурации приложения.
	cfg *Config
	// once используется для гарантии того, что инициализация конфигурации произойдет только один раз.
	once sync.Once
)

// Create инициализирует и возвращает синглтон-объект конфигурации.
//
// Функция определяет настройки, считывая их из различных источников
// в следующем порядке приоритета (от высшего к низшему):
//  1. Флаги командной строки (например, -a, -b).
//  2. Переменные окружения (например, SERVER_ADDRESS, BASE_URL).
//  3. Файл конфигурации в формате JSON (путь к которому задается флагом -c/-config или переменной окружения CONFIG).
//  4. Значения по умолчанию, заданные в коде.
//
// Благодаря использованию sync.Once, логика инициализации выполняется только при первом вызове.
// Все последующие вызовы мгновенно возвращают уже настроенный экземпляр.
func Create() *Config {
	once.Do(func() {
		configFilePath := findConfigPathUsingFlags()

		if configFilePath == "" {
			configFilePath = os.Getenv("CONFIG")
		}

		cfg = defaultConfig()

		if configFilePath != "" {
			cfg.applyJSONConfig(configFilePath)
		}

		if err := env.Parse(cfg); err != nil {
			logger.Log.Error("Failed to parse environment variables", zap.Error(err))
		}

		defineGlobalFlags()

		flag.Parse()
	})

	return cfg
}

// defaultConfig создает новый экземпляр Config со значениями по умолчанию.
func defaultConfig() *Config {
	return &Config{
		ServerAddress:   "localhost:8080",
		BaseShortURL:    "http://localhost:8080",
		LoggingLevel:    "INFO",
		StorageFilePath: "./storage.txt",
		DataSourceName:  "",
		AuthKey:         "dd109d0b86dc6a06584a835538768c6a2ceb588560755c7f7b90c0bf774237c8",
		EnableHTTPS:     false,
		TrustedSubnet:   "",
	}
}

// findConfigPathUsingFlags использует временный, изолированный FlagSet для поиска
// пути к файлу конфигурации в аргументах командной строки. Это позволяет найти путь
// до основного парсинга флагов и избежать ошибок о неопределенных флагах.
func findConfigPathUsingFlags() string {
	configFlagSet := flag.NewFlagSet("config", flag.ContinueOnError)
	// Перенаправляем вывод ошибок этого временного FlagSet в "никуда",
	// чтобы он не засорял консоль при встрече с флагами, которые он не знает (-a, -b и т.д.).
	configFlagSet.SetOutput(io.Discard)

	var configPath string
	configFlagSet.StringVar(&configPath, ConfigFileFlag, "", "Path to JSON config file")
	configFlagSet.StringVar(&configPath, ConfigFileFlagAlias, "", "Path to JSON config file (alias)")

	// Парсим аргументы. Ошибки будут проигнорированы и не будут выведены в консоль.
	_ = configFlagSet.Parse(os.Args[1:])

	return configPath
}

// applyJSONConfig читает конфигурационный файл JSON по указанному пути
// и применяет его настройки к экземпляру Config.
// Метод перезаписывает поля только в том случае, если они присутствуют в JSON-файле.
func (c *Config) applyJSONConfig(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		logger.Log.Warn("Failed to read config file, skipping", zap.String("path", path), zap.Error(err))
		return
	}

	var jsonCfg JSONConfig
	if err := json.Unmarshal(data, &jsonCfg); err != nil {
		logger.Log.Warn("Failed to parse JSON config file, skipping", zap.String("path", path), zap.Error(err))
		return
	}

	if jsonCfg.ServerAddress != nil {
		c.ServerAddress = *jsonCfg.ServerAddress
	}
	if jsonCfg.BaseShortURL != nil {
		c.BaseShortURL = *jsonCfg.BaseShortURL
	}
	if jsonCfg.StorageFilePath != nil {
		c.StorageFilePath = *jsonCfg.StorageFilePath
	}
	if jsonCfg.DataSourceName != nil {
		c.DataSourceName = *jsonCfg.DataSourceName
	}
	if jsonCfg.EnableHTTPS != nil {
		c.EnableHTTPS = *jsonCfg.EnableHTTPS
	}
	if jsonCfg.TrustedSubnet != nil {
		c.TrustedSubnet = *jsonCfg.TrustedSubnet
	}
}

// defineGlobalFlags определяет все флаги командной строки приложения в глобальном наборе flag.CommandLine.
// В качестве значений по умолчанию для флагов используются уже загруженные значения из cfg.
// Это обеспечивает правильный порядок приоритетов при вызове flag.Parse().
func defineGlobalFlags() {
	flag.StringVar(&cfg.ServerAddress, ServerAddressFlag, cfg.ServerAddress, "Address of the server (ex: localhost:8080)")
	flag.StringVar(&cfg.BaseShortURL, BaseShortURLFlag, cfg.BaseShortURL, "Base URL for short links (ex: http://localhost:8080)")
	flag.StringVar(&cfg.StorageFilePath, StorageFilePathFlag, cfg.StorageFilePath, "Path to the storage file")
	flag.StringVar(&cfg.DataSourceName, DataSourceNameFlag, cfg.DataSourceName, "Data Source Name for PostgreSQL (ex: postgres://user:pass@host:port/db)")
	flag.BoolVar(&cfg.EnableHTTPS, EnableTLSOnServerFlag, cfg.EnableHTTPS, "Enable HTTPS")
	flag.StringVar(&cfg.LoggingLevel, LoggingLevelFlag, cfg.LoggingLevel, "Level of logging to use")
	flag.StringVar(&cfg.AuthKey, AuthKeyNameFlag, cfg.AuthKey, "Auth Key for authentication")
	flag.StringVar(&cfg.TrustedSubnet, TrustedSubnetFlag, cfg.TrustedSubnet, "Trusted Subnet")

	// Определяем флаг -c/-config здесь еще раз, чтобы он отображался в справке (-h).
	// Его значение нам уже не нужно, так как мы его получили ранее.
	var dummyConfigPath string
	flag.StringVar(&dummyConfigPath, ConfigFileFlag, "", "Path to JSON config file")
	flag.StringVar(&dummyConfigPath, ConfigFileFlagAlias, "", "Path to JSON config file (alias)")
}
