package contracts

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/Tittifer/IEEE/sidechain/models"
)

// SessionContract 会话管理合约
type SessionContract struct {
	contractapi.Contract
}

// UpdateUserStatus 更新用户状态
// status: "online" 表示用户登入，"offline" 表示用户登出
func (c *SessionContract) UpdateUserStatus(ctx contractapi.TransactionContextInterface, did string, status string) error {
	// 验证DID是否存在
	didContract := new(DIDContract)
	exists, err := didContract.DIDExists(ctx, did)
	if err != nil {
		return fmt.Errorf("检查DID是否存在时出错: %v", err)
	}
	if !exists {
		return fmt.Errorf("DID %s 不存在", did)
	}

	// 验证状态参数
	var sessionStatus models.UserSessionStatus
	if status == "online" {
		sessionStatus = models.SessionStatusOnline
	} else if status == "offline" {
		sessionStatus = models.SessionStatusOffline
	} else {
		return fmt.Errorf("无效的用户状态: %s，应为 'online' 或 'offline'", status)
	}

	// 获取当前时间戳
	timestamp, err := ctx.GetStub().GetTxTimestamp()
	if err != nil {
		return fmt.Errorf("获取交易时间戳失败: %v", err)
	}
	currentTime := timestamp.Seconds

	// 获取用户会话
	sessionKey := getSessionKey(did)
	sessionJSON, err := ctx.GetStub().GetState(sessionKey)
	if err != nil {
		return fmt.Errorf("获取会话信息时出错: %v", err)
	}

	var session *models.UserSession
	if sessionJSON == nil {
		// 如果会话不存在，创建新会话
		session = models.NewUserSession(did, sessionStatus, currentTime)
	} else {
		// 如果会话存在，更新状态
		var existingSession models.UserSession
		err = json.Unmarshal(sessionJSON, &existingSession)
		if err != nil {
			return fmt.Errorf("反序列化会话信息时出错: %v", err)
		}
		
		// 检查状态变化
		if existingSession.Status == sessionStatus {
			return fmt.Errorf("用户 %s 已经处于 %s 状态", did, status)
		}
		
		// 更新状态
		existingSession.UpdateStatus(sessionStatus, currentTime)
		session = &existingSession
	}

	// 保存会话信息
	sessionJSON, err = json.Marshal(session)
	if err != nil {
		return fmt.Errorf("序列化会话信息时出错: %v", err)
	}

	err = ctx.GetStub().PutState(sessionKey, sessionJSON)
	if err != nil {
		return fmt.Errorf("保存会话信息时出错: %v", err)
	}

	return nil
}

// GetUserSession 获取用户会话信息
func (c *SessionContract) GetUserSession(ctx contractapi.TransactionContextInterface, did string) (*models.UserSession, error) {
	// 验证DID是否存在
	didContract := new(DIDContract)
	exists, err := didContract.DIDExists(ctx, did)
	if err != nil {
		return nil, fmt.Errorf("检查DID是否存在时出错: %v", err)
	}
	if !exists {
		return nil, fmt.Errorf("DID %s 不存在", did)
	}

	// 获取用户会话
	sessionKey := getSessionKey(did)
	sessionJSON, err := ctx.GetStub().GetState(sessionKey)
	if err != nil {
		return nil, fmt.Errorf("获取会话信息时出错: %v", err)
	}

	if sessionJSON == nil {
		return nil, fmt.Errorf("用户 %s 未登录过", did)
	}

	var session models.UserSession
	err = json.Unmarshal(sessionJSON, &session)
	if err != nil {
		return nil, fmt.Errorf("反序列化会话信息时出错: %v", err)
	}

	return &session, nil
}

// RecordRiskBehavior 记录风险行为
func (c *SessionContract) RecordRiskBehavior(ctx contractapi.TransactionContextInterface, did string, behaviorType string) error {
	// 验证DID是否存在
	didContract := new(DIDContract)
	exists, err := didContract.DIDExists(ctx, did)
	if err != nil {
		return fmt.Errorf("检查DID是否存在时出错: %v", err)
	}
	if !exists {
		return fmt.Errorf("DID %s 不存在", did)
	}

	// 获取用户会话
	sessionKey := getSessionKey(did)
	sessionJSON, err := ctx.GetStub().GetState(sessionKey)
	if err != nil {
		return fmt.Errorf("获取会话信息时出错: %v", err)
	}

	if sessionJSON == nil {
		return fmt.Errorf("用户 %s 未登录", did)
	}

	var session models.UserSession
	err = json.Unmarshal(sessionJSON, &session)
	if err != nil {
		return fmt.Errorf("反序列化会话信息时出错: %v", err)
	}

	if session.Status == models.SessionStatusOffline {
		return fmt.Errorf("用户 %s 已登出，无法记录风险行为", did)
	}

	// 验证风险行为类型
	riskBehaviorType := models.RiskBehaviorType(behaviorType)
	if _, exists := models.RiskBehaviorScaleFactor[riskBehaviorType]; !exists {
		return fmt.Errorf("无效的风险行为类型: %s", behaviorType)
	}

	// 增加严重程度值
	session.IncrementSeverityLevel()
	
	// 保存更新后的会话信息
	sessionJSON, err = json.Marshal(session)
	if err != nil {
		return fmt.Errorf("序列化会话信息时出错: %v", err)
	}

	err = ctx.GetStub().PutState(sessionKey, sessionJSON)
	if err != nil {
		return fmt.Errorf("保存会话信息时出错: %v", err)
	}

	return nil
}

// 生成会话键
func getSessionKey(did string) string {
	return "session_" + did
}