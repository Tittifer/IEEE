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
