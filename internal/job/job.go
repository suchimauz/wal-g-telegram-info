package job

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/suchimauz/wal-g-telegram-info/internal/config"
	"github.com/suchimauz/wal-g-telegram-info/pkg/logger"
	"github.com/suchimauz/wal-g-telegram-info/pkg/storage"
	"github.com/suchimauz/wal-g-telegram-info/pkg/utils"
)

// InfoJob - struct for manage job, which send notifications of backups and etc
type InfoJob struct {
	Storage        storage.Provider
	Config         *config.Config
	TelegramBotApi *tgbotapi.BotAPI
	commandsEnv    []string
}

// Constructor
func NewInfoJob(cfg *config.Config, botapi *tgbotapi.BotAPI, storage storage.Provider) *InfoJob {
	// Init envs for wal-g command
	var commandsEnv []string
	commandsEnv = append(commandsEnv, "AWS_ACCESS_KEY="+cfg.Minio.AccessKey)
	commandsEnv = append(commandsEnv, "AWS_SECRET_KEY="+cfg.Minio.SecretKey)
	commandsEnv = append(commandsEnv, "AWS_ENDPOINT=http://"+cfg.Minio.Endpoint)
	commandsEnv = append(commandsEnv, "AWS_S3_FORCE_PATH_STYLE=true")
	commandsEnv = append(commandsEnv, "AWS_REGION=eu-west-1")
	commandsEnv = append(commandsEnv, fmt.Sprintf("WALG_S3_PREFIX=http://%s/%s", cfg.Minio.Bucket, cfg.WalG.BackupsPath))

	return &InfoJob{
		Config:         cfg,
		TelegramBotApi: botapi,
		Storage:        storage,
		commandsEnv:    commandsEnv,
	}
}

// Main func for Run this job, implements for cron.Job interface
func (ij *InfoJob) Run() {
	logger.Info("[NotifierJob] Start processing Job!")

	// Init wal-g backup-list --json --pretty --detail command
	backupListCmd := exec.Command(ij.Config.WalG.BinaryPath, "backup-list", "--json", "--pretty", "--detail")
	backupListCmd.Env = ij.commandsEnv

	// set var to get the output
	var backupListOut bytes.Buffer

	// set the output to our variable
	backupListCmd.Stdout = &backupListOut
	err := backupListCmd.Run()
	if err != nil {
		logger.Errorf("[NotifierJob] Run wal-g command: %s", err.Error())
		return
	}
	// Parse backups info json to array of objects
	backupsInfo, err := parseBackupsInfoJson(backupListOut.String())
	if err != nil {
		logger.Errorf("[NotifierJob] parse json: %s", err.Error())
		logger.Error("[NotifierJob] Exit Job!")

		return
	}

	ij.sendNotifications(backupsInfo)

	// Save backupsInfo log file to storage
	err = ij.saveBackupsInfoFile(backupsInfo)
	if err != nil {
		logger.Errorf("[NotifierJob] Error on upload file: %s", err.Error())
	}

	logger.Info("[NotifierJob] End processing job!")
}

// Private method for send telegram notifications
func (ij *InfoJob) sendNotifications(bi []*BackupInfo) {
	var msg string
	var backups []*BackupInfo

	// Get only full backups
	fullBackupsInfo := getOnlyFullBackups(bi)

	if ij.Config.IsOnlyFullBackups {
		backups = append(backups, fullBackupsInfo...)
	} else {
		backups = append(backups, bi...)
	}

	// If not full backups send message for users that backups is not exists
	// Else send backups info
	msg = MakeBackupsInfoMessage(backups)
	if len(fullBackupsInfo) < 1 {
		logger.Warn("[NotifierJob] Full backups not found!")
		logger.Info("[NotifierJob] Send notifications of backups not found!")

		msg = "<b>Список бэкапов:</b>"
		msg += "\n<code>-------------------</code>"
		msg += "\nБэкапы отсутствуют"
	}

	// Iterate with config users chat-ids, who get notifications
	for _, chatId := range ij.Config.Telegram.ChatIds {
		tgmsg := tgbotapi.NewMessage(chatId, msg)
		tgmsg.ParseMode = "HTMl"
		tgmsg.DisableNotification = true

		go func(gij *InfoJob, gtgmsg tgbotapi.MessageConfig) {
			_, err := gij.TelegramBotApi.Send(gtgmsg)
			if err != nil {
				logger.Errorf("[NotifierJob] Can't send tg notification: %s", err.Error())
			}
		}(ij, tgmsg)
	}
}

// Help func for make message
func MakeBackupsInfoMessage(bi []*BackupInfo) string {
	msg := "<b>Список бэкапов:</b>"

	for _, backupInfo := range bi {
		// Bytes to Gigabytes
		backupSize := float32(backupInfo.CompressedSize) / (1024 * 1024 * 1024)

		msg += "\n<code>-------------------</code>"
		msg += fmt.Sprintf("\nНазвание: <b>%s</b>", backupInfo.BackupName)
		msg += fmt.Sprintf("\nДата: %s", backupInfo.Time.In(config.TimeZone).Format("02.01.2006 15:04"))
		msg += fmt.Sprintf("\nРазмер бэкапа: <b>%0.2f GB</b>", backupSize)
	}

	return msg
}

// Private function for save backups info to storage
func (ij *InfoJob) saveBackupsInfoFile(bi []*BackupInfo) error {
	var backups []*BackupInfo

	fullBackupsInfo := getOnlyFullBackups(bi)

	if ij.Config.IsOnlyFullBackups {
		backups = append(backups, fullBackupsInfo...)
	} else {
		backups = append(backups, bi...)
	}

	// Parse backups to json
	backupsInfoJson, err := json.Marshal(backups)
	if err != nil {
		return err
	}

	// Get storage from InfoJob object
	s3 := ij.Storage

	// Init new empty UploadInput object
	file := storage.UploadInput{
		Name: fmt.Sprintf("%s/logs_005/backups_%s.json",
			ij.Config.WalG.BackupsPath,
			utils.NowDateTz().Format("2006_01_02T15_04_05")),
		ContentType: "application/octet-stream",
		Size:        int64(len(backupsInfoJson)),
		File:        bytes.NewReader(backupsInfoJson),
	}

	// Upload file to storage
	path, err := s3.Upload(context.TODO(), file)
	if err != nil {
		return err
	}

	logger.Infof("[FileStorage] Save backups info: %s", path)

	return nil
}
