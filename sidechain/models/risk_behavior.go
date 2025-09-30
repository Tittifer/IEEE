package models

// RiskBehaviorType 风险行为类型
type RiskBehaviorType string

// 定义风险行为类型常量
const (
	RiskBehaviorA RiskBehaviorType = "A" // 风险行为A
	RiskBehaviorB RiskBehaviorType = "B" // 风险行为B
)

// RiskBehaviorScaleFactor 风险行为尺度系数映射
var RiskBehaviorScaleFactor = map[RiskBehaviorType]int{
	RiskBehaviorA: 5,  // 风险行为A的尺度系数K为5
	RiskBehaviorB: 10, // 风险行为B的尺度系数K为10
}

// RiskBehavior 风险行为结构体
type RiskBehavior struct {
	Type            RiskBehaviorType `json:"type"`            // 风险行为类型
	SeverityLevel   int              `json:"severityLevel"`   // 风险行为严重程度s
	OccurrenceTime  int64            `json:"occurrenceTime"`  // 风险行为发生时间
}

// 风险评估算法相关参数常量
const (
	MaxRiskScore      = 1000 // 最大风险分数S_max
	DeltaParameter    = 0.5 // 影响低分时降温速度参数δ
	AlphaParameter    = 0.1 // 影响高分时降温速度参数α
)
