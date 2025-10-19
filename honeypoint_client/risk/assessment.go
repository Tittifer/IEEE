package risk

import (
	"fmt"
	"log"
	"math"
	"strings"
	"time"

	"github.com/Tittifer/IEEE/honeypoint_client/chain"
)

// RiskAssessor 风险评估器
type RiskAssessor struct {
	chainManager *chain.ChainManager
	// 风险评估参数
	maxScore float64 // 最高得分 S_{max}
	delta    float64 // 影响低分时降温速度参数 δ
	alpha    float64 // 影响高分时降温速度参数 α
	lambda   float64 // 攻击画像指数衰减系数 λ
}

// NewRiskAssessor 创建新的风险评估器
func NewRiskAssessor(chainManager *chain.ChainManager) *RiskAssessor {
	return &RiskAssessor{
		chainManager: chainManager,
		maxScore:  1000.0, // 最大风险分数 S_{max}
		delta:     2.00,   // 影响低分时降温速度参数 δ
		alpha:     0.05,   // 影响高分时降温速度参数 α
		lambda:    0.01,   // 攻击画像指数衰减系数 λ
	}
}

// AssessRisk 评估设备风险
func (r *RiskAssessor) AssessRisk(did string, behaviorType string) (float64, float64, []string, error) {
	// 从链上获取设备信息
	device, err := r.chainManager.GetDeviceFromChain(did)
	if err != nil {
		return 0.0, 0.0, nil, fmt.Errorf("获取设备信息失败: %w", err)
	}
	if device == nil {
		return 0.0, 0.0, nil, fmt.Errorf("设备不存在: %s", did)
	}

	// 从代码中获取风险规则
	rule := GetRiskRuleByType(behaviorType)
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
func (r *RiskAssessor) calculateRiskScore(device *chain.Device, baseScore float64, category string, weight float64) (float64, float64, []string, error) {
	// 复制当前攻击画像
	attackProfile := make([]string, len(device.AttackProfile))
	copy(attackProfile, device.AttackProfile)
	
	// 步骤2：更新攻击画像指数 (I)
	var deltaI float64 = 0.0
	
	// 检查当前Category是否已存在于该设备的Attack Profile集合中
	// 注意：由于我们将Category格式从简单的“Recon”更新为“Recon.NetworkScan”等更详细的格式
	// 我们需要处理两种情况：完全匹配和主类别匹配（点号前的部分）
	categoryExists := false
	mainCategory := category
	if dotIndex := strings.Index(category, "."); dotIndex != -1 {
		mainCategory = category[:dotIndex]
	}
	
	for _, c := range attackProfile {
		// 完全匹配
		if c == category {
			categoryExists = true
			break
		}
		
		// 主类别匹配（兼容旧格式）
		if c == mainCategory || (strings.HasPrefix(c, mainCategory+".")) {
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
	// 计算时间间隔 Δt（天）
	now := time.Now()
	deltaT := now.Sub(device.LastEventTime).Hours() / 24.0
	
	// 对历史分数进行降温
	coolingFactor := (r.delta * deltaT) / (1 + r.alpha * device.RiskScore)
	cooledPreviousScore := math.Max(0.0, device.RiskScore - coolingFactor)
	
	log.Printf("设备当前分数 %.2f 经过 %.2f 天的降温后变为 %.2f", device.RiskScore, deltaT, cooledPreviousScore)
	
	// 计算最终得分
	newScore := math.Min(r.maxScore, baseScore * (1 + newAttackIndex) + cooledPreviousScore)
	
	// 步骤5：后台状态维护（周期性任务）- 在另一个函数中实现
	
	return newScore, newAttackIndex, attackProfile, nil
}

// PerformBackgroundMaintenance 执行后台状态维护
// 周期性调用此函数，用于攻击画像指数的慢速衰减
func (r *RiskAssessor) PerformBackgroundMaintenance(did string) error {
	// 从链上获取设备信息
	device, err := r.chainManager.GetDeviceFromChain(did)
	if err != nil {
		return fmt.Errorf("获取设备信息失败: %w", err)
	}
	if device == nil {
		return fmt.Errorf("设备不存在: %s", did)
	}
	
	// 计算时间间隔（天）
	now := time.Now()
	deltaT := now.Sub(device.LastEventTime).Hours() / 24.0
	
	// I慢速衰减: 确保旧的威胁记录会随时间慢慢淡化
	// I_{new} = I_{old} * e^{-λ*Δt}
	newAttackIndex := device.AttackIndexI * math.Exp(-r.lambda * deltaT)
	
	log.Printf("设备 %s 的攻击画像指数从 %.2f 衰减为 %.2f (经过 %.2f 天)", did, device.AttackIndexI, newAttackIndex, deltaT)
	
	// 更新到链上
	return r.chainManager.UpdateDeviceAttackIndex(did, newAttackIndex)
}

// GetCurrentRiskScore 获取设备当前风险评分
func (r *RiskAssessor) GetCurrentRiskScore(did string) (float64, error) {
	device, err := r.chainManager.GetDeviceFromChain(did)
	if err != nil {
		return 0.0, fmt.Errorf("获取设备信息失败: %w", err)
	}
	if device == nil {
		return 0.0, fmt.Errorf("设备不存在: %s", did)
	}

	return device.RiskScore, nil
}

// ListAvailableRiskBehaviors 列出可用的风险行为类型
func (r *RiskAssessor) ListAvailableRiskBehaviors() []RiskRule {
	return GetAllRiskRules()
}