package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/rest/httpx"

	"github.com/zero-net-panel/zero-net-panel/internal/config"
)

const (
	headerWebhookToken    = "X-ZNP-Webhook-Token"
	headerStripeSignature = "Stripe-Signature"
)

// WebhookMiddleware validates webhook callbacks with optional IP allowlist and signatures.
type WebhookMiddleware struct {
	allowedNets     []*net.IPNet
	sharedToken     string
	stripeSecret    string
	stripeTolerance time.Duration
}

// NewWebhookMiddleware builds middleware from config.
func NewWebhookMiddleware(cfg config.WebhookConfig) *WebhookMiddleware {
	var nets []*net.IPNet
	for _, cidr := range cfg.AllowCIDRs {
		_, network, err := net.ParseCIDR(strings.TrimSpace(cidr))
		if err == nil && network != nil {
			nets = append(nets, network)
		}
	}

	return &WebhookMiddleware{
		allowedNets:     nets,
		sharedToken:     strings.TrimSpace(cfg.SharedToken),
		stripeSecret:    strings.TrimSpace(cfg.Stripe.SigningSecret),
		stripeTolerance: time.Duration(cfg.Stripe.ToleranceSeconds) * time.Second,
	}
}

// Handler returns the http handler middleware.
func (m *WebhookMiddleware) Handler(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if len(m.allowedNets) > 0 && !ipAllowed(m.allowedNets, clientIP(r)) {
			httpx.WriteJsonCtx(r.Context(), w, http.StatusForbidden, map[string]any{
				"message": "webhook access denied",
			})
			return
		}

		body, err := readWebhookBody(r)
		if err != nil {
			httpx.WriteJsonCtx(r.Context(), w, http.StatusBadRequest, map[string]any{
				"message": "invalid webhook payload",
			})
			return
		}

		if err := m.verify(body, r); err != nil {
			httpx.WriteJsonCtx(r.Context(), w, http.StatusUnauthorized, map[string]any{
				"message": err.Error(),
			})
			return
		}

		r.Body = io.NopCloser(bytes.NewReader(body))
		r.ContentLength = int64(len(body))
		if r.GetBody != nil {
			r.GetBody = func() (io.ReadCloser, error) {
				return io.NopCloser(bytes.NewReader(body)), nil
			}
		}

		next(w, r)
	}
}

func (m *WebhookMiddleware) verify(body []byte, r *http.Request) error {
	if m.stripeSecret != "" {
		if header := strings.TrimSpace(r.Header.Get(headerStripeSignature)); header != "" {
			return verifyStripeSignature(m.stripeSecret, header, body, m.stripeTolerance)
		}
		if m.sharedToken == "" {
			return errors.New("missing stripe signature")
		}
	}

	if m.sharedToken != "" {
		token := strings.TrimSpace(r.Header.Get(headerWebhookToken))
		if token == "" {
			return errors.New("missing webhook token")
		}
		if subtle.ConstantTimeCompare([]byte(token), []byte(m.sharedToken)) != 1 {
			return errors.New("invalid webhook token")
		}
	}

	return nil
}

func verifyStripeSignature(secret, header string, payload []byte, tolerance time.Duration) error {
	timestamp, signatures, err := parseStripeSignature(header)
	if err != nil {
		return err
	}

	if tolerance > 0 {
		ts := time.Unix(timestamp, 0)
		diff := time.Since(ts)
		if diff > tolerance || diff < -tolerance {
			return errors.New("stripe signature timestamp outside tolerance")
		}
	}

	signedPayload := strconv.FormatInt(timestamp, 10) + "." + string(payload)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(signedPayload))
	expected := hex.EncodeToString(mac.Sum(nil))

	for _, sig := range signatures {
		if hmac.Equal([]byte(sig), []byte(expected)) {
			return nil
		}
	}
	return errors.New("stripe signature mismatch")
}

func parseStripeSignature(header string) (int64, []string, error) {
	var (
		timestamp  int64
		signatures []string
	)

	parts := strings.Split(header, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			continue
		}
		switch kv[0] {
		case "t":
			ts, err := strconv.ParseInt(kv[1], 10, 64)
			if err == nil {
				timestamp = ts
			}
		case "v1":
			signatures = append(signatures, kv[1])
		}
	}

	if timestamp == 0 || len(signatures) == 0 {
		return 0, nil, errors.New("invalid stripe signature header")
	}
	return timestamp, signatures, nil
}

func ipAllowed(networks []*net.IPNet, ip net.IP) bool {
	if ip == nil {
		return false
	}
	for _, network := range networks {
		if network.Contains(ip) {
			return true
		}
	}
	return false
}

func readWebhookBody(r *http.Request) ([]byte, error) {
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
