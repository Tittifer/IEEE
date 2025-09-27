package models

import (
	"time"
)

// UserInfo 用户信息结构体
type UserInfo struct {
	DID           string    `json:"did"`           // 用户的分布式身份标识符
	Name          string    `json:"name"`          // 用户姓名
	IDNumber      string    `json:"idNumber"`      // 身份证号
	PublicKey     string    `json:"publicKey"`     // 用户公钥
	RiskScore     int       `json:"riskScore"`     // 用户风险评分 (0-100, 0表示无风险，100表示最高风险)
	AccessLevel   int       `json:"accessLevel"`   // 用户访问级别 (1-5, 1表示最高权限，5表示最低权限)
	Status        string    `json:"status"`        // 用户状态: active, suspended, blocked
	AttackHistory []Attack  `json:"attackHistory"` // 攻击历史记录
	CreatedAt     time.Time `json:"createdAt"`     // 创建时间
	LastUpdatedAt time.Time `json:"lastUpdatedAt"` // 最后更新时间
}

// Attack 攻击记录结构体
type Attack struct {
	Timestamp   time.Time `json:"timestamp"`   // 攻击时间
	HoneypotID  string    `json:"honeypotId"`  // 蜜点ID
	AttackType  string    `json:"attackType"`  // 攻击类型
	Description string    `json:"description"` // 攻击描述
	Severity    int       `json:"severity"`    // 攻击严重程度 (1-10)
}

// 用户状态常量
const (
	StatusActive    = "active"
	StatusSuspended = "suspended"
	StatusBlocked   = "blocked"
)

// 风险评分阈值常量
const (
	RiskThresholdLow     = 20  // 低风险阈值
	RiskThresholdMedium  = 40  // 中等风险阈值
	RiskThresholdHigh    = 60  // 高风险阈值
	RiskThresholdCritical = 80 // 严重风险阈值
	RiskScoreMax         = 100 // 最高风险评分
)

// 访问级别常量
const (
	AccessLevelHighest = 1 // 最高访问级别
	AccessLevelHigh    = 2 // 高访问级别
	AccessLevelMedium  = 3 // 中等访问级别
	AccessLevelLow     = 4 // 低访问级别
	AccessLevelLowest  = 5 // 最低访问级别
)

// 根据风险评分获取对应的访问级别
func GetAccessLevelByRiskScore(riskScore int) int {
	if riskScore >= RiskThresholdCritical {
		return AccessLevelLowest
	} else if riskScore >= RiskThresholdHigh {
		return AccessLevelLow
	} else if riskScore >= RiskThresholdMedium {
		return AccessLevelMedium
	} else if riskScore >= RiskThresholdLow {
		return AccessLevelHigh
	} else {
		return AccessLevelHighest
	}
}

// 根据风险评分获取对应的用户状态
func GetStatusByRiskScore(riskScore int) string {
	if riskScore >= RiskThresholdCritical {
		return StatusBlocked
	} else if riskScore >= RiskThresholdHigh {
		return StatusSuspended
	} else {
		return StatusActive
	}
}
