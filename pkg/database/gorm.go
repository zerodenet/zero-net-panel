package database

import (
	"fmt"
	"strings"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type CloseFunc func()

func NewGorm(cfg Config) (*gorm.DB, CloseFunc, error) {
	if cfg.IsEmpty() {
		return nil, func() {}, nil
	}

	var (
		db  *gorm.DB
		err error
	)

	logLevel := logger.Silent
	switch strings.ToLower(cfg.LogLevel) {
	case "error":
		logLevel = logger.Error
	case "warn":
		logLevel = logger.Warn
	case "info":
		logLevel = logger.Info
	case "silent":
		logLevel = logger.Silent
	default:
		if cfg.LogLevel != "" {
			return nil, nil, fmt.Errorf("unsupported gorm log level: %s", cfg.LogLevel)
		}
	}

	gormConfig := &gorm.Config{Logger: logger.Default.LogMode(logLevel)}

	switch strings.ToLower(cfg.Driver) {
	case "mysql":
		db, err = gorm.Open(mysql.Open(cfg.DSN), gormConfig)
	case "postgres", "postgresql":
		db, err = gorm.Open(postgres.Open(cfg.DSN), gormConfig)
	default:
		return nil, nil, fmt.Errorf("unsupported database driver: %s", cfg.Driver)
	}

	if err != nil {
		return nil, nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, nil, err
	}

	if cfg.MaxOpenConns > 0 {
		sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns > 0 {
		sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	}
	if cfg.ConnMaxLifetime > 0 {
		sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	}

	closeFn := func() {
		_ = sqlDB.Close()
	}

	return db, closeFn, nil
}
