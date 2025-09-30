package handlers

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/Tittifer/IEEE/transmit/models"
)

// SidechainHandler 侧链处理器
type SidechainHandler struct {
	// 合约
	contract *client.Contract
}

// NewSidechainHandler 创建侧链处理器
func NewSidechainHandler(contract *client.Contract) *SidechainHandler {
	return &SidechainHandler{
		contract: contract,
	}
}

// CreateDIDRecord 创建DID记录
func (h *SidechainHandler) CreateDIDRecord(did string) error {
	log.Printf("创建DID记录: %s\n", did)
	
	// 调用链码
	_, err := h.contract.SubmitTransaction("DIDContract:CreateDIDRecord", did)
	if err != nil {
		return fmt.Errorf("创建DID记录失败: %v", err)
	}
	
	log.Printf("成功创建DID记录: %s", did)
	return nil
}

// GetDIDRecord 获取DID记录
func (h *SidechainHandler) GetDIDRecord(did string) (*models.DIDRiskRecord, error) {
	log.Printf("获取DID记录: %s\n", did)
	
	// 调用链码
	result, err := h.contract.EvaluateTransaction("DIDContract:GetDIDRecord", did)
	if err != nil {
		return nil, fmt.Errorf("获取DID记录失败: %v", err)
	}
	
	// 解析结果
	var record models.DIDRiskRecord
	err = json.Unmarshal(result, &record)
	if err != nil {
		return nil, fmt.Errorf("解析DID记录失败: %v", err)
	}
	
	return &record, nil
}

// UpdateUserStatus 更新用户状态
func (h *SidechainHandler) UpdateUserStatus(did, status string) error {
	log.Printf("更新用户状态: %s, %s\n", did, status)
	
	// 调用链码
	_, err := h.contract.SubmitTransaction("SessionContract:UpdateUserStatus", did, status)
	if err != nil {
		return fmt.Errorf("更新用户状态失败: %v", err)
	}
	
	log.Printf("成功更新用户 %s 状态为 %s", did, status)
	return nil
}

// GetUserSession 获取用户会话信息
func (h *SidechainHandler) GetUserSession(did string) (*models.UserSession, error) {
	log.Printf("获取用户会话信息: %s\n", did)
	
	// 调用链码
	result, err := h.contract.EvaluateTransaction("SessionContract:GetUserSession", did)
	if err != nil {
		return nil, fmt.Errorf("获取用户会话信息失败: %v", err)
	}
	
	// 解析结果
	var session models.UserSession
	err = json.Unmarshal(result, &session)
	if err != nil {
		return nil, fmt.Errorf("解析用户会话信息失败: %v", err)
	}
	
	return &session, nil
}

// CheckRiskThreshold 检查风险阈值
func (h *SidechainHandler) CheckRiskThreshold(did string) (bool, error) {
	log.Printf("检查风险阈值: %s\n", did)
	
	// 调用链码
	result, err := h.contract.EvaluateTransaction("RiskContract:CheckRiskThreshold", did)
	if err != nil {
		return false, fmt.Errorf("检查风险阈值失败: %v", err)
	}
	
	return string(result) == "true", nil
}

// GetAllDIDRecords 获取所有DID记录
func (h *SidechainHandler) GetAllDIDRecords() ([]*models.DIDRiskRecord, error) {
	log.Println("获取所有DID记录")
	
	// 调用链码
	result, err := h.contract.EvaluateTransaction("DIDContract:GetAllDIDRecords")
	if err != nil {
		return nil, fmt.Errorf("获取所有DID记录失败: %v", err)
	}
	
	// 解析结果
	var records []*models.DIDRiskRecord
	err = json.Unmarshal(result, &records)
	if err != nil {
		return nil, fmt.Errorf("解析DID记录列表失败: %v", err)
	}
	
	return records, nil
}