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
	maxScore float64 // 最高得分 S_{max}
	delta    float64 // 影响低分时降温速度参数 δ
	alpha    float64 // 影响高分时降温速度参数 α
	lambda   float64 // 攻击画像指数衰减系数 λ
}

// NewRiskAssessor 创建新的风险评估器
func NewRiskAssessor(dbManager *db.DBManager) *RiskAssessor {
	return &RiskAssessor{
		dbManager: dbManager,
		maxScore:  1000.0, // 最大风险分数 S_{max}
		delta:     0.05,   // 影响低分时降温速度参数 δ
		alpha:     0.02,   // 影响高分时降温速度参数 α
		lambda:    0.01,   // 攻击画像指数衰减系数 λ
	}
}

// AssessRisk 评估设备风险
func (r *RiskAssessor) AssessRisk(did string, behaviorType string) (float64, float64, []string, error) {
	// 从链上获取设备信息
	device, err := r.dbManager.GetDeviceFromChain(did)
	if err != nil {
		return 0.0, 0.0, nil, fmt.Errorf("获取设备信息失败: %w", err)
	}
	if device == nil {
		return 0.0, 0.0, nil, fmt.Errorf("设备不存在: %s", did)
	}

	// 获取风险规则
	rule, err := r.dbManager.GetRiskRuleByType(behaviorType)
	if err != nil {
		return 0.0, 0.0, nil, fmt.Errorf("获取风险规则失败: %w", err)
	}
	if rule == nil {
		return 0.0, 0.0, nil, fmt.Errorf("风险规则不存在: %s", behaviorType)
	}

	// 计算新的风险评分
	newScore, newAttackIndex, updatedProfile, err := r.calculateRiskScore(device, rule.Score, rule.Category, rule.Weight)
	if err != nil {
		return 0.0, 0.0, nil, fmt.Errorf("计算风险评分失败: %w", err)
	}

	log.Printf("设备 %s 执行风险行为 %s，风险评分从 %.2f 更新为 %.2f，攻击画像指数从 %.2f 更新为 %.2f",
		did, behaviorType, device.RiskScore, newScore, device.AttackIndexI, newAttackIndex)
	
	return newScore, newAttackIndex, updatedProfile, nil
}

// calculateRiskScore 计算风险评分
// 按照大纲中的风险评估算法实现
func (r *RiskAssessor) calculateRiskScore(device *db.Device, baseScore float64, category string, weight float64) (float64, float64, []string, error) {
	// 复制当前攻击画像
	attackProfile := make([]string, len(device.AttackProfile))
	copy(attackProfile, device.AttackProfile)
	
	// 步骤2：更新攻击画像指数 (I)
	var deltaI float64 = 0.0
	
	// 检查当前Category是否已存在于该设备的Attack Profile集合中
	categoryExists := false
	for _, c := range attackProfile {
		if c == category {
			categoryExists = true
			break
		}
	}
	
	// 如果不存在（意图升级）
	if !categoryExists {
		deltaI = weight
		// 将当前Category添加到Attack Profile集合中
		attackProfile = append(attackProfile, category)
	}
	
	// 完成I的累加
	newAttackIndex := device.AttackIndexI + deltaI
	
	// 步骤3：激活威胁状态（已经在调用此函数前更新了t_last）
	
	// 步骤4：计算实时风险分 (S_t)
	// 计算时间间隔 Δt（秒）
	now := time.Now()
	deltaT := now.Sub(device.LastEventTime).Seconds()
	
	// 对历史分数进行降温
	coolingFactor := (r.delta * deltaT) / (1 + r.alpha * device.RiskScore)
	cooledPreviousScore := math.Max(0.0, device.RiskScore - coolingFactor)
	
	log.Printf("设备当前分数 %.2f 经过 %.2f 秒的降温后变为 %.2f", device.RiskScore, deltaT, cooledPreviousScore)
	
	// 计算最终得分
	newScore := math.Min(r.maxScore, baseScore * (1 + newAttackIndex) + cooledPreviousScore)
	
	// 步骤5：后台状态维护（周期性任务）- 在另一个函数中实现
	
	return newScore, newAttackIndex, attackProfile, nil
}

// PerformBackgroundMaintenance 执行后台状态维护
// 周期性调用此函数，用于攻击画像指数的慢速衰减
func (r *RiskAssessor) PerformBackgroundMaintenance(did string) error {
	// 从链上获取设备信息
	device, err := r.dbManager.GetDeviceFromChain(did)
	if err != nil {
		return fmt.Errorf("获取设备信息失败: %w", err)
	}
	if device == nil {
		return fmt.Errorf("设备不存在: %s", did)
	}
	
	// 计算时间间隔（秒）
	now := time.Now()
	deltaT := now.Sub(device.LastEventTime).Seconds()
	
	// I慢速衰减: 确保旧的威胁记录会随时间慢慢淡化
	// I_{new} = I_{old} * e^{-λ*Δt}
	newAttackIndex := device.AttackIndexI * math.Exp(-r.lambda * deltaT)
	
	log.Printf("设备 %s 的攻击画像指数从 %.2f 衰减为 %.2f", did, device.AttackIndexI, newAttackIndex)
	
	// 更新到链上（通过数据库管理器）
	return r.dbManager.UpdateDeviceAttackIndex(did, newAttackIndex)
}

// GetCurrentRiskScore 获取设备当前风险评分
func (r *RiskAssessor) GetCurrentRiskScore(did string) (float64, error) {
	device, err := r.dbManager.GetDeviceFromChain(did)
	if err != nil {
		return 0.0, fmt.Errorf("获取设备信息失败: %w", err)
	}
	if device == nil {
		return 0.0, fmt.Errorf("设备不存在: %s", did)
	}

	return device.RiskScore, nil
}

// ListAvailableRiskBehaviors 列出可用的风险行为类型
func (r *RiskAssessor) ListAvailableRiskBehaviors() ([]db.RiskRule, error) {
	rules, err := r.dbManager.GetAllRiskRules()
	if err != nil {
		return nil, fmt.Errorf("获取风险规则失败: %w", err)
	}
	return rules, nil
}