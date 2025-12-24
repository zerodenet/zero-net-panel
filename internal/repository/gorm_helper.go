package repository

import (
	"errors"
	"strings"

	"gorm.io/gorm"
)

func translateError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrNotFound
	}
	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "duplicate") || strings.Contains(msg, "unique constraint") {
		return ErrConflict
	}
	return err
}
