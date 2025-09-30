package contracts

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/Tittifer/IEEE/sidechain/models"
	"github.com/Tittifer/IEEE/sidechain/utils"
)

// RiskContract 风险评估合约
type RiskContract struct {
	contractapi.Contract
}

// InitLedger 初始化账本
func (c *RiskContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	fmt.Println("风险评估链码初始化")
	return nil
}

// UpdateRiskScore 更新用户风险评分
func (c *RiskContract) UpdateRiskScore(ctx contractapi.TransactionContextInterface, did string, behaviorType string) error {
	// 验证DID是否存在
	didContract := new(DIDContract)
	exists, err := didContract.DIDExists(ctx, did)
	if err != nil {
		return fmt.Errorf("检查DID是否存在时出错: %v", err)
	}
	if !exists {
		return fmt.Errorf("DID %s 不存在", did)
	}

	// 验证风险行为类型
	riskBehaviorType := models.RiskBehaviorType(behaviorType)
	if _, exists := models.RiskBehaviorScaleFactor[riskBehaviorType]; !exists {
		return fmt.Errorf("无效的风险行为类型: %s", behaviorType)
	}

	// 获取用户会话信息
	sessionContract := new(SessionContract)
	session, err := sessionContract.GetUserSession(ctx, did)
	if err != nil {
		return fmt.Errorf("获取用户会话信息时出错: %v", err)
	}

	if session.Status == models.SessionStatusOffline {
		return fmt.Errorf("用户 %s 已登出，无法更新风险评分", did)
	}

	// 获取当前DID记录
	didRecord, err := didContract.GetDIDRecord(ctx, did)
	if err != nil {
		return fmt.Errorf("获取DID记录时出错: %v", err)
	}

	// 获取当前时间戳
	timestamp, err := ctx.GetStub().GetTxTimestamp()
	if err != nil {
		return fmt.Errorf("获取交易时间戳失败: %v", err)
	}
	currentTime := timestamp.Seconds

	// 计算新的风险评分
	newRiskScore := utils.CalculateRiskScore(
		didRecord.RiskScore,
		riskBehaviorType,
		session.SeverityLevel,
		didRecord.Timestamp,
		currentTime,
	)

	// 更新DID记录
	didRecord.RiskScore = newRiskScore
	
	// 只有在用户当前会话中首次触发风险行为时才更新时间戳
	// 通过检查session.HasTriggeredRisk是否为false来判断是否为首次触发
	if !session.HasTriggeredRisk {
		didRecord.Timestamp = currentTime
	}

	// 将更新后的记录保存到账本
	didRecordJSON, err := json.Marshal(didRecord)
	if err != nil {
		return fmt.Errorf("序列化DID记录时出错: %v", err)
	}

	err = ctx.GetStub().PutState(did, didRecordJSON)
	if err != nil {
		return fmt.Errorf("保存DID记录时出错: %v", err)
	}

	return nil
}

// EvaluateRiskScore 评估用户当前风险评分（不更新状态）
func (c *RiskContract) EvaluateRiskScore(ctx contractapi.TransactionContextInterface, did string) (int, error) {
	// 验证DID是否存在
	didContract := new(DIDContract)
	exists, err := didContract.DIDExists(ctx, did)
	if err != nil {
		return 0, fmt.Errorf("检查DID是否存在时出错: %v", err)
	}
	if !exists {
		return 0, fmt.Errorf("DID %s 不存在", did)
	}

	// 获取当前DID记录
	didRecord, err := didContract.GetDIDRecord(ctx, did)
	if err != nil {
		return 0, fmt.Errorf("获取DID记录时出错: %v", err)
	}

	// 获取当前时间戳
	timestamp, err := ctx.GetStub().GetTxTimestamp()
	if err != nil {
		return 0, fmt.Errorf("获取交易时间戳失败: %v", err)
	}
	currentTime := timestamp.Seconds

	// 计算衰减后的风险评分
	decayedScore := utils.CalculateScoreDecay(
		didRecord.RiskScore,
		didRecord.Timestamp,
		currentTime,
	)

	return decayedScore, nil
}

// ReportRiskBehavior 报告风险行为并更新评分
func (c *RiskContract) ReportRiskBehavior(ctx contractapi.TransactionContextInterface, did string, behaviorType string) error {
	// 验证DID是否存在
	didContract := new(DIDContract)
	exists, err := didContract.DIDExists(ctx, did)
	if err != nil {
		return fmt.Errorf("检查DID是否存在时出错: %v", err)
	}
	if !exists {
		return fmt.Errorf("DID %s 不存在", did)
	}

	// 记录风险行为
	sessionContract := new(SessionContract)
	err = sessionContract.RecordRiskBehavior(ctx, did, behaviorType)
	if err != nil {
		return fmt.Errorf("记录风险行为时出错: %v", err)
	}

	// 更新风险评分
	err = c.UpdateRiskScore(ctx, did, behaviorType)
	if err != nil {
		return fmt.Errorf("更新风险评分时出错: %v", err)
	}

	return nil
}

// CheckRiskThreshold 检查用户风险评分是否超过阈值
func (c *RiskContract) CheckRiskThreshold(ctx contractapi.TransactionContextInterface, did string) (bool, error) {
	// 评估当前风险评分
	currentScore, err := c.EvaluateRiskScore(ctx, did)
	if err != nil {
		return false, err
	}

	// 检查是否超过阈值
	return currentScore > models.RiskScoreThreshold, nil
}