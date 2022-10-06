# WAL-G TELEGRAM BACKUPS INFO
```sh
make env
```

For run project
```sh
make run
```

For build project
```
make build
```

For build docker image
```
make build
docker build -t <image_name>:<version> -f Dockerfile.build .
```

## Environment variables
```env
# Timezone
APP_TIMEZONE=UTC

# For get information about of only full backups set is true
# If set is false, get information for incremental backups too
# default: false
IS_ONLY_FULL_BACKUPS=false
# For save information about of backups to json file in minio to subfolder logs_005 set true
# default: false
IS_SAVE_INFO_TO_JSON=false

# MINIO host, required
MINIO_HOST=<minio_host>
# MINIO username, required
MINIO_ACCESS_KEY=<minio_access_key>
# MINIO password, required
MINIO_SECRET_KEY=<minio_secret_key>
# MINIO bucket name, required
MINIO_BUCKET=<minio_bucket>
# MINIO SECURE, default true
MINIO_SECURE=true

# Path to wal-g binary file, default /bin/wal-g
WALG_BINARY_PATH=/bin/wal-g
# Path to wal-g backups dir on MINIO_BUCKET, example: dev/backups or backups (without / in start)
# !!! IF NOT SUBPATH ON BUCKET, NOT PASS THIS ENV OR PASS EMPTY VALUE
WALG_BACKUPS_PATH=<walg_backups_path>

# Telegram bot token, which send backups info
# required
TELEGRAM_BOT_TOKEN=<token>
# Telegram chats, which gets backups info
# Example: 123456,283213 or 123456
# required
TELEGRAM_CHAT_IDS=<chat_ids>

# cron: Second | Minute | Hour | Dom | Month | Dow
# for execute EXEC_BACKUP command
# example: 0 0 21 * * *
# required
CRON_BACKUPS_INFO=<cron_backups_info>
```
