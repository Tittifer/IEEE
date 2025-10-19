package chain

import (
	"fmt"
	"log"
	"time"
)

// ChainManager 区块链管理器
type ChainManager struct {
	chainClient ChainClient
}

// Device 设备结构体
type Device struct {
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

// ChainClient 区块链客户端接口
type ChainClient interface {
	GetDeviceInfo(did string) (*Device, error)
	UpdateDeviceRiskScore(did string, riskScore float64, attackIndexI float64, attackProfile []string) error
}

// NewChainManager 创建新的区块链管理器
func NewChainManager(chainClient ChainClient) *ChainManager {
	return &ChainManager{
		chainClient: chainClient,
	}
}

// GetDeviceFromChain 从区块链获取设备信息
func (m *ChainManager) GetDeviceFromChain(did string) (*Device, error) {
	return m.chainClient.GetDeviceInfo(did)
}

// UpdateDeviceAttackIndex 更新设备攻击画像指数
func (m *ChainManager) UpdateDeviceAttackIndex(did string, attackIndexI float64) error {
	// 先从链上获取设备信息
	device, err := m.GetDeviceFromChain(did)
	if err != nil {
		return fmt.Errorf("获取设备信息失败: %w", err)
	}
	
	// 更新攻击画像指数
	return m.chainClient.UpdateDeviceRiskScore(did, device.RiskScore, attackIndexI, device.AttackProfile)
}

// ResetDeviceRiskData 重置设备风险数据
func (m *ChainManager) ResetDeviceRiskData(did string) error {
	// 重置链上设备风险评分和攻击画像
	emptyProfile := make([]string, 0)
	err := m.chainClient.UpdateDeviceRiskScore(did, 0.0, 0.0, emptyProfile)
	if err != nil {
		return fmt.Errorf("重置设备风险评分失败: %w", err)
	}
	
	log.Printf("已成功重置设备 %s 的风险数据", did)
	return nil
}
