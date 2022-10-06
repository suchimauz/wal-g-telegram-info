package config

import (
	"time"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

const (
	appName = "golang_template"
)

// Not modify this variable!!!
// This variable will be filled when initializing the config
var TimeZone *time.Location

type (
	Config struct {
		timezone          string `envconfig:"app_timezone" default:"UTC"` // String timezone format
		Timezone          *time.Location
		CronSpec          string `envconfig:"cron_backups_info" required:"true"`
		IsOnlyFullBackups bool   `envconfig:"is_only_full_backups" default:"false"`
		WalG              WalGConfig
		Minio             MinioConfig
		Telegram          TelegramConfig
	}
	WalGConfig struct {
		BinaryPath  string `envconfig:"walg_binary_path" default:"/bin/wal-g"`
		BackupsPath string `envconfig:"walg_backups_path"`
	}
	MinioConfig struct {
		Endpoint  string `envconfig:"minio_host" required:"true"`
		AccessKey string `envconfig:"minio_access_key" required:"true"`
		SecretKey string `envconfig:"minio_secret_key" required:"true"`
		Bucket    string `envconfig:"minio_bucket" required:"true"`
		Secure    bool   `envconfig:"minio_secure" default:"true"`
	}
	TelegramConfig struct {
		ApiEndpoint string  `envconfig:"tg_bot_api_endpoint" default:"https://api.telegram.org/bot%s/%s"`
		BotToken    string  `envconfig:"telegram_bot_token" required:"true"`
		HttpProxy   string  `envconfig:"telegram_http_proxy"`
		ChatIds     []int64 `envconfig:"telegram_chat_ids" split_words:"true" required:"true"`
	}
)

func NewConfig() (*Config, error) {
	var cfg Config

	godotenv.Load()

	// Parse variables from environment or return err
	err := envconfig.Process(appName, &cfg)
	if err != nil {
		return nil, err
	}

	// Parse timezone from cfg.tz or return err
	cfg.Timezone, err = time.LoadLocation(cfg.timezone)
	if err != nil {
		return nil, err
	}
	// Parse timezone from cfg.tz or return err
	TimeZone, err = time.LoadLocation(cfg.timezone)
	if err != nil {
		return nil, err
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (cfg *Config) validate() error {
	// pass some validations here
	return nil
}
