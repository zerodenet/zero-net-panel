package kernel

import "time"

// NodeConfig 表示自研内核返回的节点配置摘要。
type NodeConfig struct {
	NodeID      string
	Protocol    string
	Endpoint    string
	Revision    string
	Payload     map[string]any
	RetrievedAt time.Time
}
