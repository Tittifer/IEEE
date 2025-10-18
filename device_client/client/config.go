package client

import (
	"encoding/json"
	"fmt"
	"os"
)

// ConnectionConfig 连接配置结构体
type ConnectionConfig struct {
	MSPID        string `json:"mspid"`
	CertPath     string `json:"certPath"`
	KeyPath      string `json:"keyPath"`
	TLSCertPath  string `json:"tlsCertPath"`
	PeerEndpoint string `json:"peerEndpoint"`
	GatewayPeer  string `json:"gatewayPeer"`
	ChannelName  string `json:"channelName"`
	ChaincodeName string `json:"chaincodeName"`
}

// LoadConfig 加载配置
func LoadConfig(configPath string) (*ConnectionConfig, error) {
	// 读取配置文件
	configFile, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}
	
	// 解析配置
	var config ConnectionConfig
	err = json.Unmarshal(configFile, &config)
	if err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}
	
	return &config, nil
}
