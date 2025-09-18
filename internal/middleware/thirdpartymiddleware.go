package middleware

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/zeromicro/go-zero/rest/httpx"

	"github.com/zero-net-panel/zero-net-panel/internal/repository"
)

const (
	headerAPIKey    = "X-ZNP-API-Key"
	headerSignature = "X-ZNP-Signature"
	headerTimestamp = "X-ZNP-Timestamp"
	headerNonce     = "X-ZNP-Nonce"
	headerEncrypted = "X-ZNP-Encrypted"
	headerIV        = "X-ZNP-IV"
)

// ThirdPartyMiddleware 对接第三方调用的安全校验。
type ThirdPartyMiddleware struct {
	repo          repository.SecurityRepository
	cacheDuration time.Duration

	mu       sync.RWMutex
	cached   repository.SecuritySetting
	cachedAt time.Time
}

// NewThirdPartyMiddleware 构造函数。
func NewThirdPartyMiddleware(repo repository.SecurityRepository) *ThirdPartyMiddleware {
	return &ThirdPartyMiddleware{repo: repo, cacheDuration: 30 * time.Second}
}

// Handler 返回中间件函数。
func (m *ThirdPartyMiddleware) Handler(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		setting, err := m.loadSetting(r.Context())
		if err != nil {
			httpx.WriteJsonCtx(r.Context(), w, http.StatusInternalServerError, map[string]any{
				"message": "failed to load security settings",
			})
			return
		}
		if !setting.ThirdPartyAPIEnabled || setting.APIKey == "" || setting.APISecret == "" {
			next(w, r)
			return
		}

		if err := m.verifyAndPrepareRequest(r.Context(), r, setting); err != nil {
			status := http.StatusUnauthorized
			if errors.Is(err, context.Canceled) {
				status = http.StatusRequestTimeout
			}
			httpx.WriteJsonCtx(r.Context(), w, status, map[string]any{
				"message": err.Error(),
			})
			return
		}

		next(w, r)
	}
}

func (m *ThirdPartyMiddleware) loadSetting(ctx context.Context) (repository.SecuritySetting, error) {
	m.mu.RLock()
	cached := m.cached
	cachedAt := m.cachedAt
	m.mu.RUnlock()

	if !cachedAt.IsZero() && time.Since(cachedAt) < m.cacheDuration {
		return cached, nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.cachedAt.IsZero() && time.Since(m.cachedAt) < m.cacheDuration {
		return m.cached, nil
	}

	setting, err := m.repo.GetThirdPartyAPIConfig(ctx)
	if err != nil {
		return repository.SecuritySetting{}, err
	}
	if setting.NonceTTLSeconds == 0 {
		setting.NonceTTLSeconds = 300
	}

	m.cached = setting
	m.cachedAt = time.Now()
	return setting, nil
}

func (m *ThirdPartyMiddleware) verifyAndPrepareRequest(ctx context.Context, r *http.Request, setting repository.SecuritySetting) error {
	apiKey := strings.TrimSpace(r.Header.Get(headerAPIKey))
	signature := strings.TrimSpace(r.Header.Get(headerSignature))
	timestampStr := strings.TrimSpace(r.Header.Get(headerTimestamp))
	nonce := strings.TrimSpace(r.Header.Get(headerNonce))

	if apiKey == "" || signature == "" || timestampStr == "" || nonce == "" {
		return errors.New("missing third-party signature headers")
	}
	if apiKey != setting.APIKey {
		return errors.New("invalid api key")
	}

	ts, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		return errors.New("invalid timestamp")
	}

	ttl := time.Duration(setting.NonceTTLSeconds) * time.Second
	if ttl <= 0 {
		ttl = 5 * time.Minute
	}
	timestamp := time.Unix(ts, 0)
	if diff := time.Since(timestamp); diff > ttl || diff < -ttl {
		return errors.New("timestamp out of allowed window")
	}

	bodyBytes, err := readBody(r)
	if err != nil {
		return err
	}

	canonical := buildCanonicalString(r.Method, r.URL.Path, r.URL.RawQuery, timestampStr, nonce, bodyBytes)
	expectedSignature := computeHMACSHA256(setting.APISecret, canonical)
	if !hmac.Equal([]byte(signature), []byte(expectedSignature)) {
		return errors.New("signature verification failed")
	}

	encrypted := strings.EqualFold(r.Header.Get(headerEncrypted), "true") || r.Header.Get(headerEncrypted) == "1"
	if encrypted {
		decrypted, err := decryptBody(setting.APISecret, r.Header.Get(headerIV), bodyBytes)
		if err != nil {
			return err
		}
		bodyBytes = decrypted
		r.Header.Del(headerEncrypted)
		r.Header.Del(headerIV)
	}

	r.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	r.ContentLength = int64(len(bodyBytes))
	if r.GetBody != nil {
		r.GetBody = func() (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewReader(bodyBytes)), nil
		}
	}

	return nil
}

func readBody(r *http.Request) ([]byte, error) {
	if r.Body == nil {
		return []byte{}, nil
	}
	data, err := io.ReadAll(r.Body)
	closeErr := r.Body.Close()
	if err != nil {
		return nil, err
	}
	if closeErr != nil {
		return nil, closeErr
	}
	return data, nil
}

func buildCanonicalString(method, path, rawQuery, timestamp, nonce string, body []byte) string {
	payload := base64.StdEncoding.EncodeToString(body)
	parts := []string{
		strings.ToUpper(method),
		path,
		rawQuery,
		timestamp,
		nonce,
		payload,
	}
	return strings.Join(parts, "\n")
}

func computeHMACSHA256(secret, data string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(data))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

func decryptBody(secret, ivHeader string, cipherBody []byte) ([]byte, error) {
	if ivHeader == "" {
		return nil, errors.New("missing encryption iv")
	}
	iv, err := base64.StdEncoding.DecodeString(ivHeader)
	if err != nil {
		return nil, errors.New("invalid iv")
	}
	if len(iv) == 0 {
		return nil, errors.New("invalid iv length")
	}
	key := sha256.Sum256([]byte(secret))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	if len(iv) != gcm.NonceSize() {
		return nil, errors.New("invalid iv size")
	}
	plaintext, err := gcm.Open(nil, iv, cipherBody, nil)
	if err != nil {
		return nil, errors.New("failed to decrypt body")
	}
	return plaintext, nil
}
