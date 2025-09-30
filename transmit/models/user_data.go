package models

// UserInfo 用户信息
type UserInfo struct {
	DID       string `json:"did"`       // 用户的分布式身份标识符
	Name      string `json:"name"`      // 用户姓名
	Status    string `json:"status"`    // 用户状态
	RiskScore int    `json:"riskScore"` // 风险评分
}

// DIDRiskRecord DID风险记录
type DIDRiskRecord struct {
	DID       string `json:"did"`       // 用户的分布式身份标识符
	RiskScore int    `json:"riskScore"` // 风险评分
	Timestamp int64  `json:"timestamp"` // 时间戳
}

// UserSession 用户会话
type UserSession struct {
	DID            string `json:"did"`            // 用户的分布式身份标识符
	Status         string `json:"status"`         // 会话状态：online/offline
	SeverityLevel  int    `json:"severityLevel"`  // 当前会话中的风险行为严重程度累计值s
	Timestamp      int64  `json:"timestamp"`      // 最后更新时间戳
	HasTriggeredRisk bool  `json:"hasTriggeredRisk"` // 当前会话中是否已触发过风险行为
}

// 常量定义
const (
	RiskScoreThreshold = 50 // 风险评分阈值
)