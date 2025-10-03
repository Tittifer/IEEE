package client

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

// ConnectionConfig 连接配置
type ConnectionConfig struct {
	MSPID        string `json:"mspID"`
	CryptoPath   string `json:"cryptoPath"`
	CertPath     string `json:"certPath"`
	KeyPath      string `json:"keyPath"`
	TLSCertPath  string `json:"tlsCertPath"`
	PeerEndpoint string `json:"peerEndpoint"`
	GatewayPeer  string `json:"gatewayPeer"`
	ChannelName  string `json:"channelName"`
	ChaincodeName string `json:"chaincodeName"`
}

// LoadConfig 从文件加载配置
func LoadConfig(configPath string) (*ConnectionConfig, error) {
	// 如果配置文件不存在，则创建默认配置
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// 确保目录存在
		dir := filepath.Dir(configPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("创建配置目录失败: %w", err)
		}

		// 创建默认配置
		defaultConfig := &ConnectionConfig{
			MSPID:        "Org1MSP",
			CryptoPath:   "../chain_docker/crypto-config/peerOrganizations/org1.chain.com",
			CertPath:     "../chain_docker/crypto-config/peerOrganizations/org1.chain.com/users/User1@org1.chain.com/msp/signcerts/User1@org1.chain.com-cert.pem",
			KeyPath:      "../chain_docker/crypto-config/peerOrganizations/org1.chain.com/users/User1@org1.chain.com/msp/keystore/",
			TLSCertPath:  "../chain_docker/crypto-config/peerOrganizations/org1.chain.com/peers/peer0.org1.chain.com/tls/ca.crt",
			PeerEndpoint: "localhost:8051",
			GatewayPeer:  "peer0.org1.chain.com",
			ChannelName:  "mainchannel",
			ChaincodeName: "chaincc",
		}

		// 将默认配置写入文件
		configJSON, err := json.MarshalIndent(defaultConfig, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("序列化配置失败: %w", err)
		}

		if err := ioutil.WriteFile(configPath, configJSON, 0644); err != nil {
			return nil, fmt.Errorf("写入配置文件失败: %w", err)
		}

		return defaultConfig, nil
	}

	// 读取配置文件
	configJSON, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	// 解析配置
	var config ConnectionConfig
	if err := json.Unmarshal(configJSON, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	return &config, nil
}
