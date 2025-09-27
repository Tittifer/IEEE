package contracts

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/Tittifer/IEEE/models"
)

// RiskContract 风险评估合约
type RiskContract struct {
	contractapi.Contract
}

// RecordAttack 记录用户攻击行为
func (c *RiskContract) RecordAttack(ctx contractapi.TransactionContextInterface, did, honeypotID, attackType, description string, severity int) error {
	// 获取用户信息
	userContract := new(UserContract)
	userInfo, err := userContract.GetUser(ctx, did)
	if err != nil {
		return err
	}

	// 创建新的攻击记录
	// 使用交易时间戳而不是当前时间，确保确定性
	timestamp, err := ctx.GetStub().GetTxTimestamp()
	if err != nil {
		return fmt.Errorf("获取交易时间戳失败: %v", err)
	}
	txTime := time.Unix(timestamp.Seconds, int64(timestamp.Nanos))
	
	attack := models.Attack{
		Timestamp:   txTime,
		HoneypotID:  honeypotID,
		AttackType:  attackType,
		Description: description,
		Severity:    severity,
	}

	// 更新用户的攻击历史
	userInfo.AttackHistory = append(userInfo.AttackHistory, attack)
	
	// 更新用户的风险评分
	// 根据攻击严重程度增加风险评分，最高不超过100
	newRiskScore := userInfo.RiskScore + (severity * 5)
	if newRiskScore > models.RiskScoreMax {
		newRiskScore = models.RiskScoreMax
	}
	userInfo.RiskScore = newRiskScore
	
	// 根据风险评分调整访问级别和状态
	userInfo.AccessLevel = models.GetAccessLevelByRiskScore(newRiskScore)
	userInfo.Status = models.GetStatusByRiskScore(newRiskScore)
	
	userInfo.LastUpdatedAt = txTime

	// 将更新后的用户信息转换为JSON并存储
	userInfoJSON, err := json.Marshal(userInfo)
	if err != nil {
		return fmt.Errorf("用户信息序列化失败: %v", err)
	}

	// 将更新后的用户信息写入账本
	err = ctx.GetStub().PutState(did, userInfoJSON)
	if err != nil {
		return fmt.Errorf("更新用户信息时出错: %v", err)
	}

	return nil
}

// UpdateRiskScore 手动更新用户风险评分
func (c *RiskContract) UpdateRiskScore(ctx contractapi.TransactionContextInterface, did string, newRiskScore int) error {
	// 检查风险评分是否在有效范围内
	if newRiskScore < 0 || newRiskScore > models.RiskScoreMax {
		return fmt.Errorf("风险评分必须在0到%d之间", models.RiskScoreMax)
	}

	// 获取用户信息
	userContract := new(UserContract)
	userInfo, err := userContract.GetUser(ctx, did)
	if err != nil {
		return err
	}

	// 更新风险评分
	userInfo.RiskScore = newRiskScore
	
	// 根据风险评分调整访问级别和状态
	userInfo.AccessLevel = models.GetAccessLevelByRiskScore(newRiskScore)
	userInfo.Status = models.GetStatusByRiskScore(newRiskScore)
	
	// 使用交易时间戳而不是当前时间，确保确定性
	timestamp, err := ctx.GetStub().GetTxTimestamp()
	if err != nil {
		return fmt.Errorf("获取交易时间戳失败: %v", err)
	}
	userInfo.LastUpdatedAt = time.Unix(timestamp.Seconds, int64(timestamp.Nanos))

	// 将更新后的用户信息转换为JSON并存储
	userInfoJSON, err := json.Marshal(userInfo)
	if err != nil {
		return fmt.Errorf("用户信息序列化失败: %v", err)
	}

	// 将更新后的用户信息写入账本
	err = ctx.GetStub().PutState(did, userInfoJSON)
	if err != nil {
		return fmt.Errorf("更新用户信息时出错: %v", err)
	}

	return nil
}

// GetUsersByRiskScore 获取风险评分在指定范围内的用户
func (c *RiskContract) GetUsersByRiskScore(ctx contractapi.TransactionContextInterface, minScore, maxScore int) ([]*models.UserInfo, error) {
	// 检查分数范围是否有效
	if minScore < 0 || maxScore > models.RiskScoreMax || minScore > maxScore {
		return nil, fmt.Errorf("无效的风险评分范围")
	}

	// 获取所有用户
	userContract := new(UserContract)
	allUsers, err := userContract.GetAllUsers(ctx)
	if err != nil {
		return nil, err
	}

	// 筛选风险评分在指定范围内的用户
	var filteredUsers []*models.UserInfo
	for _, user := range allUsers {
		if user.RiskScore >= minScore && user.RiskScore <= maxScore {
			filteredUsers = append(filteredUsers, user)
		}
	}

	return filteredUsers, nil
}

// GetHighRiskUsers 获取高风险用户（风险评分>=70）
func (c *RiskContract) GetHighRiskUsers(ctx contractapi.TransactionContextInterface) ([]*models.UserInfo, error) {
	return c.GetUsersByRiskScore(ctx, models.RiskThresholdHigh, models.RiskScoreMax)
}
