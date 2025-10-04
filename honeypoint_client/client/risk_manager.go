package client

import (
	"fmt"
	"log"
	"sync"
	"time"
)

// RiskBehavior 风险行为类型
type RiskBehavior string

const (
	// RiskBehaviorA 风险行为A
	RiskBehaviorA RiskBehavior = "A"
	// RiskBehaviorB 风险行为B
	RiskBehaviorB RiskBehavior = "B"
)

// RiskBehaviorScore 风险行为对应的评分
var RiskBehaviorScore = map[RiskBehavior]int{
	RiskBehaviorA: 5,  // 风险行为A的K值为5
	RiskBehaviorB: 10, // 风险行为B的K值为10
}

// RiskScoreThreshold 风险评分阈值，超过此值将强制用户登出
const RiskScoreThreshold = 50

// RiskScoreManager 风险评分管理器
type RiskScoreManager struct {
	userScores     map[string]*UserRiskScore // DID -> 风险评分信息
	mutex          sync.RWMutex              // 读写锁
	honeypointClient *HoneypointClient       // 后台客户端引用
}

// UserRiskScore 用户风险评分信息
type UserRiskScore struct {
	DID           string    // 用户DID
	Score         int       // 当前风险评分
	LastUpdateTime time.Time // 上次更新时间
}

// NewRiskScoreManager 创建新的风险评分管理器
func NewRiskScoreManager(client *HoneypointClient) *RiskScoreManager {
	return &RiskScoreManager{
		userScores:     make(map[string]*UserRiskScore),
		honeypointClient: client,
	}
}

// RegisterUser 注册新用户，初始化风险评分为0
func (r *RiskScoreManager) RegisterUser(did string) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	if _, exists := r.userScores[did]; !exists {
		r.userScores[did] = &UserRiskScore{
			DID:           did,
			Score:         0, // 初始风险评分为0
			LastUpdateTime: time.Now(),
		}
		log.Printf("用户 %s 已注册，初始风险评分为0", did)
	} else {
		log.Printf("用户 %s 已存在，不需要重新注册", did)
	}
}

// UserLogin 用户登录，记录用户登录事件
func (r *RiskScoreManager) UserLogin(did string) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	// 确保用户已注册
	if _, exists := r.userScores[did]; !exists {
		r.userScores[did] = &UserRiskScore{
			DID:           did,
			Score:         0,
			LastUpdateTime: time.Now(),
		}
		log.Printf("用户 %s 首次登录，初始化风险评分为0", did)
	}
	
	log.Printf("用户 %s 已登录", did)
}

// UserLogout 用户登出，记录用户登出事件
func (r *RiskScoreManager) UserLogout(did string) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	log.Printf("用户 %s 已登出", did)
}

// ProcessRiskBehavior 处理风险行为，更新风险评分
func (r *RiskScoreManager) ProcessRiskBehavior(did string, behavior RiskBehavior) (int, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	// 获取用户当前风险评分信息
	userScore, exists := r.userScores[did]
	if !exists {
		return 0, fmt.Errorf("用户 %s 不存在", did)
	}
	
	// 获取行为对应的K值
	k, exists := RiskBehaviorScore[behavior]
	if !exists {
		return 0, fmt.Errorf("未知的风险行为类型: %s", behavior)
	}
	
	// 计算时间间隔（秒）
	now := time.Now()
	deltaT := now.Sub(userScore.LastUpdateTime).Seconds()
	
	// 应用降温曲线，计算衰减后的旧得分
	alpha := 0.1 // 影响高分时降温速度参数
	delta := 0.5 // 影响低分时降温速度参数
	
	// 将整数转换为浮点数进行计算
	scoreFloat := float64(userScore.Score)
	decayedScore := scoreFloat - (delta*deltaT)/(1+alpha*scoreFloat)
	if decayedScore < 0 {
		decayedScore = 0
	}
	
	// 应用分数更新公式
	const MaxScore = 100 // 最大风险评分
	newScore := int(decayedScore) + k
	if newScore > MaxScore {
		newScore = MaxScore
	}
	
	// 更新用户风险评分
	userScore.Score = newScore
	userScore.LastUpdateTime = now
	
	log.Printf("用户 %s 触发风险行为 %s，K值=%d，时间间隔=%.2f秒，旧分数=%.2f，新分数=%d", 
		did, behavior, k, deltaT, decayedScore, newScore)
	
	// 检查是否超过风险阈值
	if newScore >= RiskScoreThreshold {
		log.Printf("警告：用户 %s 风险评分 %d 已超过阈值 %d", did, newScore, RiskScoreThreshold)
	}
	
	return newScore, nil
}

// GetUserRiskScore 获取用户当前风险评分
func (r *RiskScoreManager) GetUserRiskScore(did string) (int, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	userScore, exists := r.userScores[did]
	if !exists {
		return 0, fmt.Errorf("用户 %s 不存在", did)
	}
	
	return userScore.Score, nil
}

// ReportRiskScore 向链上报告用户最新风险评分
func (r *RiskScoreManager) ReportRiskScore(did string) error {
	r.mutex.RLock()
	userScore, exists := r.userScores[did]
	r.mutex.RUnlock()
	
	if !exists {
		return fmt.Errorf("用户 %s 不存在", did)
	}
	
	// 调用链上合约更新风险评分
	log.Printf("向链上报告用户 %s 的风险评分: %d", did, userScore.Score)
	
	// 实际调用链上合约的代码
	_, err := r.honeypointClient.contract.SubmitTransaction("UpdateRiskScore", did, fmt.Sprintf("%d", userScore.Score))
	if err != nil {
		return fmt.Errorf("向链上报告风险评分失败: %w", err)
	}
	
	log.Printf("成功向链上报告用户 %s 的风险评分: %d", did, userScore.Score)
	return nil
}

// GetAllUserScores 获取所有用户的风险评分
func (r *RiskScoreManager) GetAllUserScores() map[string]int {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	result := make(map[string]int)
	for did, userScore := range r.userScores {
		result[did] = userScore.Score
	}
	
	return result
}

