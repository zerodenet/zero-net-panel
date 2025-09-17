package repository

import (
	"context"
	"sort"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"
)

// Node 表示节点元信息。
type Node struct {
	ID           uint64
	Name         string
	Region       string
	Country      string
	ISP          string
	Status       string
	Tags         []string
	Protocols    []string
	CapacityMbps int
	Description  string
	LastSyncedAt time.Time
	UpdatedAt    time.Time
	CreatedAt    time.Time
}

// NodeKernel 表示节点某一协议的配置摘要。
type NodeKernel struct {
	NodeID       uint64
	Protocol     string
	Endpoint     string
	Revision     string
	Status       string
	Config       map[string]any
	LastSyncedAt time.Time
	UpdatedAt    time.Time
	CreatedAt    time.Time
}

// ListNodesOptions 为节点列表提供过滤与排序选项。
type ListNodesOptions struct {
	Page      int
	PerPage   int
	Sort      string
	Direction string
	Query     string
	Status    string
	Protocol  string
}

// NodeRepository 定义节点仓储接口。
type NodeRepository interface {
	List(ctx context.Context, opts ListNodesOptions) ([]Node, int64, error)
	Get(ctx context.Context, nodeID uint64) (Node, error)
	GetKernels(ctx context.Context, nodeID uint64) ([]NodeKernel, error)
	RecordKernelSync(ctx context.Context, nodeID uint64, kernel NodeKernel) (NodeKernel, error)
}

type nodeRepository struct {
	db *gorm.DB

	mu      sync.RWMutex
	nodes   map[uint64]*Node
	kernels map[uint64]map[string]*NodeKernel
	nextID  uint64
}

// NewNodeRepository 创建节点仓储，当前以内存数据为主，后续可接入 GORM。
func NewNodeRepository(db *gorm.DB) NodeRepository {
	repo := &nodeRepository{
		db:      db,
		nodes:   make(map[uint64]*Node),
		kernels: make(map[uint64]map[string]*NodeKernel),
		nextID:  1,
	}
	repo.seed()
	return repo
}

func (r *nodeRepository) seed() {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now().UTC()

	hkNode := &Node{
		Name:         "edge-hk-1",
		Region:       "Hong Kong",
		Country:      "HK",
		ISP:          "HKIX",
		Status:       "online",
		Tags:         []string{"premium", "asia"},
		CapacityMbps: 1000,
		Description:  "香港高带宽边缘节点示例",
		CreatedAt:    now.Add(-36 * time.Hour),
		UpdatedAt:    now.Add(-30 * time.Minute),
		LastSyncedAt: now.Add(-30 * time.Minute),
	}
	hkID := r.addNodeLocked(hkNode)

	r.setKernelLocked(hkID, &NodeKernel{
		NodeID:       hkID,
		Protocol:     "http",
		Endpoint:     "https://kernel-hk.example.com/api",
		Revision:     "rev-hk-http",
		Status:       "synced",
		Config:       map[string]any{"transport": "ws", "heartbeat": 30},
		LastSyncedAt: hkNode.LastSyncedAt,
		CreatedAt:    hkNode.CreatedAt,
	})
	r.setKernelLocked(hkID, &NodeKernel{
		NodeID:       hkID,
		Protocol:     "grpc",
		Endpoint:     "kernel-hk.example.com:9000",
		Revision:     "rev-hk-grpc",
		Status:       "synced",
		Config:       map[string]any{"transport": "grpc", "heartbeat": 15},
		LastSyncedAt: hkNode.LastSyncedAt.Add(-5 * time.Minute),
		CreatedAt:    hkNode.CreatedAt,
	})

	laNode := &Node{
		Name:         "edge-la-1",
		Region:       "Los Angeles",
		Country:      "US",
		ISP:          "NTT",
		Status:       "maintenance",
		Tags:         []string{"standard", "america"},
		CapacityMbps: 600,
		Description:  "北美标准线路示例节点",
		CreatedAt:    now.Add(-72 * time.Hour),
		UpdatedAt:    now.Add(-6 * time.Hour),
		LastSyncedAt: now.Add(-12 * time.Hour),
	}
	laID := r.addNodeLocked(laNode)
	r.setKernelLocked(laID, &NodeKernel{
		NodeID:       laID,
		Protocol:     "http",
		Endpoint:     "https://kernel-la.example.com/api",
		Revision:     "rev-la-http",
		Status:       "synced",
		Config:       map[string]any{"transport": "http", "heartbeat": 45},
		LastSyncedAt: laNode.LastSyncedAt,
		CreatedAt:    laNode.CreatedAt,
	})
}

func (r *nodeRepository) List(ctx context.Context, opts ListNodesOptions) ([]Node, int64, error) {
	if err := ctx.Err(); err != nil {
		return nil, 0, err
	}

	opts = normalizeListNodesOptions(opts)

	r.mu.RLock()
	defer r.mu.RUnlock()

	query := strings.TrimSpace(strings.ToLower(opts.Query))
	status := strings.TrimSpace(strings.ToLower(opts.Status))
	protocol := strings.TrimSpace(strings.ToLower(opts.Protocol))

	items := make([]Node, 0, len(r.nodes))
	for _, node := range r.nodes {
		if query != "" {
			if !strings.Contains(strings.ToLower(node.Name), query) &&
				!strings.Contains(strings.ToLower(node.Region), query) &&
				!strings.Contains(strings.ToLower(node.Description), query) {
				continue
			}
		}
		if status != "" && strings.ToLower(node.Status) != status {
			continue
		}
		if protocol != "" && !containsIgnoreCase(node.Protocols, protocol) {
			continue
		}

		items = append(items, cloneNode(node))
	}

	sortField := opts.Sort
	desc := strings.EqualFold(opts.Direction, "desc")
	sort.SliceStable(items, func(i, j int) bool {
		if desc {
			return nodeLess(items[j], items[i], sortField)
		}
		return nodeLess(items[i], items[j], sortField)
	})

	total := int64(len(items))
	start := (opts.Page - 1) * opts.PerPage
	if start >= len(items) {
		return []Node{}, total, nil
	}

	end := start + opts.PerPage
	if end > len(items) {
		end = len(items)
	}

	result := make([]Node, end-start)
	copy(result, items[start:end])
	return result, total, nil
}

func (r *nodeRepository) Get(ctx context.Context, nodeID uint64) (Node, error) {
	if err := ctx.Err(); err != nil {
		return Node{}, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	node, ok := r.nodes[nodeID]
	if !ok {
		return Node{}, ErrNotFound
	}

	return cloneNode(node), nil
}

func (r *nodeRepository) GetKernels(ctx context.Context, nodeID uint64) ([]NodeKernel, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	kernels, ok := r.kernels[nodeID]
	if !ok {
		return []NodeKernel{}, nil
	}

	items := make([]NodeKernel, 0, len(kernels))
	for _, kernel := range kernels {
		items = append(items, cloneNodeKernel(kernel))
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].Protocol == items[j].Protocol {
			return items[i].LastSyncedAt.After(items[j].LastSyncedAt)
		}
		return items[i].Protocol < items[j].Protocol
	})

	return items, nil
}

func (r *nodeRepository) RecordKernelSync(ctx context.Context, nodeID uint64, kernel NodeKernel) (NodeKernel, error) {
	if err := ctx.Err(); err != nil {
		return NodeKernel{}, err
	}

	proto := normalizeProtocol(kernel.Protocol)

	r.mu.Lock()
	defer r.mu.Unlock()

	node, ok := r.nodes[nodeID]
	if !ok {
		return NodeKernel{}, ErrNotFound
	}

	now := time.Now().UTC()

	kernel.NodeID = nodeID
	kernel.Protocol = proto
	if kernel.LastSyncedAt.IsZero() {
		kernel.LastSyncedAt = now
	}
	if kernel.CreatedAt.IsZero() {
		kernel.CreatedAt = now
	}
	kernel.UpdatedAt = now
	if kernel.Status == "" {
		kernel.Status = "synced"
	}
	if kernel.Config == nil {
		kernel.Config = map[string]any{}
	}

	if r.kernels[nodeID] == nil {
		r.kernels[nodeID] = make(map[string]*NodeKernel)
	}

	copied := cloneNodeKernel(&kernel)
	r.kernels[nodeID][proto] = &copied

	if !containsIgnoreCase(node.Protocols, proto) {
		node.Protocols = append(node.Protocols, proto)
		sort.Strings(node.Protocols)
	}

	if kernel.LastSyncedAt.After(node.LastSyncedAt) {
		node.LastSyncedAt = kernel.LastSyncedAt
	}
	if kernel.UpdatedAt.After(node.UpdatedAt) {
		node.UpdatedAt = kernel.UpdatedAt
	}
	node.Status = "online"

	return cloneNodeKernel(&copied), nil
}

func (r *nodeRepository) addNodeLocked(node *Node) uint64 {
	id := r.nextID
	r.nextID++

	copied := cloneNode(node)
	copied.ID = id
	if copied.CreatedAt.IsZero() {
		copied.CreatedAt = time.Now().UTC()
	}
	if copied.UpdatedAt.IsZero() {
		copied.UpdatedAt = copied.CreatedAt
	}
	if copied.Status == "" {
		copied.Status = "online"
	}

	r.nodes[id] = &copied
	return id
}

func (r *nodeRepository) setKernelLocked(nodeID uint64, kernel *NodeKernel) {
	if r.kernels[nodeID] == nil {
		r.kernels[nodeID] = make(map[string]*NodeKernel)
	}

	proto := normalizeProtocol(kernel.Protocol)
	copied := cloneNodeKernel(kernel)
	copied.NodeID = nodeID
	copied.Protocol = proto
	if copied.CreatedAt.IsZero() {
		copied.CreatedAt = time.Now().UTC()
	}
	if copied.LastSyncedAt.IsZero() {
		copied.LastSyncedAt = copied.CreatedAt
	}
	copied.UpdatedAt = copied.LastSyncedAt
	if copied.Status == "" {
		copied.Status = "synced"
	}

	r.kernels[nodeID][proto] = &copied

	if node, ok := r.nodes[nodeID]; ok {
		if !containsIgnoreCase(node.Protocols, proto) {
			node.Protocols = append(node.Protocols, proto)
			sort.Strings(node.Protocols)
		}
		if copied.LastSyncedAt.After(node.LastSyncedAt) {
			node.LastSyncedAt = copied.LastSyncedAt
		}
		if copied.UpdatedAt.After(node.UpdatedAt) {
			node.UpdatedAt = copied.UpdatedAt
		}
	}
}

func normalizeListNodesOptions(opts ListNodesOptions) ListNodesOptions {
	if opts.Page <= 0 {
		opts.Page = 1
	}
	if opts.PerPage <= 0 {
		opts.PerPage = 20
	}
	if opts.PerPage > 100 {
		opts.PerPage = 100
	}
	if opts.Sort == "" {
		opts.Sort = "updated_at"
	}
	opts.Sort = strings.ToLower(opts.Sort)
	if opts.Direction == "" {
		opts.Direction = "desc"
	}
	return opts
}

func nodeLess(a, b Node, field string) bool {
	switch field {
	case "name":
		if strings.EqualFold(a.Name, b.Name) {
			return a.ID < b.ID
		}
		return strings.ToLower(a.Name) < strings.ToLower(b.Name)
	case "region":
		if strings.EqualFold(a.Region, b.Region) {
			return a.ID < b.ID
		}
		return strings.ToLower(a.Region) < strings.ToLower(b.Region)
	case "last_synced_at":
		if a.LastSyncedAt.Equal(b.LastSyncedAt) {
			return a.ID < b.ID
		}
		return a.LastSyncedAt.Before(b.LastSyncedAt)
	case "created_at":
		if a.CreatedAt.Equal(b.CreatedAt) {
			return a.ID < b.ID
		}
		return a.CreatedAt.Before(b.CreatedAt)
	case "status":
		if strings.EqualFold(a.Status, b.Status) {
			return a.ID < b.ID
		}
		return strings.ToLower(a.Status) < strings.ToLower(b.Status)
	default: // updated_at
		if a.UpdatedAt.Equal(b.UpdatedAt) {
			return a.ID < b.ID
		}
		return a.UpdatedAt.Before(b.UpdatedAt)
	}
}

func cloneNode(node *Node) Node {
	copied := *node
	if node.Tags != nil {
		copied.Tags = append([]string(nil), node.Tags...)
	}
	if node.Protocols != nil {
		copied.Protocols = append([]string(nil), node.Protocols...)
	}
	return copied
}

func cloneNodeKernel(kernel *NodeKernel) NodeKernel {
	copied := *kernel
	if kernel.Config != nil {
		copied.Config = cloneAnyMap(kernel.Config)
	}
	return copied
}

func cloneAnyMap(src map[string]any) map[string]any {
	if src == nil {
		return nil
	}
	dst := make(map[string]any, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

func containsIgnoreCase(values []string, target string) bool {
	for _, val := range values {
		if strings.EqualFold(val, target) {
			return true
		}
	}
	return false
}

func normalizeProtocol(protocol string) string {
	if protocol == "" {
		return "http"
	}
	return strings.ToLower(protocol)
}
