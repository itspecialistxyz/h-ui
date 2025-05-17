package dao

import (
	"errors"
	"fmt"
	"h-ui/model/constant"
	"log"
	"os"
	"strings"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

var sqlInitStr = "CREATE TABLE IF NOT EXISTS account\n(\n    id             INTEGER PRIMARY KEY AUTOINCREMENT,\n    username       TEXT    NOT NULL UNIQUE DEFAULT '',\n    pass           TEXT    NOT NULL        DEFAULT '',\n    con_pass       TEXT    NOT NULL        DEFAULT '',\n    quota          INTEGER NOT NULL        DEFAULT 0,\n    download       INTEGER NOT NULL        DEFAULT 0,\n    upload         INTEGER NOT NULL        DEFAULT 0,\n    expire_time    INTEGER NOT NULL        DEFAULT 0,\n    kick_util_time INTEGER NOT NULL        DEFAULT 0,\n    device_no      INTEGER NOT NULL        DEFAULT 3,\n    role           TEXT    NOT NULL        DEFAULT 'user',\n    deleted        INTEGER NOT NULL        DEFAULT 0,\n    create_time    TIMESTAMP               DEFAULT CURRENT_TIMESTAMP,\n    update_time    TIMESTAMP               DEFAULT CURRENT_TIMESTAMP\n);\nALTER TABLE account\n    ADD COLUMN login_at INTEGER NOT NULL DEFAULT 0;\nALTER TABLE account\n    ADD COLUMN con_at INTEGER NOT NULL DEFAULT 0;\nCREATE INDEX IF NOT EXISTS account_deleted_index ON account (deleted);\nCREATE INDEX IF NOT EXISTS account_username_index ON account (username);\nCREATE INDEX IF NOT EXISTS account_con_pass_index ON account (con_pass);\nCREATE INDEX IF NOT EXISTS account_pass_index ON account (pass);\nINSERT INTO account (id, username, pass, con_pass, quota, download, upload, expire_time, device_no, role)\nSELECT 1 ,'sysadmin', '02f382b76ca1ab7aa06ab03345c7712fd5b971fb0c0f2aef98bac9cd', 'sysadmin.sysadmin', -1, 0, 0, 253370736000000, 6, 'admin'\n    WHERE NOT EXISTS (SELECT 1 FROM account WHERE id = 1);\nCREATE TABLE IF NOT EXISTS config\n(\n    id          INTEGER PRIMARY KEY AUTOINCREMENT,\n    key         TEXT NOT NULL UNIQUE DEFAULT '',\n    value       TEXT NOT NULL        DEFAULT '',\n    remark      TEXT NOT NULL        DEFAULT '',\n    create_time TIMESTAMP            DEFAULT CURRENT_TIMESTAMP,\n    update_time TIMESTAMP            DEFAULT CURRENT_TIMESTAMP\n);\nCREATE INDEX IF NOT EXISTS config_key_index ON config (key);\nINSERT INTO config (key, value, remark)\nSELECT 'H_UI_WEB_PORT', '8081', 'H UI Web Port'\n    WHERE NOT EXISTS (SELECT 1 FROM config WHERE key = 'H_UI_WEB_PORT');\nINSERT INTO config (key, value, remark)\nSELECT 'H_UI_WEB_CONTEXT', '/', 'H UI Web Context'\n    WHERE NOT EXISTS (SELECT 1 FROM config WHERE key = 'H_UI_WEB_CONTEXT');\nINSERT INTO config (key, value, remark)\nSELECT 'H_UI_CRT_PATH', '', 'H UI Crt File Path'\n    WHERE NOT EXISTS (SELECT 1 FROM config WHERE key = 'H_UI_CRT_PATH');\nINSERT INTO config (key, value, remark)\nSELECT 'H_UI_KEY_PATH', '', 'H UI Key File Path'\n    WHERE NOT EXISTS (SELECT 1 FROM config WHERE key = 'H_UI_KEY_PATH');\nINSERT INTO config (key, value, remark)\nSELECT 'JWT_SECRET', hex(randomblob(10)), 'JWT Secret'\n    WHERE NOT EXISTS (SELECT 1 FROM config WHERE key = 'JWT_SECRET');\nINSERT INTO config (key, value, remark)\nSELECT 'HYSTERIA2_ENABLE', '0', 'Hysteria2 Switch'\n    WHERE NOT EXISTS (SELECT 1 FROM config WHERE key = 'HYSTERIA2_ENABLE');\nINSERT INTO config (key, value, remark)\nSELECT 'HYSTERIA2_CONFIG', '', 'Hysteria2 Config'\n    WHERE NOT EXISTS (SELECT 1 FROM config WHERE key = 'HYSTERIA2_CONFIG');\nINSERT INTO config (key, value, remark)\nSELECT 'HYSTERIA2_TRAFFIC_TIME', '1', 'Hysteria2 Traffic Time'\n    WHERE NOT EXISTS (SELECT 1 FROM config WHERE key = 'HYSTERIA2_TRAFFIC_TIME');\nINSERT INTO config (key, value, remark)\nSELECT 'HYSTERIA2_CONFIG_REMARK', '', 'Hysteria2 Config Remark'\n    WHERE NOT EXISTS (SELECT 1 FROM config WHERE key = 'HYSTERIA2_CONFIG_REMARK');\nINSERT INTO config (key, value, remark)\nSELECT 'HYSTERIA2_CONFIG_PORT_HOPPING', '', 'Hysteria2 Config Port Hopping'\n    WHERE NOT EXISTS (SELECT 1 FROM config WHERE key = 'HYSTERIA2_CONFIG_PORT_HOPPING');\nINSERT INTO config (key, value, remark)\nSELECT 'RESET_TRAFFIC_CRON', '', 'Reset Traffic Cron'\n    WHERE NOT EXISTS (SELECT 1 FROM config WHERE key = 'RESET_TRAFFIC_CRON');\nINSERT INTO config (key, value, remark)\nSELECT 'TELEGRAM_ENABLE', '0', 'Telegram Switch'\n    WHERE NOT EXISTS (SELECT 1 FROM config WHERE key = 'TELEGRAM_ENABLE');\nINSERT INTO config (key, value, remark)\nSELECT 'TELEGRAM_TOKEN', '', 'Telegram Token'\n    WHERE NOT EXISTS (SELECT 1 FROM config WHERE key = 'TELEGRAM_TOKEN');\nINSERT INTO config (key, value, remark)\nSELECT 'TELEGRAM_CHAT_ID', '', 'Telegram ChatId'\n    WHERE NOT EXISTS (SELECT 1 FROM config WHERE key = 'TELEGRAM_CHAT_ID');\nINSERT INTO config (key, value, remark)\nSELECT 'TELEGRAM_LOGIN_JOB_ENABLE', '0', 'TELEGRAM LOGIN Notification'\n    WHERE NOT EXISTS (SELECT 1 FROM config WHERE key = 'TELEGRAM_LOGIN_JOB_ENABLE');\nINSERT INTO config (key, value, remark)\nSELECT 'TELEGRAM_LOGIN_JOB_TEXT', '[time], [username] logged into the panel, IP address is [ip]', 'TELEGRAM LOGIN Notification Text'\n    WHERE NOT EXISTS (SELECT 1 FROM config WHERE key = 'TELEGRAM_LOGIN_JOB_TEXT');\nINSERT INTO config (key, value, remark)\nSELECT 'CLASH_EXTENSION', '', 'Clash Subscription Extension'\n    WHERE NOT EXISTS (SELECT 1 FROM config WHERE key = 'CLASH_EXTENSION');"

var sqliteDB *gorm.DB

func InitSqliteDB() error {
	var err error
	huiData := os.Getenv("HUI_DATA")
	if huiData == "" {
		huiData = "./" // Default to current directory if environment variable is not set
		logrus.Warn("HUI_DATA environment variable not set, using current directory")
	}
	sqliteDB, err = gorm.Open(sqlite.Open(fmt.Sprintf("%s%s", huiData, constant.SqliteDBPath)), &gorm.Config{
		TranslateError: true,
		Logger: logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags),
			logger.Config{
				SlowThreshold:             time.Second,
				LogLevel:                  logger.Silent,
				IgnoreRecordNotFoundError: true,
				ParameterizedQueries:      true,
				Colorful:                  false,
			},
		),
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})
	if err != nil {
		logrus.Errorf("sqlite open err: %v", err)
		return errors.New("sqlite open err")
	}
	return nil
}

func InitSql(port string) error {
	if err := InitSqliteDB(); err != nil {
		return fmt.Errorf("failed to initialize SQLite DB: %w", err)
	}

	if err := sqliteInit(sqlInitStr); err != nil {
		return fmt.Errorf("failed to run SQL initialization: %w", err)
	}

	if port != "" {
		var result string
		db, err := sqliteDB.DB()
		if err != nil {
			return fmt.Errorf("failed to get DB instance: %w", err)
		}

		if err := db.QueryRow("SELECT value from config where key = 'H_UI_WEB_PORT' limit 1").Scan(&result); err != nil {
			logrus.Errorf("sqlite query for H_UI_WEB_PORT err: %v", err)
			return fmt.Errorf("sqlite query for H_UI_WEB_PORT err: %w", err)
		}

		if result == "8081" {
			if tx := sqliteDB.Exec("UPDATE config set value = ? where key = 'H_UI_WEB_PORT'", port); tx.Error != nil {
				logrus.Errorf("sqlite update H_UI_WEB_PORT err: %v", tx.Error)
				return fmt.Errorf("sqlite update H_UI_WEB_PORT err: %w", tx.Error)
			}
		}
	}
	return nil
}

func sqliteInit(sqlStr string) error {
	if sqliteDB == nil {
		return errors.New("database connection not initialized")
	}

	sqls := strings.Split(strings.Replace(sqlStr, "\r\n", "\n", -1), ";\n")
	for i, s := range sqls {
		s = strings.TrimSpace(s)
		if s != "" {
			tx := sqliteDB.Exec(s)
			if tx.Error != nil {
				// Ignore expected errors for duplicate columns
				if strings.HasPrefix(tx.Error.Error(), "SQL logic error: duplicate column name") {
					continue
				}
				logrus.Errorf("sqlite exec err on statement %d: %v", i, tx.Error)
				return fmt.Errorf("sqlite exec err: %w", tx.Error)
			}
		}
	}
	return nil
}

func CloseSqliteDB() error {
	if sqliteDB != nil {
		db, err := sqliteDB.DB()
		if err != nil {
			logrus.Errorf("sqlite get DB instance err: %v", err)
			return fmt.Errorf("sqlite get DB instance err: %w", err)
		}
		if err = db.Close(); err != nil {
			logrus.Errorf("sqlite close err: %v", err)
			return fmt.Errorf("sqlite close err: %w", err)
		}
	}
	return nil
}

func Paginate(pageNum *int64, pageSize *int64) func(db *gorm.DB) *gorm.DB {
	var num int64 = 1
	var size int64 = 10
	if pageNum != nil && *pageNum > 0 {
		num = *pageNum
	}
	if pageSize != nil && *pageSize > 0 {
		size = *pageSize
	}
	return func(db *gorm.DB) *gorm.DB {
		return db.Offset(int((num - 1) * size)).Limit(int(size))
	}
}
