package models

// UserSessionStatus 用户会话状态
type UserSessionStatus string

// 用户会话状态常量
const (
	SessionStatusOnline  UserSessionStatus = "online"  // 用户在线
	SessionStatusOffline UserSessionStatus = "offline" // 用户离线
)

// UserSession 用户会话信息
type UserSession struct {
	DID                string           `json:"did"`                // 用户的分布式身份标识符
	Status             UserSessionStatus `json:"status"`            // 会话状态：online/offline
	SeverityLevel      int              `json:"severityLevel"`      // 当前会话中的风险行为严重程度累计值s
	Timestamp          int64            `json:"timestamp"`          // 最后更新时间戳
	HasTriggeredRisk   bool             `json:"hasTriggeredRisk"`   // 当前会话中是否已触发过风险行为
}

// NewUserSession 创建新的用户会话
func NewUserSession(did string, status UserSessionStatus, timestamp int64) *UserSession {
	severityLevel := 0
	if status == SessionStatusOffline {
		// 用户离线时严重程度为0
		severityLevel = 0
	}
	
	return &UserSession{
		DID:              did,
		Status:           status,
		SeverityLevel:    severityLevel,
		Timestamp:        timestamp,
		HasTriggeredRisk: false, // 初始化为未触发风险行为
	}
}

// IncrementSeverityLevel 增加严重程度值并标记已触发风险行为
func (s *UserSession) IncrementSeverityLevel() {
	if s.Status == SessionStatusOnline {
		s.SeverityLevel++
		s.HasTriggeredRisk = true // 标记已触发风险行为
	}
}

// UpdateStatus 更新用户状态
func (s *UserSession) UpdateStatus(status UserSessionStatus, timestamp int64) {
	s.Status = status
	s.Timestamp = timestamp
	
	// 如果用户登出，重置严重程度为0和风险触发标记
	if status == SessionStatusOffline {
		s.SeverityLevel = 0
		s.HasTriggeredRisk = false
	}
}