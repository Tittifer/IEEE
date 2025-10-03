package contracts

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/Tittifer/IEEE/chain/models"
	"github.com/Tittifer/IEEE/chain/utils"
)

// RiskContract 风险管理合约
type RiskContract struct {
	contractapi.Contract
}

// 使用models包中定义的风险评分阈值常量

// InitRiskLedger 初始化风险管理账本
func (c *RiskContract) InitRiskLedger(ctx contractapi.TransactionContextInterface) error {
	fmt.Println("风险管理链码初始化")
	return nil
}

// UpdateRiskScore 更新用户风险评分
func (c *RiskContract) UpdateRiskScore(ctx contractapi.TransactionContextInterface, did string, newScoreStr string) error {
	// 验证DID格式
	if !utils.ValidateDID(did) {
		return fmt.Errorf("无效的DID格式: %s", did)
	}
	
	// 将字符串转换为整数
	newScore, err := strconv.Atoi(newScoreStr)
	if err != nil {
		return fmt.Errorf("风险评分格式无效: %v", err)
	}
	
	// 验证风险评分范围
	if newScore < 0 || newScore > 100 {
		return fmt.Errorf("风险评分必须在0-100之间")
	}
	
	// 从账本中读取用户信息
	userInfoJSON, err := ctx.GetStub().GetState(did)
	if err != nil {
		return fmt.Errorf("读取用户信息时出错: %v", err)
	}
	if userInfoJSON == nil {
		return fmt.Errorf("用户DID %s 不存在", did)
	}
	
	// 反序列化用户信息
	var userInfo models.UserInfo
	err = json.Unmarshal(userInfoJSON, &userInfo)
	if err != nil {
		return fmt.Errorf("用户信息反序列化失败: %v", err)
	}
	
	// 更新风险评分
	userInfo.RiskScore = newScore
	
	// 将用户信息转换为JSON并存储
	updatedUserInfoJSON, err := json.Marshal(userInfo)
	if err != nil {
		return fmt.Errorf("用户信息序列化失败: %v", err)
	}
	
	// 将更新后的用户信息写入账本
	err = ctx.GetStub().PutState(did, updatedUserInfoJSON)
	if err != nil {
		return fmt.Errorf("存储用户信息时出错: %v", err)
	}
	
	return nil
}

// GetRiskScore 获取用户风险评分
func (c *RiskContract) GetRiskScore(ctx contractapi.TransactionContextInterface, did string) (int, error) {
	// 验证DID格式
	if !utils.ValidateDID(did) {
		return 0, fmt.Errorf("无效的DID格式: %s", did)
	}
	
	// 从账本中读取用户信息
	userInfoJSON, err := ctx.GetStub().GetState(did)
	if err != nil {
		return 0, fmt.Errorf("读取用户信息时出错: %v", err)
	}
	if userInfoJSON == nil {
		return 0, fmt.Errorf("用户DID %s 不存在", did)
	}
	
	// 反序列化用户信息
	var userInfo models.UserInfo
	err = json.Unmarshal(userInfoJSON, &userInfo)
	if err != nil {
		return 0, fmt.Errorf("用户信息反序列化失败: %v", err)
	}
	
	return userInfo.RiskScore, nil
}

// CheckUserLoginEligibility 检查用户是否有资格登录
func (c *RiskContract) CheckUserLoginEligibility(ctx contractapi.TransactionContextInterface, did string) (string, error) {
	// 获取用户风险评分
	riskScore, err := c.GetRiskScore(ctx, did)
	if err != nil {
		return "", err
	}
	
	// 检查风险评分是否超过阈值
	if riskScore > models.RiskScoreThreshold {
		return "禁止用户登录", nil
	}
	
	return "用户可以登录", nil
}

// GetHighRiskUsers 获取高风险用户
func (c *RiskContract) GetHighRiskUsers(ctx contractapi.TransactionContextInterface) ([]*models.UserInfo, error) {
	// 获取所有用户的迭代器
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, fmt.Errorf("获取状态范围时出错: %v", err)
	}
	defer resultsIterator.Close()
	
	// 存储高风险用户的切片
	var highRiskUsers []*models.UserInfo
	
	// 遍历所有状态
	for resultsIterator.HasNext() {
		// 获取下一个键值对
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, fmt.Errorf("获取下一个状态时出错: %v", err)
		}
		
		// 反序列化用户信息
		var userInfo models.UserInfo
		err = json.Unmarshal(queryResponse.Value, &userInfo)
		if err != nil {
			// 跳过非用户信息的状态
			continue
		}
		
		// 检查用户是否为高风险用户
		if userInfo.RiskScore > models.RiskScoreThreshold {
			highRiskUsers = append(highRiskUsers, &userInfo)
		}
	}
	
	return highRiskUsers, nil
}

// GetUsersByRiskScoreRange 获取特定风险评分范围内的用户
func (c *RiskContract) GetUsersByRiskScoreRange(ctx contractapi.TransactionContextInterface, minScoreStr, maxScoreStr string) ([]*models.UserInfo, error) {
	// 将字符串转换为整数
	minScore, err := strconv.Atoi(minScoreStr)
	if err != nil {
		return nil, fmt.Errorf("最小风险评分格式无效: %v", err)
	}
	
	maxScore, err := strconv.Atoi(maxScoreStr)
	if err != nil {
		return nil, fmt.Errorf("最大风险评分格式无效: %v", err)
	}
	
	// 验证风险评分范围
	if minScore < 0 || minScore > 100 || maxScore < 0 || maxScore > 100 || minScore > maxScore {
		return nil, fmt.Errorf("风险评分范围无效")
	}
	
	// 获取所有用户的迭代器
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, fmt.Errorf("获取状态范围时出错: %v", err)
	}
	defer resultsIterator.Close()
	
	// 存储符合条件的用户的切片
	var users []*models.UserInfo
	
	// 遍历所有状态
	for resultsIterator.HasNext() {
		// 获取下一个键值对
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, fmt.Errorf("获取下一个状态时出错: %v", err)
		}
		
		// 反序列化用户信息
		var userInfo models.UserInfo
		err = json.Unmarshal(queryResponse.Value, &userInfo)
		if err != nil {
			// 跳过非用户信息的状态
			continue
		}
		
		// 检查用户是否在指定风险评分范围内
		if userInfo.RiskScore >= minScore && userInfo.RiskScore <= maxScore {
			users = append(users, &userInfo)
		}
	}
	
	return users, nil
}
