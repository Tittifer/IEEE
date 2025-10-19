package client

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/Tittifer/IEEE/honeypoint_client/chain"
)

// ChainClient 区块链客户端，实现chain.ChainClient接口
type ChainClient struct {
	honeypointClient *HoneypointClient
}

// NewChainClient 创建新的区块链客户端
func NewChainClient(honeypointClient *HoneypointClient) *ChainClient {
	return &ChainClient{
		honeypointClient: honeypointClient,
	}
}

// GetDeviceInfo 从区块链获取设备信息
func (c *ChainClient) GetDeviceInfo(did string) (*chain.Device, error) {
	// 调用链码获取设备信息
	deviceJSON, err := c.honeypointClient.contract.EvaluateTransaction("IdentityContract:GetDevice", did)
	if err != nil {
		return nil, fmt.Errorf("评估交易失败: %w", err)
	}
	if len(deviceJSON) == 0 {
		return nil, fmt.Errorf("设备DID %s 不存在", did)
	}

	// 解析设备信息
	var deviceInfo struct {
		DID           string    `json:"did"`
		Name          string    `json:"name"`
		Model         string    `json:"model"`
		Vendor        string    `json:"vendor"`
		RiskScore     float64   `json:"riskScore"`
		AttackIndexI  float64   `json:"attackIndexI"`
		AttackProfile []string  `json:"attackProfile"`
		LastEventTime time.Time `json:"lastEventTime"`
		Status        string    `json:"status"`
		CreatedAt     time.Time `json:"createdAt"`
		LastUpdatedAt time.Time `json:"lastUpdatedAt"`
	}
	if err := json.Unmarshal(deviceJSON, &deviceInfo); err != nil {
		return nil, fmt.Errorf("设备信息解析失败: %w", err)
	}

	// 转换为Device结构体
	device := &chain.Device{
		DID:           deviceInfo.DID,
		Name:          deviceInfo.Name,
		Model:         deviceInfo.Model,
		Vendor:        deviceInfo.Vendor,
		RiskScore:     deviceInfo.RiskScore,
		AttackIndexI:  deviceInfo.AttackIndexI,
		AttackProfile: deviceInfo.AttackProfile,
		LastEventTime: deviceInfo.LastEventTime,
		Status:        deviceInfo.Status,
		CreatedAt:     deviceInfo.CreatedAt,
		LastUpdatedAt: deviceInfo.LastUpdatedAt,
	}

	return device, nil
}

// UpdateDeviceRiskScore 更新设备风险评分
func (c *ChainClient) UpdateDeviceRiskScore(did string, riskScore float64, attackIndexI float64, attackProfile []string) error {
	// 将浮点数转换为字符串
	riskScoreStr := fmt.Sprintf("%.2f", riskScore)
	attackIndexStr := fmt.Sprintf("%.2f", attackIndexI)

	// 序列化攻击画像
	attackProfileJSON, err := json.Marshal(attackProfile)
	if err != nil {
		return fmt.Errorf("攻击画像序列化失败: %w", err)
	}

	// 调用链码更新设备风险评分
	_, err = c.honeypointClient.contract.SubmitTransaction(
		"IdentityContract:UpdateDeviceRiskScore",
		did,
		riskScoreStr,
		attackIndexStr,
		string(attackProfileJSON),
	)
	if err != nil {
		return fmt.Errorf("提交交易失败: %w", err)
	}

	log.Printf("已更新设备 %s 的风险评分为 %.2f，攻击画像指数为 %.2f", did, riskScore, attackIndexI)
	return nil
}
