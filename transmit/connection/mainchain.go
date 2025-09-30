package connection

import (
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
	"github.com/Tittifer/IEEE/transmit/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// MainchainConnection 主链连接
type MainchainConnection struct {
	// 配置信息
	config *config.NetworkConfig
	// 网关
	gateway *client.Gateway
	// 网络
	network *client.Network
	// 合约
	contract *client.Contract
}

// NewMainchainConnection 创建主链连接
func NewMainchainConnection(config *config.NetworkConfig) (*MainchainConnection, error) {
	conn := &MainchainConnection{
		config: config,
	}
	
	// 连接到Fabric网络
	err := conn.connect()
	if err != nil {
		return nil, fmt.Errorf("连接主链失败: %v", err)
	}
	
	return conn, nil
}

// 连接到Fabric网络
func (c *MainchainConnection) connect() error {
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
	
	log.Printf("成功连接到主链: %s\n", c.config.ChannelName)
	
	return nil
}

// Close 关闭连接
func (c *MainchainConnection) Close() {
	if c.gateway != nil {
		c.gateway.Close()
		log.Println("主链连接已关闭")
	}
}

// GetContract 获取合约
func (c *MainchainConnection) GetContract() *client.Contract {
	return c.contract
}

// 加载证书
func loadCertificate(certPath string) (*x509.Certificate, error) {
	certPEM, err := ioutil.ReadFile(certPath)
	if err != nil {
		return nil, fmt.Errorf("读取证书文件失败: %v", err)
	}
	
	cert, err := identity.CertificateFromPEM(certPEM)
	if err != nil {
		return nil, fmt.Errorf("解析证书失败: %v", err)
	}
	
	return cert, nil
}

// 加载私钥
func loadPrivateKey(keyPath string) (interface{}, error) {
	keyPEM, err := ioutil.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("读取私钥文件失败: %v", err)
	}
	
	key, err := identity.PrivateKeyFromPEM(keyPEM)
	if err != nil {
		return nil, fmt.Errorf("解析私钥失败: %v", err)
	}
	
	return key, nil
}

// 创建gRPC连接
func newGrpcConnection(peerEndpoint string, tlsCertPath string) (*grpc.ClientConn, error) {
	certificate, err := loadTLSCertificate(tlsCertPath)
	if err != nil {
		return nil, err
	}
	
	certPool := x509.NewCertPool()
	certPool.AddCert(certificate)
	transportCredentials := credentials.NewClientTLSFromCert(certPool, "")
	
	connection, err := grpc.Dial(peerEndpoint, grpc.WithTransportCredentials(transportCredentials))
	if err != nil {
		return nil, fmt.Errorf("创建gRPC连接失败: %v", err)
	}
	
	return connection, nil
}

// 加载TLS证书
func loadTLSCertificate(tlsCertPath string) (*x509.Certificate, error) {
	certPEM, err := ioutil.ReadFile(tlsCertPath)
	if err != nil {
		return nil, fmt.Errorf("读取TLS证书文件失败: %v", err)
	}
	
	cert, err := identity.CertificateFromPEM(certPEM)
	if err != nil {
		return nil, fmt.Errorf("解析TLS证书失败: %v", err)
	}
	
	return cert, nil
}