package models

// UserInfo 用户信息结构体
type UserInfo struct {
	DID           string `json:"did"`           // 用户的分布式身份标识符
	Name          string `json:"name"`          // 用户姓名
	IDNumber      string `json:"idNumber"`      // 身份证号
	PhoneNumber   string `json:"phoneNumber"`   // 电话号码
	VehicleID     string `json:"vehicleID"`     // 车辆标识号
	RiskScore     int    `json:"riskScore"`     // 用户风险评分 (0-100)
	Status        string `json:"status"`        // 用户状态: active
	CreatedAt     string `json:"createdAt"`     // 创建时间
	LastUpdatedAt string `json:"lastUpdatedAt"` // 最后更新时间
}

// 用户状态常量
const (
	StatusActive = "active"
)

// 初始风险值
const (
	InitialRiskScore = 0 // 初始风险评分
)
