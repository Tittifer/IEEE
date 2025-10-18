package models

import (
	"time"
)

// DeviceInfo 设备信息结构体
type DeviceInfo struct {
	DID              string    `json:"did"`              // 设备的分布式身份标识符
	Name             string    `json:"name"`             // 设备名称
	Model            string    `json:"model"`            // 设备型号
	Vendor           string    `json:"vendor"`           // 设备供应商
	RiskScore        float64   `json:"riskScore"`        // 设备历史风险分数 (S_{t-1})，范围 [0, S_{max}]
	AttackIndexI     float64   `json:"attackIndexI"`     // 攻击画像指数 (I)，范围 [0, ∞)
	AttackProfile    []string  `json:"attackProfile"`    // 攻击画像，存储设备已触发过的不重复的行为类别
	LastEventTime    time.Time `json:"lastEventTime"`    // 上次事件时间 (t_{last})
	Status           string    `json:"status"`           // 设备状态: active, inactive, risky
	CreatedAt        time.Time `json:"createdAt"`        // 创建时间
	LastUpdatedAt    time.Time `json:"lastUpdatedAt"`    // 最后更新时间
}

// DeviceEvent 设备事件结构体，用于链码事件
type DeviceEvent struct {
	EventType    string    `json:"eventType"`    // 事件类型
	DID          string    `json:"did"`          // 设备DID
	Name         string    `json:"name"`         // 设备名称
	Timestamp    int64     `json:"timestamp"`    // 事件时间戳
	RiskScore    float64   `json:"riskScore"`    // 风险评分
	Category     string    `json:"category"`     // 行为类别
	BehaviorType string    `json:"behaviorType"` // 具体行为类型
}

// 事件类型常量
const (
	EventTypeRegister   = "register"    // 设备注册事件
	EventTypeConnect    = "connect"     // 设备连接事件
	EventTypeDisconnect = "disconnect"  // 设备断开连接事件
	EventTypeRiskUpdate = "risk_update" // 风险评分更新事件
	EventTypeRiskReset  = "risk_reset"  // 风险评分重置事件
)

// 设备状态常量
const (
	StatusActive   = "active"   // 设备活跃状态
	StatusInactive = "inactive" // 设备非活跃状态
	StatusRisky    = "risky"    // 设备风险状态
	StatusOnline   = "online"   // 设备在线状态
	StatusOffline  = "offline"  // 设备离线状态
)

// 风险评分相关常量
const (
	InitialRiskScore   = 0.00  // 初始风险评分
	RiskScoreThreshold = 50.00 // 风险评分阈值，超过此值将禁止设备连接
	MaxRiskScore       = 1000.0 // 最大风险分数 S_{max}
)

// 风险评估参数常量
const (
	Delta = 0.05  // 影响低分时降温速度参数
	Alpha = 0.02  // 影响高分时降温速度参数
	Lambda = 0.01 // 攻击画像指数衰减系数
)
