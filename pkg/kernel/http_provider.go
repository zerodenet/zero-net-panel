package kernel

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// HTTPDiscoveryProvider 实现基于 HTTP 的节点配置发现。
type HTTPDiscoveryProvider struct {
	baseURL string
	token   string
	client  *http.Client
}

// NewHTTPProvider 创建 HTTP Provider。
func NewHTTPProvider(opts HTTPOptions) (*HTTPDiscoveryProvider, error) {
	if opts.BaseURL == "" {
		return nil, fmt.Errorf("kernel http provider: base url required")
	}

	timeout := opts.Timeout
	if timeout <= 0 {
		timeout = 5 * time.Second
	}

	return &HTTPDiscoveryProvider{
		baseURL: strings.TrimSuffix(opts.BaseURL, "/"),
		token:   opts.Token,
		client: &http.Client{
			Timeout: timeout,
		},
	}, nil
}

// Name 返回 Provider 名称。
func (p *HTTPDiscoveryProvider) Name() string {
	return "http"
}

// FetchNodeConfig 从 HTTP 服务拉取节点配置。
func (p *HTTPDiscoveryProvider) FetchNodeConfig(ctx context.Context, nodeID string) (NodeConfig, error) {
	endpoint := fmt.Sprintf("%s/nodes/%s/config", p.baseURL, url.PathEscape(nodeID))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return NodeConfig{}, err
	}

	if p.token != "" {
		req.Header.Set("Authorization", "Bearer "+p.token)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return NodeConfig{}, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		// continue
	case http.StatusNotFound:
		return NodeConfig{}, ErrNotFound
	default:
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return NodeConfig{}, fmt.Errorf("kernel http provider: %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}

	var payload struct {
		NodeID   string         `json:"node_id"`
		Protocol string         `json:"protocol"`
		Endpoint string         `json:"endpoint"`
		Revision string         `json:"revision"`
		Payload  map[string]any `json:"payload"`
		Metadata map[string]any `json:"metadata"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return NodeConfig{}, fmt.Errorf("kernel http provider: decode response: %w", err)
	}

	revision := payload.Revision
	if revision == "" {
		if etag := strings.Trim(resp.Header.Get("Etag"), "\""); etag != "" {
			revision = etag
		}
	}

	if revision == "" && payload.Payload != nil {
		raw, err := json.Marshal(payload.Payload)
		if err == nil {
			hash := sha256.Sum256(raw)
			revision = hex.EncodeToString(hash[:])
		}
	}

	if revision == "" && payload.Metadata != nil {
		raw, err := json.Marshal(payload.Metadata)
		if err == nil {
			hash := sha256.Sum256(raw)
			revision = hex.EncodeToString(hash[:])
		}
	}

	result := NodeConfig{
		NodeID:      payload.NodeID,
		Protocol:    payload.Protocol,
		Endpoint:    payload.Endpoint,
		Revision:    revision,
		Payload:     payload.Payload,
		RetrievedAt: time.Now().UTC(),
	}

	if result.NodeID == "" {
		result.NodeID = nodeID
	}
	if result.Protocol == "" {
		result.Protocol = "http"
	}

	return result, nil
}

// Close 实现接口，无额外资源需释放。
func (p *HTTPDiscoveryProvider) Close() error {
	return nil
}
