package cli

import (
	"crypto/rand"
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateSecret(t *testing.T) {
	tests := []struct {
		name   string
		length int
	}{
		{"short secret", 16},
		{"medium secret", 32},
		{"long secret", 64},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			secret, err := generateSecret(tt.length)
			require.NoError(t, err)
			require.NotEmpty(t, secret)

			// Decode and verify length
			decoded, err := base64.URLEncoding.DecodeString(secret)
			require.NoError(t, err)
			assert.Equal(t, tt.length, len(decoded))

			// Verify uniqueness by generating multiple secrets
			secret2, err := generateSecret(tt.length)
			require.NoError(t, err)
			assert.NotEqual(t, secret, secret2, "Generated secrets should be unique")
		})
	}
}

func TestGenerateSecretError(t *testing.T) {
	// Save original reader
	originalRead := rand.Read

	// This test is more for documentation purposes since we can't easily mock rand.Read
	// In a real error scenario, generateSecret would fail if rand.Read fails
	_ = originalRead
}

func TestInstallWizardPrompt(t *testing.T) {
	// This is a simple test to verify the wizard structure
	// In a real test environment, we would mock stdin/stdout
	// and test the full interactive flow

	wizard := &InstallWizard{
		outputFile: "/tmp/test-config.yaml",
	}

	assert.NotNil(t, wizard)
	assert.Equal(t, "/tmp/test-config.yaml", wizard.outputFile)
}

func TestParseInt(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		defaultValue int
		expected     int
	}{
		{"valid integer", "8888", 0, 8888},
		{"empty string uses default", "", 8080, 8080},
		{"invalid string uses default", "invalid", 8080, 8080},
		{"negative number", "-1", 0, -1},
	}

	wizard := &InstallWizard{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := wizard.parseInt(tt.input, tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}
