package security

import (
	"github.com/zero-net-panel/zero-net-panel/internal/repository"
	"github.com/zero-net-panel/zero-net-panel/internal/types"
)

func toSecuritySetting(setting repository.SecuritySetting) types.SecuritySetting {
	return types.SecuritySetting{
		ID:                   setting.ID,
		ThirdPartyAPIEnabled: setting.ThirdPartyAPIEnabled,
		APIKey:               setting.APIKey,
		APISecret:            setting.APISecret,
		EncryptionAlgorithm:  setting.EncryptionAlgorithm,
		NonceTTLSeconds:      setting.NonceTTLSeconds,
		CreatedAt:            setting.CreatedAt.Unix(),
		UpdatedAt:            setting.UpdatedAt.Unix(),
	}
}
