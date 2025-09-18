package repository

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Node 表示节点元信息。
type Node struct {
	ID           uint64    `gorm:"primaryKey"`
	Name         string    `gorm:"size:255;uniqueIndex"`
	Region       string    `gorm:"size:128"`
	Country      string    `gorm:"size:8"`
	ISP          string    `gorm:"size:128"`
	Status       string    `gorm:"size:32"`
	Tags         []string  `gorm:"serializer:json"`
	Protocols    []string  `gorm:"serializer:json"`
	CapacityMbps int       `gorm:"column:capacity_mbps"`
	Description  string    `gorm:"type:text"`
	LastSyncedAt time.Time `gorm:"column:last_synced_at"`
	UpdatedAt    time.Time
	CreatedAt    time.Time
}

// TableName 自定义节点表名。
func (Node) TableName() string { return "nodes" }

// NodeKernel 表示节点某一协议的配置摘要。
type NodeKernel struct {
	NodeID       uint64         `gorm:"primaryKey;autoIncrement:false"`
	Protocol     string         `gorm:"primaryKey;size:32"`
	Endpoint     string         `gorm:"size:512"`
	Revision     string         `gorm:"size:128"`
	Status       string         `gorm:"size:32"`
	Config       map[string]any `gorm:"serializer:json"`
	LastSyncedAt time.Time      `gorm:"column:last_synced_at"`
	UpdatedAt    time.Time
	CreatedAt    time.Time
}

// TableName 自定义节点内核表名。
func (NodeKernel) TableName() string { return "node_kernels" }

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
}

// NewNodeRepository 创建节点仓储。
func NewNodeRepository(db *gorm.DB) (NodeRepository, error) {
	if db == nil {
		return nil, errors.New("repository: database connection is required")
	}
	return &nodeRepository{db: db}, nil
}

func (r *nodeRepository) List(ctx context.Context, opts ListNodesOptions) ([]Node, int64, error) {
	if err := ctx.Err(); err != nil {
		return nil, 0, err
	}

	opts = normalizeListNodesOptions(opts)

	base := r.db.WithContext(ctx).Model(&Node{})

	if query := strings.TrimSpace(strings.ToLower(opts.Query)); query != "" {
		like := fmt.Sprintf("%%%s%%", query)
		base = base.Where("(LOWER(name) LIKE ? OR LOWER(region) LIKE ? OR LOWER(description) LIKE ?)", like, like, like)
	}
	if status := strings.TrimSpace(strings.ToLower(opts.Status)); status != "" {
		base = base.Where("LOWER(status) = ?", status)
	}
	if protocol := strings.TrimSpace(strings.ToLower(opts.Protocol)); protocol != "" {
		base = base.Joins("JOIN node_kernels nk ON nk.node_id = nodes.id AND LOWER(nk.protocol) = ?", protocol)
	}

	countQuery := base.Session(&gorm.Session{}).Distinct("nodes.id")
	var total int64
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []Node{}, 0, nil
	}

	orderClause := buildNodeOrderClause(opts.Sort, opts.Direction)
	offset := (opts.Page - 1) * opts.PerPage
	listQuery := base.Session(&gorm.Session{}).Distinct().Order(orderClause).Limit(opts.PerPage).Offset(offset)

	var nodes []Node
	if err := listQuery.Find(&nodes).Error; err != nil {
		return nil, 0, err
	}

	return nodes, total, nil
}

func (r *nodeRepository) Get(ctx context.Context, nodeID uint64) (Node, error) {
	if err := ctx.Err(); err != nil {
		return Node{}, err
	}

	var node Node
	if err := r.db.WithContext(ctx).First(&node, nodeID).Error; err != nil {
		return Node{}, translateError(err)
	}

	return node, nil
}

func (r *nodeRepository) GetKernels(ctx context.Context, nodeID uint64) ([]NodeKernel, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	var kernels []NodeKernel
	if err := r.db.WithContext(ctx).
		Where("node_id = ?", nodeID).
		Order("protocol ASC, last_synced_at DESC").
		Find(&kernels).Error; err != nil {
		return nil, err
	}

	return kernels, nil
}

func (r *nodeRepository) RecordKernelSync(ctx context.Context, nodeID uint64, kernel NodeKernel) (NodeKernel, error) {
	if err := ctx.Err(); err != nil {
		return NodeKernel{}, err
	}

	proto := normalizeProtocol(kernel.Protocol)
	now := time.Now().UTC()

	if kernel.LastSyncedAt.IsZero() {
		kernel.LastSyncedAt = now
	}
	if kernel.CreatedAt.IsZero() {
		kernel.CreatedAt = now
	}
	kernel.UpdatedAt = now
	kernel.Protocol = proto
	kernel.NodeID = nodeID
	if kernel.Status == "" {
		kernel.Status = "synced"
	}
	if kernel.Config == nil {
		kernel.Config = map[string]any{}
	}

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var node Node
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&node, nodeID).Error; err != nil {
			return err
		}

		var existing NodeKernel
		err := tx.Where("node_id = ? AND LOWER(protocol) = ?", nodeID, proto).First(&existing).Error
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			if err := tx.Create(&kernel).Error; err != nil {
				return err
			}
		case err != nil:
			return err
		default:
			existing.Endpoint = kernel.Endpoint
			existing.Revision = kernel.Revision
			existing.Status = kernel.Status
			existing.Config = kernel.Config
			existing.LastSyncedAt = kernel.LastSyncedAt
			existing.UpdatedAt = kernel.UpdatedAt
			if err := tx.Save(&existing).Error; err != nil {
				return err
			}
			kernel = existing
		}

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

		return tx.Save(&node).Error
	})

	if err != nil {
		return NodeKernel{}, translateError(err)
	}

	return kernel, nil
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

func buildNodeOrderClause(field, direction string) string {
	column := "nodes.updated_at"
	switch strings.ToLower(field) {
	case "name":
		column = "nodes.name"
	case "region":
		column = "nodes.region"
	case "last_synced_at":
		column = "nodes.last_synced_at"
	case "capacity_mbps":
		column = "nodes.capacity_mbps"
	}

	dir := "ASC"
	if strings.EqualFold(direction, "desc") {
		dir = "DESC"
	}

	return fmt.Sprintf("%s %s", column, dir)
}

func containsIgnoreCase(items []string, target string) bool {
	target = strings.ToLower(strings.TrimSpace(target))
	for _, item := range items {
		if strings.ToLower(item) == target {
			return true
		}
	}
	return false
}

func normalizeProtocol(protocol string) string {
	proto := strings.TrimSpace(strings.ToLower(protocol))
	if proto == "" {
		return "http"
	}
	return proto
}
