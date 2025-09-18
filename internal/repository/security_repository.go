package repository

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
)

// SecuritySetting stores API 安全开关与凭据配置。
type SecuritySetting struct {
	ID                   uint64 `gorm:"primaryKey"`
	ThirdPartyAPIEnabled bool   `gorm:"column:third_party_api_enabled"`
	APIKey               string `gorm:"size:128"`
	APISecret            string `gorm:"size:256"`
	EncryptionAlgorithm  string `gorm:"size:32"`
	NonceTTLSeconds      int    `gorm:"column:nonce_ttl_seconds"`
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

// TableName custom binding.
func (SecuritySetting) TableName() string { return "security_settings" }

// SecurityRepository exposes accessors for安全配置。
type SecurityRepository interface {
	GetThirdPartyAPIConfig(ctx context.Context) (SecuritySetting, error)
	UpsertThirdPartyAPIConfig(ctx context.Context, setting SecuritySetting) (SecuritySetting, error)
}

type securityRepository struct {
	db *gorm.DB
}

// NewSecurityRepository constructs repo.
func NewSecurityRepository(db *gorm.DB) (SecurityRepository, error) {
	if db == nil {
		return nil, errors.New("repository: database connection is required")
	}
	return &securityRepository{db: db}, nil
}

func (r *securityRepository) GetThirdPartyAPIConfig(ctx context.Context) (SecuritySetting, error) {
	if err := ctx.Err(); err != nil {
		return SecuritySetting{}, err
	}

	var setting SecuritySetting
	if err := r.db.WithContext(ctx).Limit(1).First(&setting).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			now := time.Now().UTC()
			setting = SecuritySetting{
				ThirdPartyAPIEnabled: false,
				EncryptionAlgorithm:  "aes-gcm",
				NonceTTLSeconds:      300,
				CreatedAt:            now,
				UpdatedAt:            now,
			}
			if err := r.db.WithContext(ctx).Create(&setting).Error; err != nil {
				return SecuritySetting{}, err
			}
			return setting, nil
		}
		return SecuritySetting{}, err
	}

	if setting.EncryptionAlgorithm == "" {
		setting.EncryptionAlgorithm = "aes-gcm"
	}
	if setting.NonceTTLSeconds == 0 {
		setting.NonceTTLSeconds = 300
	}
	return setting, nil
}

func (r *securityRepository) UpsertThirdPartyAPIConfig(ctx context.Context, setting SecuritySetting) (SecuritySetting, error) {
	if err := ctx.Err(); err != nil {
		return SecuritySetting{}, err
	}

	now := time.Now().UTC()
	setting.UpdatedAt = now
	if setting.EncryptionAlgorithm == "" {
		setting.EncryptionAlgorithm = "aes-gcm"
	}
	if setting.NonceTTLSeconds == 0 {
		setting.NonceTTLSeconds = 300
	}

	if setting.ID == 0 {
		setting.CreatedAt = now
		if err := r.db.WithContext(ctx).Create(&setting).Error; err != nil {
			return SecuritySetting{}, err
		}
		return setting, nil
	}

	if err := r.db.WithContext(ctx).Model(&SecuritySetting{}).
		Where("id = ?", setting.ID).
		Updates(map[string]any{
			"third_party_api_enabled": setting.ThirdPartyAPIEnabled,
			"api_key":                 setting.APIKey,
			"api_secret":              setting.APISecret,
			"encryption_algorithm":    setting.EncryptionAlgorithm,
			"nonce_ttl_seconds":       setting.NonceTTLSeconds,
			"updated_at":              setting.UpdatedAt,
		}).Error; err != nil {
		return SecuritySetting{}, err
	}

	return r.GetThirdPartyAPIConfig(ctx)
}
