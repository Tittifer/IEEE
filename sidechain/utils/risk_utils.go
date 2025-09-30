package utils

import (
	"math"
	"time"

	"github.com/Tittifer/IEEE/sidechain/models"
)

// CalculateRiskScore 计算风险评分
// S_t = min(S_max, S_{t-1}^{'} + K*s)
// 其中：
// S_t：本次计算分数
// S_{t-1}^{'}：衰减后旧得分
// K：用于把严重程度换算为得分的尺度系数
// s：用户恶意行为严重程度
func CalculateRiskScore(oldScore int, behaviorType models.RiskBehaviorType, severityLevel int, lastUpdateTime int64, currentTime int64) int {
	// 获取尺度系数K
	scaleFactor := models.RiskBehaviorScaleFactor[behaviorType]
	
	// 计算衰减后的旧得分
	decayedOldScore := CalculateScoreDecay(oldScore, lastUpdateTime, currentTime)
	
	// 计算新得分
	newScore := decayedOldScore + scaleFactor*severityLevel
	
	// 确保不超过最大分数
	if newScore > models.MaxRiskScore {
		newScore = models.MaxRiskScore
	}
	
	return newScore
}

// CalculateScoreDecay 计算分数衰减
// S_{t-1}^{'} = max(0, S_{t-1} - (δ*Δt)/(1+α*S_{t-1}))
// 其中：
// S_{t-1}：上次计算得分
// δ：影响低分时降温速度参数
// Δt：时间间隔（秒）
// α：影响高分时降温速度参数
func CalculateScoreDecay(oldScore int, lastUpdateTime int64, currentTime int64) int {
	// 如果是初始状态或时间戳无效，直接返回原始分数
	if lastUpdateTime <= 0 || currentTime <= lastUpdateTime {
		return oldScore
	}
	
	// 计算时间间隔（秒）
	deltaTime := currentTime - lastUpdateTime
	
	// 计算分母
	denominator := 1.0 + models.AlphaParameter*float64(oldScore)
	
	// 计算衰减值
	decayValue := (models.DeltaParameter * float64(deltaTime)) / denominator
	
	// 计算衰减后的分数
	decayedScore := float64(oldScore) - decayValue
	
	// 确保不小于0
	if decayedScore < 0 {
		decayedScore = 0
	}
	
	return int(math.Floor(decayedScore))
}

// GetCurrentTimestamp 获取当前时间戳（秒）
func GetCurrentTimestamp() int64 {
	return time.Now().Unix()
}
