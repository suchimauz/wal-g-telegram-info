package app

import (
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/suchimauz/wal-g-telegram-info/internal/config"
	"github.com/suchimauz/wal-g-telegram-info/internal/job"
	"github.com/suchimauz/wal-g-telegram-info/pkg/logger"
	"github.com/suchimauz/wal-g-telegram-info/pkg/storage"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	cr "github.com/robfig/cron/v3"
)

func Run() {
	// Initialize config
	cfg, err := config.NewConfig()
	if err != nil {
		logger.Errorf("[ENV] %s", err.Error())

		return
	}

	// Create new http client for telegram api
	tgclient := &http.Client{}

	// When http proxy for telegram api is declared,
	if cfg.Telegram.HttpProxy != "" {
		proxy, err := url.Parse(cfg.Telegram.HttpProxy)
		if err != nil {
			logger.Errorf("[TelegramBotApi] Proxy: %s", err.Error())
		}

		transport := &http.Transport{}
		transport.Proxy = http.ProxyURL(proxy)

		tgclient.Transport = transport
	}

	tgbot, err := tgbotapi.NewBotAPIWithClient(cfg.Telegram.BotToken, cfg.Telegram.ApiEndpoint, tgclient)
	if err != nil {
		logger.Errorf("[TelegramBotApi] %s", err.Error())

		return
	}

	// Make new cron object, calls constructor
	cron := cr.New(cr.WithSeconds(), cr.WithLocation(cfg.Timezone))

	storageProvider, err := newStorageProvider(cfg)
	if err != nil {
		logger.Errorf("[FileStorage] Provider: %s", err.Error())

		return
	}

	ij := job.NewInfoJob(cfg, tgbot, storageProvider)

	jobId, err := cron.AddJob(cfg.CronSpec, ij)
	if err != nil {
		logger.Errorf("[Cron] AddJob: %s", err.Error())
	}

	cron.Start()

	logger.Info("[Cron] Started job ", jobId)

	// Graceful Shutdown

	// Make new channel of size = 1
	quit := make(chan os.Signal, 1)

	// Listen system 15 and 2 signals, when one of they called, send info to quit channel
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	// Read channel, this block of code lock this thread, until someone writes to the channel
	<-quit

	// When someone call SIGTERM or SIGINT signals, we'll get to here
	// cron.Stop() -> Wait jobs and stop
	cron.Stop()

	logger.Info("[Cron] Stopped! Exit")
}

func newStorageProvider(cfg *config.Config) (storage.Provider, error) {
	client, err := minio.New(cfg.Minio.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.Minio.AccessKey, cfg.Minio.SecretKey, ""),
		Secure: cfg.Minio.Secure,
	})
	if err != nil {
		return nil, err
	}

	provider := storage.NewFileStorage(client, cfg.Minio.Bucket, cfg.Minio.Endpoint)

	return provider, nil
}
