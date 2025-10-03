package client

import (
	"bytes"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"path"
	"time"

	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// UserClient 用户客户端结构体
type UserClient struct {
	contract *client.Contract
	gateway  *client.Gateway
	conn     *grpc.ClientConn
	config   *ConnectionConfig
}

// 用户事件结构体，用于解析链码事件
type UserEvent struct {
	EventType string `json:"eventType"` // 事件类型
	DID       string `json:"did"`       // 用户DID
	Name      string `json:"name"`      // 用户名称
	Timestamp int64  `json:"timestamp"` // 事件时间戳
	RiskScore int    `json:"riskScore"` // 风险评分（可选）
}

// 合约名称常量
const (
	identityContract = "IdentityContract"
	configPath       = "../config.json"
)

// NewUserClient 创建新的用户客户端
func NewUserClient() (*UserClient, error) {
	// 加载配置
	config, err := LoadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("加载配置失败: %w", err)
	}

	// 创建客户端证书
	clientCert, err := loadCertificate(config.CertPath)
	if err != nil {
		return nil, fmt.Errorf("加载客户端证书失败: %w", err)
	}

	// 加载客户端私钥
	clientKey, err := loadPrivateKey(config.KeyPath)
	if err != nil {
		return nil, fmt.Errorf("加载客户端私钥失败: %w", err)
	}

	// 创建身份
	id, err := identity.NewX509Identity(config.MSPID, clientCert)
	if err != nil {
		return nil, fmt.Errorf("创建X509身份失败: %w", err)
	}

	// 加载TLS证书
	tlsCert, err := loadCertificate(config.TLSCertPath)
	if err != nil {
		return nil, fmt.Errorf("加载TLS证书失败: %w", err)
	}
	
	// 创建TLS证书池
	certPool := x509.NewCertPool()
	certPool.AddCert(tlsCert)
	
	// 创建TLS凭证
	transportCredentials := credentials.NewClientTLSFromCert(certPool, config.GatewayPeer)
	
	// 创建gRPC连接
	conn, err := grpc.Dial(config.PeerEndpoint, grpc.WithTransportCredentials(transportCredentials))
	if err != nil {
		return nil, fmt.Errorf("创建gRPC连接失败: %w", err)
	}
	
	// 创建签名函数
	sign, err := identity.NewPrivateKeySign(clientKey)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("创建签名函数失败: %w", err)
	}
	
	// 创建Gateway连接
	gw, err := client.Connect(
		id,
		client.WithSign(sign),
		client.WithClientConnection(conn),
		client.WithEvaluateTimeout(5*time.Second),
		client.WithEndorseTimeout(15*time.Second),
		client.WithSubmitTimeout(5*time.Second),
		client.WithCommitStatusTimeout(1*time.Minute),
	)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("创建Gateway连接失败: %w", err)
	}
	
	// 获取网络
	network := gw.GetNetwork(config.ChannelName)
	
	// 获取智能合约
	contract := network.GetContract(config.ChaincodeName)
	
	// 创建用户客户端
	userClient := &UserClient{
		contract: contract,
		gateway:  gw,
		conn:     conn,
		config:   config,
	}
	
	return userClient, nil
}

// Close 关闭连接
func (c *UserClient) Close() {
	c.gateway.Close()
	c.conn.Close()
}

// RegisterUser 注册新用户
func (c *UserClient) RegisterUser(name, idNumber, phoneNumber, vehicleID string) (string, error) {
	log.Printf("注册用户: %s, %s, %s, %s", name, idNumber, phoneNumber, vehicleID)
	
	// 参数验证
	if name == "" || idNumber == "" || phoneNumber == "" || vehicleID == "" {
		return "", fmt.Errorf("所有参数都不能为空")
	}
	
	// 先生成DID，检查用户是否已存在
	did, err := c.GetDIDByInfo(name, idNumber, phoneNumber, vehicleID)
	if err != nil {
		return "", fmt.Errorf("生成DID失败: %w", err)
	}
	
	// 检查用户是否已存在
	userJSON, err := c.contract.EvaluateTransaction(identityContract+":GetUser", did)
	if err == nil && len(userJSON) > 0 {
		// 用户已存在
		return "", fmt.Errorf("用户已存在，DID: %s", did)
	}
	
	// 调用链码注册用户
	_, err = c.contract.SubmitTransaction(identityContract+":RegisterUser", name, idNumber, phoneNumber, vehicleID)
	if err != nil {
		return "", fmt.Errorf("提交交易失败: %w", err)
	}
	
	log.Printf("用户注册成功，DID: %s", did)
	return fmt.Sprintf("用户注册成功! DID: %s", did), nil
}

// UserLogin 用户登录
func (c *UserClient) UserLogin(did, name string) (string, error) {
	log.Printf("用户登录: %s, %s", did, name)
	
	// 参数验证
	if did == "" || name == "" {
		return "", fmt.Errorf("DID和姓名不能为空")
	}
	
	// 调用链码用户登录
	result, err := c.contract.SubmitTransaction(identityContract+":UserLogin", did, name)
	if err != nil {
		return "", fmt.Errorf("提交交易失败: %w", err)
	}
	
	loginResult := string(result)
	log.Printf("登录结果: %s", loginResult)
	
	return loginResult, nil
}

// UserLogout 用户登出
func (c *UserClient) UserLogout(did string) (string, error) {
	log.Printf("用户登出: %s", did)
	
	// 参数验证
	if did == "" {
		return "", fmt.Errorf("DID不能为空")
	}
	
	// 调用链码用户登出
	result, err := c.contract.SubmitTransaction(identityContract+":UserLogout", did)
	if err != nil {
		return "", fmt.Errorf("提交交易失败: %w", err)
	}
	
	logoutResult := string(result)
	log.Printf("登出结果: %s", logoutResult)
	
	return logoutResult, nil
}

// GetUser 获取用户信息
func (c *UserClient) GetUser(did string) (string, error) {
	log.Printf("获取用户信息: %s", did)
	
	// 参数验证
	if did == "" {
		return "", fmt.Errorf("DID不能为空")
	}
	
	// 调用链码获取用户信息
	result, err := c.contract.EvaluateTransaction(identityContract+":GetUser", did)
	if err != nil {
		return "", fmt.Errorf("评估交易失败: %w", err)
	}
	
	// 格式化JSON输出
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, result, "", "  "); err != nil {
		return string(result), nil // 如果格式化失败，返回原始结果
	}
	
	return prettyJSON.String(), nil
}

// GetDIDByInfo 根据用户信息获取DID
func (c *UserClient) GetDIDByInfo(name, idNumber, phoneNumber, vehicleID string) (string, error) {
	log.Printf("获取DID: %s, %s, %s, %s", name, idNumber, phoneNumber, vehicleID)
	
	// 参数验证
	if name == "" || idNumber == "" || phoneNumber == "" || vehicleID == "" {
		return "", fmt.Errorf("所有参数都不能为空")
	}
	
	// 调用链码获取DID
	result, err := c.contract.EvaluateTransaction(identityContract+":GetDIDByInfo", name, idNumber, phoneNumber, vehicleID)
	if err != nil {
		return "", fmt.Errorf("评估交易失败: %w", err)
	}
	
	return string(result), nil
}

// 辅助函数：加载证书
func loadCertificate(filename string) (*x509.Certificate, error) {
	certificatePEM, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("读取证书文件失败: %w", err)
	}
	return identity.CertificateFromPEM(certificatePEM)
}

// 辅助函数：加载私钥
func loadPrivateKey(dirPath string) (interface{}, error) {
	// 读取目录中的文件
	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("读取私钥目录失败: %w", err)
	}
	
	// 查找私钥文件
	for _, file := range files {
		if !file.IsDir() {
			privateKeyPEM, err := ioutil.ReadFile(path.Join(dirPath, file.Name()))
			if err != nil {
				return nil, fmt.Errorf("读取私钥文件失败: %w", err)
			}
			return identity.PrivateKeyFromPEM(privateKeyPEM)
		}
	}
	
	return nil, fmt.Errorf("在目录中未找到私钥文件")
}
