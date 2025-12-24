package types

// OrderItem 订单条目。
type OrderItem struct {
	ID             uint64         `json:"id"`
	OrderID        uint64         `json:"order_id"`
	ItemType       string         `json:"item_type"`
	ItemID         uint64         `json:"item_id"`
	Name           string         `json:"name"`
	Quantity       int            `json:"quantity"`
	UnitPriceCents int64          `json:"unit_price_cents"`
	Currency       string         `json:"currency"`
	SubtotalCents  int64          `json:"subtotal_cents"`
	Metadata       map[string]any `json:"metadata"`
	CreatedAt      int64          `json:"created_at"`
}

// OrderRefund 订单退款记录。
type OrderRefund struct {
	ID          uint64         `json:"id"`
	OrderID     uint64         `json:"order_id"`
	AmountCents int64          `json:"amount_cents"`
	Reason      string         `json:"reason"`
	Reference   string         `json:"reference"`
	Metadata    map[string]any `json:"metadata,omitempty"`
	CreatedAt   int64          `json:"created_at"`
}

// OrderDetail 订单详情。
type OrderDetail struct {
	ID                    uint64         `json:"id"`
	Number                string         `json:"number"`
	UserID                uint64         `json:"user_id"`
	Status                string         `json:"status"`
	PaymentStatus         string         `json:"payment_status"`
	TotalCents            int64          `json:"total_cents"`
	RefundedCents         int64          `json:"refunded_cents"`
	Currency              string         `json:"currency"`
	PaymentMethod         string         `json:"payment_method"`
	PaymentIntentID       string         `json:"payment_intent_id,omitempty"`
	PaymentReference      string         `json:"payment_reference,omitempty"`
	PaymentFailureCode    string         `json:"payment_failure_code,omitempty"`
	PaymentFailureMessage string         `json:"payment_failure_message,omitempty"`
	PlanID                *uint64        `json:"plan_id,omitempty"`
	PlanSnapshot          map[string]any `json:"plan_snapshot,omitempty"`
	Metadata              map[string]any `json:"metadata,omitempty"`
	PaidAt                *int64         `json:"paid_at,omitempty"`
	CancelledAt           *int64         `json:"cancelled_at,omitempty"`
	RefundedAt            *int64         `json:"refunded_at,omitempty"`
	CreatedAt             int64          `json:"created_at"`
	UpdatedAt             int64          `json:"updated_at"`
	Items                 []OrderItem    `json:"items"`
	Refunds               []OrderRefund  `json:"refunds,omitempty"`
	Payments              []OrderPayment `json:"payments,omitempty"`
}

// UserCreateOrderRequest 创建订单请求。
type UserCreateOrderRequest struct {
	PlanID           uint64 `json:"plan_id"`
	Quantity         int    `json:"quantity"`
	PaymentMethod    string `json:"payment_method,omitempty"`
	PaymentChannel   string `json:"payment_channel,omitempty"`
	PaymentReturnURL string `json:"payment_return_url,omitempty"`
	IdempotencyKey   string `json:"idempotency_key,omitempty"`
}

// UserOrderListRequest 用户订单列表查询参数。
type UserOrderListRequest struct {
	Page          int    `form:"page"`
	PerPage       int    `form:"per_page"`
	Status        string `form:"status"`
	PaymentMethod string `form:"payment_method"`
	PaymentStatus string `form:"payment_status"`
	Number        string `form:"number"`
	Sort          string `form:"sort"`
	Direction     string `form:"direction"`
}

// UserOrderListResponse 用户订单列表响应。
type UserOrderListResponse struct {
	Orders     []OrderDetail  `json:"orders"`
	Pagination PaginationMeta `json:"pagination"`
}

// UserOrderResponse 用户订单详情响应。
type UserOrderResponse struct {
	Order       OrderDetail                `json:"order"`
	Balance     BalanceSnapshot            `json:"balance"`
	Transaction *BalanceTransactionSummary `json:"transaction,omitempty"`
}

// UserGetOrderRequest 用户订单详情请求。
type UserGetOrderRequest struct {
	OrderID uint64 `path:"id"`
}

// UserCancelOrderRequest 用户主动取消订单。
type UserCancelOrderRequest struct {
	OrderID uint64 `path:"id"`
	Reason  string `json:"reason,omitempty"`
}

// AdminListOrdersRequest 管理端订单列表查询。
type AdminListOrdersRequest struct {
	Page          int    `form:"page"`
	PerPage       int    `form:"per_page"`
	Status        string `form:"status"`
	PaymentMethod string `form:"payment_method"`
	PaymentStatus string `form:"payment_status"`
	Number        string `form:"number"`
	Sort          string `form:"sort"`
	Direction     string `form:"direction"`
	UserID        uint64 `form:"user_id"`
}

// OrderPayment 外部支付流水。
type OrderPayment struct {
	ID             uint64         `json:"id"`
	OrderID        uint64         `json:"order_id"`
	Provider       string         `json:"provider"`
	Method         string         `json:"method"`
	IntentID       string         `json:"intent_id"`
	Reference      string         `json:"reference"`
	Status         string         `json:"status"`
	AmountCents    int64          `json:"amount_cents"`
	Currency       string         `json:"currency"`
	FailureCode    string         `json:"failure_code"`
	FailureMessage string         `json:"failure_message"`
	Metadata       map[string]any `json:"metadata"`
	CreatedAt      int64          `json:"created_at"`
	UpdatedAt      int64          `json:"updated_at"`
}

// OrderUserSummary 订单关联用户摘要。
type OrderUserSummary struct {
	ID          uint64 `json:"id"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
}

// AdminOrderDetail 管理端订单详情。
type AdminOrderDetail struct {
	OrderDetail
	User OrderUserSummary `json:"user"`
}

// AdminOrderListResponse 管理端订单列表响应。
type AdminOrderListResponse struct {
	Orders     []AdminOrderDetail `json:"orders"`
	Pagination PaginationMeta     `json:"pagination"`
}

// AdminGetOrderRequest 管理端订单详情请求。
type AdminGetOrderRequest struct {
	OrderID uint64 `path:"id"`
}

// AdminOrderResponse 管理端订单详情响应。
type AdminOrderResponse struct {
	Order AdminOrderDetail `json:"order"`
}

// AdminMarkOrderPaidRequest 管理端手动标记订单为已支付。
type AdminMarkOrderPaidRequest struct {
	OrderID       uint64 `path:"id"`
	PaymentMethod string `json:"payment_method,omitempty"`
	PaidAt        *int64 `json:"paid_at,omitempty"`
	Note          string `json:"note,omitempty"`
	Reference     string `json:"reference,omitempty"`
	ChargeBalance bool   `json:"charge_balance,omitempty"`
}

// AdminCancelOrderRequest 管理端取消订单。
type AdminCancelOrderRequest struct {
	OrderID     uint64 `path:"id"`
	Reason      string `json:"reason,omitempty"`
	CancelledAt *int64 `json:"cancelled_at,omitempty"`
}

// AdminRefundOrderRequest 管理端发起退款。
type AdminRefundOrderRequest struct {
	OrderID       uint64         `path:"id"`
	AmountCents   int64          `json:"amount_cents"`
	Reason        string         `json:"reason,omitempty"`
	Metadata      map[string]any `json:"metadata,omitempty"`
	RefundAt      *int64         `json:"refund_at,omitempty"`
	CreditBalance bool           `json:"credit_balance,omitempty"`
}

// AdminPaymentCallbackRequest 外部支付回调请求。
type AdminPaymentCallbackRequest struct {
	OrderID        uint64 `json:"order_id"`
	PaymentID      uint64 `json:"payment_id"`
	Status         string `json:"status"`
	Reference      string `json:"reference,omitempty"`
	FailureCode    string `json:"failure_code,omitempty"`
	FailureMessage string `json:"failure_message,omitempty"`
	PaidAt         *int64 `json:"paid_at,omitempty"`
}
