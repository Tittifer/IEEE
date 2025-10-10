package client

import (
	"fmt"
	"log"
	"math"
	"sync"
	"time"
)

// RiskBehavior 风险行为类型
type RiskBehavior string

// 风险行为类型常量
const (
	RiskBehaviorA RiskBehavior = "A"
	RiskBehaviorB RiskBehavior = "B"
)

// 风险评分相关常量
const (
	InitialRiskScore   = 0   // 初始风险评分
	RiskScoreThreshold = 50  // 风险评分阈值
	MaxRiskScore       = 100 // 最高风险评分
	
	// 风险评分算法参数
	DecayFactor = 0.1  // 衰减因子 β
	NormFactor  = 100  // 归一化基数 T
)

// RiskScoreManager 风险评分管理器
type RiskScoreManager struct {
	userScores      map[string]*UserRiskScore // DID -> 风险评分信息
	mutex           sync.RWMutex              // 读写锁
	honeypointClient *HoneypointClient        // 后台客户端引用
	dbManager       *DBManager                // 数据库管理器
}

// UserRiskScore 用户风险评分信息
type UserRiskScore struct {
	DID           string    // 用户DID
	Score         int       // 当前风险评分
	LastUpdateTime time.Time // 上次更新时间
}

// NewRiskScoreManager 创建新的风险评分管理器
func NewRiskScoreManager(honeypointClient *HoneypointClient, dbManager *DBManager) *RiskScoreManager {
	return &RiskScoreManager{
		userScores:      make(map[string]*UserRiskScore),
		honeypointClient: honeypointClient,
		dbManager:       dbManager,
	}
}

// RegisterUser 注册新用户，初始化风险评分为0
func (r *RiskScoreManager) RegisterUser(did string, name string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// 检查用户是否已存在
	if _, exists := r.userScores[did]; exists {
		log.Printf("用户 %s 已存在，无需重复注册", did)
		return nil
	}

	// 初始化用户风险评分
	r.userScores[did] = &UserRiskScore{
		DID:           did,
		Score:         InitialRiskScore,
		LastUpdateTime: time.Now(),
	}

	// 将用户信息保存到数据库
	if err := r.dbManager.RegisterUser(did, name); err != nil {
		log.Printf("注册用户到数据库失败: %v", err)
		// 即使数据库操作失败，仍保留内存中的用户信息
	}

	log.Printf("用户 %s 注册成功，初始风险评分: %d", did, InitialRiskScore)
	return nil
}

// UserLogin 用户登录，记录登录事件
func (r *RiskScoreManager) UserLogin(did string) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// 如果用户不存在，则创建用户记录
	if _, exists := r.userScores[did]; !exists {
		r.userScores[did] = &UserRiskScore{
			DID:           did,
			Score:         InitialRiskScore,
			LastUpdateTime: time.Now(),
		}
		log.Printf("用户 %s 首次登录，初始化风险评分: %d", did, InitialRiskScore)
	} else {
		log.Printf("用户 %s 登录，当前风险评分: %d", did, r.userScores[did].Score)
	}
}

// UserLogout 用户登出，记录登出事件
func (r *RiskScoreManager) UserLogout(did string) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	if userScore, exists := r.userScores[did]; exists {
		log.Printf("用户 %s 登出，最终风险评分: %d", did, userScore.Score)
	} else {
		log.Printf("用户 %s 登出，但用户记录不存在", did)
	}
}

// ProcessRiskBehavior 处理风险行为，更新风险评分
func (r *RiskScoreManager) ProcessRiskBehavior(did string, behavior RiskBehavior) (int, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// 检查用户是否存在
	userScore, exists := r.userScores[did]
	if !exists {
		return 0, fmt.Errorf("用户 %s 不存在", did)
	}

	// 记录风险行为到数据库
	behaviorScore, err := r.dbManager.RecordRiskBehavior(did, string(behavior))
	if err != nil {
		log.Printf("记录风险行为到数据库失败: %v", err)
		// 如果数据库操作失败，使用内存中的风险行为分数
		if rule, err := r.dbManager.GetRiskRule(string(behavior)); err == nil {
			behaviorScore = rule.Score
		} else {
			// 如果无法从数据库获取规则，使用默认值
			switch behavior {
			case RiskBehaviorA:
				behaviorScore = 5
			case RiskBehaviorB:
				behaviorScore = 10
			default:
				return 0, fmt.Errorf("未知的风险行为类型: %s", behavior)
			}
		}
	}

	// 获取用户历史风险行为记录
	behaviors, err := r.dbManager.GetUserRiskBehaviors(did, 50)
	if err != nil {
		log.Printf("获取用户历史风险行为失败: %v", err)
		// 如果无法获取历史记录，只使用当前行为计算
		behaviors = []RiskBehaviorRecord{}
	}

	// 计算历史风险行为影响因子 λ
	now := time.Now()
	lambda := 1.0
	
	// 如果有历史行为记录，计算历史影响因子
	if len(behaviors) > 0 {
		var historyImpact float64
		for _, b := range behaviors {
			// 计算时间差（秒）
			timeDiff := now.Sub(b.Timestamp).Seconds()
			// 根据公式计算历史行为的影响: e^(-β(t-tj)) * Sp,j
			impact := math.Exp(-DecayFactor * timeDiff) * float64(b.Score)
			historyImpact += impact
		}
		
		// 根据公式: λ = 1 + (∑e^(-β(t-tj)) * Sp,j) / T
		lambda = 1.0 + (historyImpact / NormFactor)
	}

	// 计算当前时间与上次更新时间的差值（秒）
	deltaT := now.Sub(userScore.LastUpdateTime).Seconds()
	
	// 应用冷却曲线，计算衰减后的分数
	// 冷却曲线公式: decayedScore = score - (delta * deltaT) / (1 + alpha * score)
	alpha := 0.1
	delta := 0.5
	scoreFloat := float64(userScore.Score)
	decayedScore := scoreFloat - (delta * deltaT) / (1 + alpha * scoreFloat)
	if decayedScore < 0 {
		decayedScore = 0
	}
	
	// 根据新的风险评分算法计算最终分数
	// 1. 计算归一化得分 Score (这里简化为当前行为得分)
	normalizedScore := float64(behaviorScore)
	
	// 2. 应用历史影响因子: St = min(Smax, Score * λ)
	newScoreFloat := math.Min(float64(MaxRiskScore), normalizedScore * lambda)
	newScore := int(math.Round(decayedScore + newScoreFloat))
	
	// 确保分数不超过最大值
	if newScore > MaxRiskScore {
		newScore = MaxRiskScore
	}

	// 更新用户风险评分
	userScore.Score = newScore
	userScore.LastUpdateTime = now

	// 更新数据库中的用户风险评分
	if err := r.dbManager.UpdateUserRiskScore(did, newScore); err != nil {
		log.Printf("更新数据库中的用户风险评分失败: %v", err)
		// 即使数据库更新失败，仍保留内存中的更新
	}

	log.Printf("用户 %s 触发风险行为 %s，行为分数=%d，历史影响因子=%.2f，时间间隔=%.2f秒，衰减后分数=%.2f，新分数=%d",
		did, behavior, behaviorScore, lambda, deltaT, decayedScore, newScore)

	if newScore >= RiskScoreThreshold {
		log.Printf("警告：用户 %s 风险评分 %d 已超过阈值 %d", did, newScore, RiskScoreThreshold)
	}

	return newScore, nil
}

// GetUserRiskScore 获取用户当前风险评分
func (r *RiskScoreManager) GetUserRiskScore(did string) (int, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	// 先尝试从内存获取
	if userScore, exists := r.userScores[did]; exists {
		return userScore.Score, nil
	}

	// 如果内存中不存在，则从数据库获取
	user, err := r.dbManager.GetUser(did)
	if err != nil {
		return 0, fmt.Errorf("获取用户风险评分失败: %w", err)
	}

	// 更新内存中的用户信息
	r.userScores[did] = &UserRiskScore{
		DID:           did,
		Score:         user.CurrentScore,
		LastUpdateTime: user.LastUpdate,
	}

	return user.CurrentScore, nil
}

// ReportRiskScore 向链上报告用户风险评分
func (r *RiskScoreManager) ReportRiskScore(did string) error {
	score, err := r.GetUserRiskScore(did)
	if err != nil {
		return fmt.Errorf("获取用户风险评分失败: %w", err)
	}

	// 调用链码更新风险评分
	_, err = r.honeypointClient.contract.SubmitTransaction("IdentityContract:UpdateRiskScore", did, fmt.Sprintf("%d", score))
	if err != nil {
		return fmt.Errorf("提交风险评分更新交易失败: %w", err)
	}

	log.Printf("已向链上报告用户 %s 的风险评分: %d", did, score)
	return nil
}

// GetAllUserScores 获取所有用户的风险评分
func (r *RiskScoreManager) GetAllUserScores() map[string]int {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	scores := make(map[string]int)
	for did, userScore := range r.userScores {
		scores[did] = userScore.Score
	}
	return scores
}