package types

// PingResponse 保留健康检查响应。
type PingResponse struct {
	Status    string `json:"status"`
	Service   string `json:"service"`
	Version   string `json:"version"`
	Timestamp int64  `json:"timestamp"`
}

// PaginationMeta 统一 GitHub 风格分页返回。
type PaginationMeta struct {
        Page       int   `json:"page"`
        PerPage    int   `json:"per_page"`
        TotalCount int64 `json:"total_count"`
        HasNext    bool  `json:"has_next"`
        HasPrev    bool  `json:"has_prev"`
}

// AuthLoginRequest 登录请求。
type AuthLoginRequest struct {
        Email    string `json:"email"`
        Password string `json:"password"`
}

// AuthRefreshRequest 刷新令牌请求。
type AuthRefreshRequest struct {
        RefreshToken string `json:"refresh_token"`
}

// AuthenticatedUser 鉴权用户信息。
type AuthenticatedUser struct {
        ID          uint64   `json:"id"`
        Email       string   `json:"email"`
        DisplayName string   `json:"display_name"`
        Roles       []string `json:"roles"`
        CreatedAt   int64    `json:"created_at"`
        UpdatedAt   int64    `json:"updated_at"`
}

// AuthLoginResponse 登录响应。
type AuthLoginResponse struct {
        AccessToken       string            `json:"access_token"`
        RefreshToken      string            `json:"refresh_token"`
        TokenType         string            `json:"token_type"`
        ExpiresIn         int64             `json:"expires_in"`
        RefreshExpiresIn  int64             `json:"refresh_expires_in"`
        User              AuthenticatedUser `json:"user"`
}

// AuthRefreshResponse 刷新响应。
type AuthRefreshResponse struct {
        AccessToken       string            `json:"access_token"`
        RefreshToken      string            `json:"refresh_token"`
        TokenType         string            `json:"token_type"`
        ExpiresIn         int64             `json:"expires_in"`
        RefreshExpiresIn  int64             `json:"refresh_expires_in"`
        User              AuthenticatedUser `json:"user"`
}

// AdminListNodesRequest 管理端节点列表查询参数。
type AdminListNodesRequest struct {
        Page      int    `form:"page"`
        PerPage   int    `form:"per_page"`
        Sort      string `form:"sort"`
        Direction string `form:"direction"`
        Query     string `form:"q"`
        Status    string `form:"status"`
        Protocol  string `form:"protocol"`
}

// NodeSummary 节点摘要信息。
type NodeSummary struct {
	ID           uint64   `json:"id"`
	Name         string   `json:"name"`
	Region       string   `json:"region"`
	Country      string   `json:"country"`
	ISP          string   `json:"isp"`
	Status       string   `json:"status"`
	Tags         []string `json:"tags"`
	Protocols    []string `json:"protocols"`
	CapacityMbps int      `json:"capacity_mbps"`
	Description  string   `json:"description"`
	LastSyncedAt int64    `json:"last_synced_at"`
	UpdatedAt    int64    `json:"updated_at"`
}

// AdminNodeListResponse 节点列表响应。
type AdminNodeListResponse struct {
	Nodes      []NodeSummary  `json:"nodes"`
	Pagination PaginationMeta `json:"pagination"`
}

// AdminNodeKernelsRequest 请求节点协议配置。
type AdminNodeKernelsRequest struct {
	NodeID uint64 `path:"id"`
}

// NodeKernelSummary 节点协议详情。
type NodeKernelSummary struct {
	Protocol     string         `json:"protocol"`
	Endpoint     string         `json:"endpoint"`
	Revision     string         `json:"revision"`
	Status       string         `json:"status"`
	Config       map[string]any `json:"config"`
	LastSyncedAt int64          `json:"last_synced_at"`
}

// AdminNodeKernelResponse 节点协议列表返回。
type AdminNodeKernelResponse struct {
	NodeID  uint64              `json:"node_id"`
	Kernels []NodeKernelSummary `json:"kernels"`
}

// AdminSyncNodeKernelRequest 触发节点同步请求。
type AdminSyncNodeKernelRequest struct {
	NodeID   uint64 `path:"id"`
	Protocol string `json:"protocol"`
}

// AdminSyncNodeKernelResponse 返回最新同步信息。
type AdminSyncNodeKernelResponse struct {
	NodeID   uint64 `json:"node_id"`
	Protocol string `json:"protocol"`
	Revision string `json:"revision"`
	SyncedAt int64  `json:"synced_at"`
	Message  string `json:"message"`
}

// AdminListSubscriptionTemplatesRequest 管理端模板列表查询。
type AdminListSubscriptionTemplatesRequest struct {
	Page          int    `form:"page"`
	PerPage       int    `form:"per_page"`
	Sort          string `form:"sort"`
	Direction     string `form:"direction"`
	Query         string `form:"q"`
	ClientType    string `form:"client_type"`
	Format        string `form:"format"`
	IncludeDrafts bool   `form:"include_drafts"`
}

// TemplateVariable 模板变量描述。
type TemplateVariable struct {
	Type        string `json:"type"`
	Required    bool   `json:"required"`
	Description string `json:"description"`
	Default     any    `json:"default"`
}

// SubscriptionTemplateSummary 模板摘要信息。
type SubscriptionTemplateSummary struct {
	ID              uint64                      `json:"id"`
	Name            string                      `json:"name"`
	Description     string                      `json:"description"`
	ClientType      string                      `json:"client_type"`
	Format          string                      `json:"format"`
	Content         string                      `json:"content,omitempty"`
	Variables       map[string]TemplateVariable `json:"variables"`
	IsDefault       bool                        `json:"is_default"`
	Version         uint32                      `json:"version"`
	UpdatedAt       int64                       `json:"updated_at"`
	PublishedAt     int64                       `json:"published_at"`
	LastPublishedBy string                      `json:"last_published_by"`
}

// AdminSubscriptionTemplateListResponse 模板列表。
type AdminSubscriptionTemplateListResponse struct {
	Templates  []SubscriptionTemplateSummary `json:"templates"`
	Pagination PaginationMeta                `json:"pagination"`
}

// AdminCreateSubscriptionTemplateRequest 创建模板。
type AdminCreateSubscriptionTemplateRequest struct {
	Name        string                      `json:"name"`
	Description string                      `json:"description"`
	ClientType  string                      `json:"client_type"`
	Format      string                      `json:"format"`
	Content     string                      `json:"content"`
	Variables   map[string]TemplateVariable `json:"variables"`
	IsDefault   bool                        `json:"is_default"`
}

// AdminUpdateSubscriptionTemplateRequest 更新模板。
type AdminUpdateSubscriptionTemplateRequest struct {
	TemplateID  uint64                      `path:"id"`
	Name        *string                     `json:"name"`
	Description *string                     `json:"description"`
	Format      *string                     `json:"format"`
	Content     *string                     `json:"content"`
	Variables   map[string]TemplateVariable `json:"variables"`
	IsDefault   *bool                       `json:"is_default"`
}

// AdminPublishSubscriptionTemplateRequest 发布模板。
type AdminPublishSubscriptionTemplateRequest struct {
	TemplateID uint64 `path:"id"`
	Changelog  string `json:"changelog"`
	Operator   string `json:"operator"`
}

// AdminPublishSubscriptionTemplateResponse 发布结果。
type AdminPublishSubscriptionTemplateResponse struct {
	Template SubscriptionTemplateSummary      `json:"template"`
	History  SubscriptionTemplateHistoryEntry `json:"history"`
}

// SubscriptionTemplateHistoryEntry 模板历史条目。
type SubscriptionTemplateHistoryEntry struct {
	Version     uint32                      `json:"version"`
	Changelog   string                      `json:"changelog"`
	PublishedAt int64                       `json:"published_at"`
	PublishedBy string                      `json:"published_by"`
	Variables   map[string]TemplateVariable `json:"variables"`
}

// AdminSubscriptionTemplateHistoryRequest 查询历史。
type AdminSubscriptionTemplateHistoryRequest struct {
	TemplateID uint64 `path:"id"`
}

// AdminSubscriptionTemplateHistoryResponse 历史列表。
type AdminSubscriptionTemplateHistoryResponse struct {
	TemplateID uint64                             `json:"template_id"`
	History    []SubscriptionTemplateHistoryEntry `json:"history"`
}

// UserListSubscriptionsRequest 用户订阅列表查询。
type UserListSubscriptionsRequest struct {
	Page      int    `form:"page"`
	PerPage   int    `form:"per_page"`
	Sort      string `form:"sort"`
	Direction string `form:"direction"`
	Query     string `form:"q"`
	Status    string `form:"status"`
}

// UserSubscriptionSummary 用户订阅摘要。
type UserSubscriptionSummary struct {
	ID                   uint64   `json:"id"`
	Name                 string   `json:"name"`
	PlanName             string   `json:"plan_name"`
	Status               string   `json:"status"`
	TemplateID           uint64   `json:"template_id"`
	AvailableTemplateIDs []uint64 `json:"available_template_ids"`
	ExpiresAt            int64    `json:"expires_at"`
	TrafficTotalBytes    int64    `json:"traffic_total_bytes"`
	TrafficUsedBytes     int64    `json:"traffic_used_bytes"`
	DevicesLimit         int      `json:"devices_limit"`
	LastRefreshedAt      int64    `json:"last_refreshed_at"`
}

// UserSubscriptionListResponse 用户订阅列表。
type UserSubscriptionListResponse struct {
	Subscriptions []UserSubscriptionSummary `json:"subscriptions"`
	Pagination    PaginationMeta            `json:"pagination"`
}

// UserSubscriptionPreviewRequest 用户订阅预览请求。
type UserSubscriptionPreviewRequest struct {
	SubscriptionID uint64 `path:"id"`
	TemplateID     uint64 `form:"template_id"`
}

// UserSubscriptionPreviewResponse 用户订阅预览内容。
type UserSubscriptionPreviewResponse struct {
	SubscriptionID uint64 `json:"subscription_id"`
	TemplateID     uint64 `json:"template_id"`
	Content        string `json:"content"`
	ContentType    string `json:"content_type"`
	ETag           string `json:"etag"`
	GeneratedAt    int64  `json:"generated_at"`
}

// UserUpdateSubscriptionTemplateRequest 用户更新订阅模板。
type UserUpdateSubscriptionTemplateRequest struct {
	SubscriptionID uint64 `path:"id"`
	TemplateID     uint64 `json:"template_id"`
}

// UserUpdateSubscriptionTemplateResponse 更新结果。
type UserUpdateSubscriptionTemplateResponse struct {
	SubscriptionID uint64 `json:"subscription_id"`
	TemplateID     uint64 `json:"template_id"`
	UpdatedAt      int64  `json:"updated_at"`
}
