package risk

import (
	"fmt"
	"log"
	"math"
	"time"

	"github.com/Tittifer/IEEE/honeypoint_client/db"
)

// RiskAssessor 风险评估器
type RiskAssessor struct {
	dbManager *db.DBManager
	// 风险评估参数
	maxScore       float64 // 最高得分，修改为float64类型
	decayRate      float64 // 衰减速率系数 beta
	normalizationT float64 // 风险分数归一化基数 T
	delta          float64 // 影响低分时降温速度参数
	alpha          float64 // 影响高分时降温速度参数
}

// NewRiskAssessor 创建新的风险评估器
func NewRiskAssessor(dbManager *db.DBManager) *RiskAssessor {
	return &RiskAssessor{
		dbManager:      dbManager,
		maxScore:       100.00, // 修改为浮点数
		decayRate:      0.01,   // 衰减速率系数，可调整
		normalizationT: 100.00, // 归一化基数，可调整
		delta:          0.05,   // 影响低分时降温速度参数，可调整
		alpha:          0.02,   // 影响高分时降温速度参数，可调整
	}
}

// AssessRisk 评估用户风险
func (r *RiskAssessor) AssessRisk(did string, behaviorType string) (float64, error) { // 返回值修改为float64类型
	// 获取用户信息
	user, err := r.dbManager.GetUserByDID(did)
	if err != nil {
		return 0.00, fmt.Errorf("获取用户信息失败: %w", err)
	}
	if user == nil {
		return 0.00, fmt.Errorf("用户不存在: %s", did)
	}

	// 获取风险规则
	rule, err := r.dbManager.GetRiskRuleByType(behaviorType)
	if err != nil {
		return 0.00, fmt.Errorf("获取风险规则失败: %w", err)
	}
	if rule == nil {
		return 0.00, fmt.Errorf("风险规则不存在: %s", behaviorType)
	}

	// 记录风险行为
	if err := r.dbManager.RecordRiskBehavior(user.ID, behaviorType, rule.Score); err != nil {
		return 0.00, fmt.Errorf("记录风险行为失败: %w", err)
	}

	// 计算新的风险评分
	newScore, err := r.calculateRiskScore(user.ID, rule.Score)
	if err != nil {
		return 0.00, fmt.Errorf("计算风险评分失败: %w", err)
	}

	// 更新用户风险评分
	if err := r.dbManager.UpdateUserRiskScore(user.ID, newScore); err != nil {
		return 0.00, fmt.Errorf("更新用户风险评分失败: %w", err)
	}

	log.Printf("用户 %s 执行风险行为 %s，风险评分从 %.2f 更新为 %.2f", did, behaviorType, user.CurrentScore, newScore)
	return newScore, nil
}

// calculateRiskScore 计算风险评分
func (r *RiskAssessor) calculateRiskScore(userID int, currentBehaviorScore float64) (float64, error) { // 参数和返回值修改为float64类型
	// 获取用户信息，以获取当前分数
	user, err := r.dbManager.GetUserByID(userID)
	if err != nil {
		return 0.00, fmt.Errorf("获取用户信息失败: %w", err)
	}
	if user == nil {
		return 0.00, fmt.Errorf("用户不存在: ID=%d", userID)
	}

	// 获取用户历史风险行为
	behaviors, err := r.dbManager.GetUserRiskBehaviors(userID, 50) // 获取最近50条记录
	if err != nil {
		return 0.00, fmt.Errorf("获取用户历史风险行为失败: %w", err)
	}

	// 计算历史风险行为影响因子 lambda
	now := time.Now()
	var historyImpact float64 = 0.00

	// 如果有历史行为，计算历史影响因子
	if len(behaviors) > 0 {
		for _, behavior := range behaviors {
			// 计算时间差（秒）
			timeDiff := now.Sub(behavior.Timestamp).Seconds()
			// 应用衰减公式: e^(-beta*(t-tj))
			decayFactor := math.Exp(-r.decayRate * timeDiff)
			// 累加历史影响
			historyImpact += decayFactor * behavior.Score
		}
		historyImpact /= r.normalizationT
	}

	// 计算 lambda = 1 + historyImpact
	lambda := 1.0 + historyImpact

	// 应用降温曲线计算上次得分的衰减值
	// S_{t-1}^{'}=max(0,S_{t-1}-\frac{\delta*\Delta t}{1+\alpha*S_{t-1}})
	var cooledPreviousScore float64 = 0.00
	
	// 使用用户表中的last_update字段计算时间差
	// 计算时间差（秒）
	deltaT := now.Sub(user.LastUpdate).Seconds()
	
	// 应用降温曲线
	coolingFactor := (r.delta * deltaT) / (1 + r.alpha * user.CurrentScore)
	cooledPreviousScore = math.Max(0.00, user.CurrentScore - coolingFactor)
	log.Printf("用户当前分数 %.2f 经过 %.2f 秒的降温后变为 %.2f", user.CurrentScore, deltaT, cooledPreviousScore)

	// 计算得分归一化
	// 当前我们简化处理，直接使用当前行为得分
	normalizedScore := currentBehaviorScore

	// 应用分数更新公式: S_t = min(S_max, (Score + S_{t-1}^{'}) * lambda)
	newScore := math.Min(r.maxScore, (normalizedScore + cooledPreviousScore) * lambda)

	// 保留两位小数
	newScore = math.Round(newScore*100) / 100

	return newScore, nil
}

// GetCurrentRiskScore 获取用户当前风险评分
func (r *RiskAssessor) GetCurrentRiskScore(did string) (float64, error) { // 返回值修改为float64类型
	user, err := r.dbManager.GetUserByDID(did)
	if err != nil {
		return 0.00, fmt.Errorf("获取用户信息失败: %w", err)
	}
	if user == nil {
		return 0.00, fmt.Errorf("用户不存在: %s", did)
	}

	return user.CurrentScore, nil
}

// ListAvailableRiskBehaviors 列出可用的风险行为类型
func (r *RiskAssessor) ListAvailableRiskBehaviors() ([]db.RiskRule, error) {
	rules, err := r.dbManager.GetAllRiskRules()
	if err != nil {
		return nil, fmt.Errorf("获取风险规则失败: %w", err)
	}
	return rules, nil
}