package connection

import (
	"fmt"
	"log"

	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
	"github.com/Tittifer/IEEE/transmit/config"
)

// SidechainConnection 侧链连接
type SidechainConnection struct {
	// 配置信息
	config *config.NetworkConfig
	// 网关
	gateway *client.Gateway
	// 网络
	network *client.Network
	// 合约
	contract *client.Contract
}

// NewSidechainConnection 创建侧链连接
func NewSidechainConnection(config *config.NetworkConfig) (*SidechainConnection, error) {
	conn := &SidechainConnection{
		config: config,
	}
	
	// 连接到Fabric网络
	err := conn.connect()
	if err != nil {
		return nil, fmt.Errorf("连接侧链失败: %v", err)
	}
	
	return conn, nil
}

// 连接到Fabric网络
func (c *SidechainConnection) connect() error {
	// 读取证书
	certificate, err := loadCertificate(c.config.CertPath)
	if err != nil {
		return err
	}
	
	// 读取私钥
	privateKey, err := loadPrivateKey(c.config.KeyPath)
	if err != nil {
		return err
	}
	
	// 创建身份
	id, err := identity.NewX509Identity(c.config.MspID, certificate)
	if err != nil {
		return fmt.Errorf("创建身份失败: %v", err)
	}
	
	// 创建签名函数
	sign, err := identity.NewPrivateKeySign(privateKey)
	if err != nil {
		return fmt.Errorf("创建签名函数失败: %v", err)
	}
	
	// 创建TLS凭证
	clientConnection, err := newGrpcConnection(c.config.PeerEndpoint, c.config.TlsCertPath)
	if err != nil {
		return fmt.Errorf("创建gRPC连接失败: %v", err)
	}
	
	// 创建网关
	gateway, err := client.Connect(
		id,
		client.WithClientConnection(clientConnection),
		client.WithSign(sign),
	)
	if err != nil {
		return fmt.Errorf("创建网关失败: %v", err)
	}
	
	// 获取网络
	network := gateway.GetNetwork(c.config.ChannelName)
	
	// 获取合约
	contract := network.GetContract(c.config.ChaincodeName)
	
	// 保存连接信息
	c.gateway = gateway
	c.network = network
	c.contract = contract
	
	log.Printf("成功连接到侧链: %s\n", c.config.ChannelName)
	
	return nil
}

// Close 关闭连接
func (c *SidechainConnection) Close() {
	if c.gateway != nil {
		c.gateway.Close()
		log.Println("侧链连接已关闭")
	}
}

// GetContract 获取合约
func (c *SidechainConnection) GetContract() *client.Contract {
	return c.contract
}