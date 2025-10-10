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
	maxScore       int     // 最高得分
	decayRate      float64 // 衰减速率系数 beta
	normalizationT float64 // 风险分数归一化基数 T
}

// NewRiskAssessor 创建新的风险评估器
func NewRiskAssessor(dbManager *db.DBManager) *RiskAssessor {
	return &RiskAssessor{
		dbManager:      dbManager,
		maxScore:       100,
		decayRate:      0.01, // 衰减速率系数，可调整
		normalizationT: 100,  // 归一化基数，可调整
	}
}

// AssessRisk 评估用户风险
func (r *RiskAssessor) AssessRisk(did string, behaviorType string) (int, error) {
	// 获取用户信息
	user, err := r.dbManager.GetUserByDID(did)
	if err != nil {
		return 0, fmt.Errorf("获取用户信息失败: %w", err)
	}
	if user == nil {
		return 0, fmt.Errorf("用户不存在: %s", did)
	}

	// 获取风险规则
	rule, err := r.dbManager.GetRiskRuleByType(behaviorType)
	if err != nil {
		return 0, fmt.Errorf("获取风险规则失败: %w", err)
	}
	if rule == nil {
		return 0, fmt.Errorf("风险规则不存在: %s", behaviorType)
	}

	// 记录风险行为
	if err := r.dbManager.RecordRiskBehavior(user.ID, behaviorType, rule.Score); err != nil {
		return 0, fmt.Errorf("记录风险行为失败: %w", err)
	}

	// 计算新的风险评分
	newScore, err := r.calculateRiskScore(user.ID, rule.Score)
	if err != nil {
		return 0, fmt.Errorf("计算风险评分失败: %w", err)
	}

	// 更新用户风险评分
	if err := r.dbManager.UpdateUserRiskScore(user.ID, newScore); err != nil {
		return 0, fmt.Errorf("更新用户风险评分失败: %w", err)
	}

	log.Printf("用户 %s 执行风险行为 %s，风险评分从 %d 更新为 %d", did, behaviorType, user.CurrentScore, newScore)
	return newScore, nil
}

// calculateRiskScore 计算风险评分
func (r *RiskAssessor) calculateRiskScore(userID int, currentBehaviorScore int) (int, error) {
	// 获取用户历史风险行为
	behaviors, err := r.dbManager.GetUserRiskBehaviors(userID, 50) // 获取最近50条记录
	if err != nil {
		return 0, fmt.Errorf("获取用户历史风险行为失败: %w", err)
	}

	// 计算历史风险行为影响因子 lambda
	now := time.Now()
	var historyImpact float64 = 0

	// 如果有历史行为，计算历史影响因子
	if len(behaviors) > 0 {
		for _, behavior := range behaviors {
			// 计算时间差（秒）
			timeDiff := now.Sub(behavior.Timestamp).Seconds()
			// 应用衰减公式: e^(-beta*(t-tj))
			decayFactor := math.Exp(-r.decayRate * timeDiff)
			// 累加历史影响
			historyImpact += decayFactor * float64(behavior.Score)
		}
		historyImpact /= r.normalizationT
	}

	// 计算 lambda = 1 + historyImpact
	lambda := 1.0 + historyImpact

	// 计算得分归一化
	// 当前我们简化处理，直接使用当前行为得分
	normalizedScore := float64(currentBehaviorScore)

	// 应用分数更新公式: S_t = min(S_max, Score * lambda)
	newScore := int(math.Min(float64(r.maxScore), normalizedScore*lambda))

	return newScore, nil
}

// GetCurrentRiskScore 获取用户当前风险评分
func (r *RiskAssessor) GetCurrentRiskScore(did string) (int, error) {
	user, err := r.dbManager.GetUserByDID(did)
	if err != nil {
		return 0, fmt.Errorf("获取用户信息失败: %w", err)
	}
	if user == nil {
		return 0, fmt.Errorf("用户不存在: %s", did)
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
