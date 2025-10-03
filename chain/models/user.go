package models

// UserInfo 用户信息结构体
type UserInfo struct {
	DID           string `json:"did"`           // 用户的分布式身份标识符
	Name          string `json:"name"`          // 用户姓名
	RiskScore     int    `json:"riskScore"`     // 用户风险评分 (0-100)
	Status        string `json:"status"`        // 用户状态: active
	CreatedAt     string `json:"createdAt"`     // 创建时间
	LastUpdatedAt string `json:"lastUpdatedAt"` // 最后更新时间
}

// UserEvent 用户事件结构体，用于链码事件
type UserEvent struct {
	EventType string `json:"eventType"` // 事件类型
	DID       string `json:"did"`       // 用户DID
	Name      string `json:"name"`      // 用户名称
	Timestamp int64  `json:"timestamp"` // 事件时间戳
	RiskScore int    `json:"riskScore"` // 风险评分（可选，仅在登出时使用）
}

// 事件类型常量
const (
	EventTypeRegister   = "register"    // 用户注册事件
	EventTypeLogin      = "login"       // 用户登录事件
	EventTypeLogout     = "logout"      // 用户登出事件
	EventTypeRiskUpdate = "risk_update" // 风险评分更新事件
)

// 用户状态常量
const (
	StatusActive   = "active"   // 用户活跃状态
	StatusInactive = "inactive" // 用户非活跃状态
	StatusRisky    = "risky"    // 用户风险状态
	StatusOnline   = "online"   // 用户在线状态
	StatusOffline  = "offline"  // 用户离线状态
)

// 风险评分相关常量
const (
	InitialRiskScore   = 0  // 初始风险评分
	RiskScoreThreshold = 50 // 风险评分阈值，超过此值将禁止用户登录
)
