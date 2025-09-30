package handlers

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/Tittifer/IEEE/transmit/models"
)

// MainchainHandler 主链处理器
type MainchainHandler struct {
	// 合约
	contract *client.Contract
}

// NewMainchainHandler 创建主链处理器
func NewMainchainHandler(contract *client.Contract) *MainchainHandler {
	return &MainchainHandler{
		contract: contract,
	}
}

// GetUser 获取用户信息
func (h *MainchainHandler) GetUser(did string) (*models.UserInfo, error) {
	log.Printf("获取用户信息: %s\n", did)
	
	// 调用链码
	result, err := h.contract.EvaluateTransaction("IdentityContract:GetUser", did)
	if err != nil {
		return nil, fmt.Errorf("获取用户信息失败: %v", err)
	}
	
	// 解析结果
	var userInfo models.UserInfo
	err = json.Unmarshal(result, &userInfo)
	if err != nil {
		return nil, fmt.Errorf("解析用户信息失败: %v", err)
	}
	
	return &userInfo, nil
}

// GetAllUsers 获取所有用户
func (h *MainchainHandler) GetAllUsers() ([]*models.UserInfo, error) {
	log.Println("获取所有用户")
	
	// 调用链码
	result, err := h.contract.EvaluateTransaction("IdentityContract:GetAllUsers")
	if err != nil {
		return nil, fmt.Errorf("获取所有用户失败: %v", err)
	}
	
	// 解析结果
	var users []*models.UserInfo
	err = json.Unmarshal(result, &users)
	if err != nil {
		return nil, fmt.Errorf("解析用户列表失败: %v", err)
	}
	
	return users, nil
}

// UpdateUserStatus 更新用户状态
func (h *MainchainHandler) UpdateUserStatus(did string, status string) error {
	log.Printf("更新用户状态: %s -> %s\n", did, status)
	
	// 获取用户信息
	userInfo, err := h.GetUser(did)
	if err != nil {
		return fmt.Errorf("获取用户信息失败: %v", err)
	}
	
	// 如果状态已经是目标状态，则不需要更新
	if userInfo.Status == status {
		log.Printf("用户 %s 状态已经是 %s，无需更新", did, status)
		return nil
	}
	
	// 调用链码更新用户状态
	// 注意：这里假设主链有UpdateUserStatus方法，如果没有，需要根据实际情况调整
	_, err = h.contract.SubmitTransaction("IdentityContract:UpdateUserStatus", did, status)
	if err != nil {
		return fmt.Errorf("更新用户状态失败: %v", err)
	}
	
	log.Printf("成功更新用户 %s 状态为 %s", did, status)
	return nil
}

// UserLogin 用户登录
func (h *MainchainHandler) UserLogin(did string, name string) (string, error) {
	log.Printf("用户登录: %s, %s\n", did, name)
	
	// 调用链码
	result, err := h.contract.SubmitTransaction("IdentityContract:UserLogin", did, name)
	if err != nil {
		return "", fmt.Errorf("用户登录失败: %v", err)
	}
	
	response := string(result)
	log.Printf("用户登录结果: %s", response)
	return response, nil
}

// UserLogout 用户登出
func (h *MainchainHandler) UserLogout(did string) (string, error) {
	log.Printf("用户登出: %s\n", did)
	
	// 调用链码
	result, err := h.contract.SubmitTransaction("IdentityContract:UserLogout", did)
	if err != nil {
		return "", fmt.Errorf("用户登出失败: %v", err)
	}
	
	response := string(result)
	log.Printf("用户登出结果: %s", response)
	return response, nil
}

// UpdateRiskScore 更新用户风险评分
func (h *MainchainHandler) UpdateRiskScore(did string, riskScore int) error {
	log.Printf("更新用户风险评分: %s -> %d\n", did, riskScore)
	
	// 获取用户信息
	userInfo, err := h.GetUser(did)
	if err != nil {
		return fmt.Errorf("获取用户信息失败: %v", err)
	}
	
	// 如果风险评分没有变化，则不需要更新
	if userInfo.RiskScore == riskScore {
		log.Printf("用户 %s 风险评分已经是 %d，无需更新", did, riskScore)
		return nil
	}
	
	// 调用链码更新用户风险评分
	// 注意：这里假设主链有UpdateRiskScore方法，如果没有，需要根据实际情况调整
	_, err = h.contract.SubmitTransaction("IdentityContract:UpdateRiskScore", did, fmt.Sprintf("%d", riskScore))
	if err != nil {
		return fmt.Errorf("更新用户风险评分失败: %v", err)
	}
	
	log.Printf("成功更新用户 %s 风险评分为 %d", did, riskScore)
	return nil
}