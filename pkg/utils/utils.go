package utils

import (
	"time"

	"github.com/suchimauz/wal-g-telegram-info/internal/config"
)

// Func for get now datetime with timezone in cfg
func NowDateTz() time.Time {
	return time.Now().In(config.TimeZone)
}
