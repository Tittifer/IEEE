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

// DeviceClient 设备客户端结构体
type DeviceClient struct {
	contract    *client.Contract
	gateway     *client.Gateway
	conn        *grpc.ClientConn
	config      *ConnectionConfig
	deviceName  string // 设备名称
	deviceModel string // 设备型号
	deviceVendor string // 设备供应商
}

// 设备事件结构体，用于解析链码事件
type DeviceEvent struct {
	EventType    string    `json:"eventType"`    // 事件类型
	DID          string    `json:"did"`          // 设备DID
	Name         string    `json:"name"`         // 设备名称
	Timestamp    int64     `json:"timestamp"`    // 事件时间戳
	RiskScore    float64   `json:"riskScore"`    // 风险评分
	Category     string    `json:"category"`     // 行为类别
	BehaviorType string    `json:"behaviorType"` // 具体行为类型
}

// 合约名称常量
const (
	identityContract = "IdentityContract"
	riskContract     = "RiskContract"
	configPath       = "../config.json"
)

// NewDeviceClient 创建新的设备客户端
func NewDeviceClient() (*DeviceClient, error) {
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
	
	// 创建设备客户端
	deviceClient := &DeviceClient{
		contract: contract,
		gateway:  gw,
		conn:     conn,
		config:   config,
	}
	
	return deviceClient, nil
}

// Close 关闭连接
func (c *DeviceClient) Close() {
	c.gateway.Close()
	c.conn.Close()
}

// RegisterDevice 注册新设备
func (c *DeviceClient) RegisterDevice(name, model, vendor, deviceID string) (string, error) {
	log.Printf("注册设备: %s, %s, %s, %s", name, model, vendor, deviceID)
	
	// 参数验证
	if name == "" || model == "" || vendor == "" || deviceID == "" {
		return "", fmt.Errorf("所有参数都不能为空")
	}
	
	// 先生成DID，检查设备是否已存在
	did, err := c.GetDIDByInfo(name, model, vendor, deviceID)
	if err != nil {
		return "", fmt.Errorf("生成DID失败: %w", err)
	}
	
	// 检查设备是否已存在
	deviceJSON, err := c.contract.EvaluateTransaction(identityContract+":GetDevice", did)
	if err == nil && len(deviceJSON) > 0 {
		// 设备已存在
		return "", fmt.Errorf("设备已存在，DID: %s", did)
	}
	
	// 调用链码注册设备
	_, err = c.contract.SubmitTransaction(identityContract+":RegisterDevice", name, model, vendor, deviceID)
	if err != nil {
		return "", fmt.Errorf("提交交易失败: %w", err)
	}
	
	// 保存设备信息
	c.deviceName = name
	c.deviceModel = model
	c.deviceVendor = vendor
	
	log.Printf("设备注册成功，DID: %s", did)
	return fmt.Sprintf("设备注册成功! DID: %s", did), nil
}

// GetDevice 获取设备信息
func (c *DeviceClient) GetDevice(did string) (string, error) {
	log.Printf("获取设备信息: %s", did)
	
	// 参数验证
	if did == "" {
		return "", fmt.Errorf("DID不能为空")
	}
	
	// 调用链码获取设备信息
	result, err := c.contract.EvaluateTransaction(identityContract+":GetDevice", did)
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

// GetDIDByInfo 根据设备信息获取DID
func (c *DeviceClient) GetDIDByInfo(name, model, vendor, deviceID string) (string, error) {
	log.Printf("获取DID: %s, %s, %s, %s", name, model, vendor, deviceID)
	
	// 参数验证
	if name == "" || model == "" || vendor == "" || deviceID == "" {
		return "", fmt.Errorf("所有参数都不能为空")
	}
	
	// 调用链码获取DID
	result, err := c.contract.EvaluateTransaction(identityContract+":GetDIDByInfo", name, model, vendor, deviceID)
	if err != nil {
		return "", fmt.Errorf("评估交易失败: %w", err)
	}
	
	return string(result), nil
}

// ResetDeviceRiskScore 重置设备风险评分
func (c *DeviceClient) ResetDeviceRiskScore(did string) (string, error) {
	log.Printf("重置设备风险评分: %s", did)
	
	// 参数验证
	if did == "" {
		return "", fmt.Errorf("DID不能为空")
	}
	
	// 检查设备是否存在
	deviceJSON, err := c.contract.EvaluateTransaction(identityContract+":GetDevice", did)
	if err != nil {
		return "", fmt.Errorf("获取设备信息失败: %w", err)
	}
	if len(deviceJSON) == 0 {
		return "", fmt.Errorf("设备DID %s 不存在", did)
	}
	
	// 调用链码重置设备风险评分
	_, err = c.contract.SubmitTransaction(identityContract+":ResetDeviceRiskScore", did)
	if err != nil {
		return "", fmt.Errorf("提交交易失败: %w", err)
	}
	
	log.Printf("设备 %s 的风险评分已重置为0", did)
	return "设备风险评分已重置为0", nil
}

// GetRiskResponse 获取风险响应策略
func (c *DeviceClient) GetRiskResponse(did string) (string, error) {
	log.Printf("获取设备风险响应策略: %s", did)
	
	// 参数验证
	if did == "" {
		return "", fmt.Errorf("DID不能为空")
	}
	
	// 调用链码获取风险响应策略
	result, err := c.contract.EvaluateTransaction(riskContract+":GetDeviceRiskResponse", did)
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