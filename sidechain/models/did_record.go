package models

// DIDRiskRecord 表示DID与风险评分和时间戳的映射
type DIDRiskRecord struct {
	DID       string `json:"did"`       // 用户的分布式身份标识符
	RiskScore int    `json:"riskScore"` // 风险评分
	Timestamp int64  `json:"timestamp"` // 时间戳
}

// 风险评分相关常量
const (
	InitialRiskScore   = 0  // 初始风险评分
	InitialTimestamp   = 0  // 初始时间戳
	RiskScoreThreshold = 50 // 风险评分阈值
)
