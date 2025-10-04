package client

import (
	"context"
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

// HoneypointClient 后台客户端结构体
type HoneypointClient struct {
	contract *client.Contract
	gateway  *client.Gateway
	conn     *grpc.ClientConn
	config   *ConnectionConfig
	riskManager *RiskScoreManager // 风险评分管理器
	riskInput   *RiskInputManager // 风险行为输入管理器
}

// UserEvent 用户事件结构体，用于解析链码事件
type UserEvent struct {
	EventType string `json:"eventType"` // 事件类型：register, login, logout, risk_update
	DID       string `json:"did"`       // 用户DID
	Name      string `json:"name"`      // 用户名称
	Timestamp int64  `json:"timestamp"` // 事件时间戳
	RiskScore int    `json:"riskScore"` // 风险评分（可选）
}

// 常量定义
const (
	identityContract = "IdentityContract"
	configPath       = "../config.json"
)

// NewHoneypointClient 创建新的后台客户端
func NewHoneypointClient() (*HoneypointClient, error) {
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
	
	// 创建后台客户端
	honeypointClient := &HoneypointClient{
		contract: contract,
		gateway:  gw,
		conn:     conn,
		config:   config,
	}
	
	// 创建风险评分管理器并关联到客户端
	honeypointClient.riskManager = NewRiskScoreManager(honeypointClient)
	
	// 创建风险行为输入管理器并关联到客户端
	honeypointClient.riskInput = NewRiskInputManager(honeypointClient)
	
	return honeypointClient, nil
}

// Close 关闭连接
func (c *HoneypointClient) Close() {
	// 如果风险行为输入管理器正在运行，先停止它
	if c.riskInput != nil {
		c.riskInput.Stop()
	}
	
	c.gateway.Close()
	c.conn.Close()
}

// StartEventListener 启动事件监听
func (c *HoneypointClient) StartEventListener() (chan UserEvent, error) {
	log.Println("启动事件监听...")
	
	// 创建事件通道
	eventCh := make(chan UserEvent, 100)
	
	// 获取网络
	network := c.gateway.GetNetwork(c.config.ChannelName)
	
	// 定义要监听的事件
	eventNames := []string{
		"UserRegistered", // 用户注册事件
		"UserLoggedIn",   // 用户登录事件
		"UserLoggedOut",  // 用户登出事件
		"RiskScoreUpdated", // 风险评分更新事件
	}
	
	// 监听所有事件
	for _, eventName := range eventNames {
		go func(name string) {
			log.Printf("开始监听 %s 事件...", name)
			
			// 创建上下文
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			
			events, err := network.ChaincodeEvents(ctx, c.config.ChaincodeName)
			if err != nil {
				log.Printf("注册 %s 事件监听失败: %v", name, err)
				return
			}
			
			// 处理事件
			for event := range events {
				// 如果不是我们要监听的事件，则跳过
				if event.EventName != name {
					continue
				}
				
				log.Printf("收到链码事件: %s", event.EventName)
				
				// 解析事件数据
				var userEvent UserEvent
				if err := json.Unmarshal(event.Payload, &userEvent); err != nil {
					log.Printf("解析事件数据失败: %v", err)
					continue
				}
				
				// 根据事件名称设置事件类型
				switch event.EventName {
				case "UserRegistered":
					userEvent.EventType = "register"
				case "UserLoggedIn":
					userEvent.EventType = "login"
				case "UserLoggedOut":
					userEvent.EventType = "logout"
				case "RiskScoreUpdated":
					userEvent.EventType = "risk_update"
				}
				
				// 发送事件到通道
				eventCh <- userEvent
			}
		}(eventName)
	}
	
	return eventCh, nil
}

// StartRiskInputListener 启动风险行为输入监听
func (c *HoneypointClient) StartRiskInputListener() {
	if c.riskInput != nil {
		c.riskInput.Start()
	} else {
		log.Println("风险行为输入管理器未初始化")
	}
}

// ProcessEvent 处理用户事件
func (c *HoneypointClient) ProcessEvent(event UserEvent) {
	log.Printf("处理事件: %s, DID: %s, 用户名: %s", event.EventType, event.DID, event.Name)
	
	switch event.EventType {
	case "register":
		c.handleUserRegistration(event)
	case "login":
		c.handleUserLogin(event)
	case "logout":
		c.handleUserLogout(event)
	case "risk_update":
		c.handleRiskUpdate(event)
	default:
		log.Printf("未知事件类型: %s", event.EventType)
	}
}

// 处理用户注册事件
func (c *HoneypointClient) handleUserRegistration(event UserEvent) {
	log.Printf("处理用户注册: DID=%s, 用户名=%s, 时间戳=%d", event.DID, event.Name, event.Timestamp)
	
	// 在风险评分管理器中注册用户，初始化风险评分为0
	c.riskManager.RegisterUser(event.DID)
	
	log.Printf("用户 %s 注册成功处理完成", event.Name)
}

// 处理用户登录事件
func (c *HoneypointClient) handleUserLogin(event UserEvent) {
	log.Printf("处理用户登录: DID=%s, 用户名=%s, 时间戳=%d", event.DID, event.Name, event.Timestamp)
	
	// 确保用户在风险评分管理器中注册
	c.riskManager.UserLogin(event.DID)
	
	log.Printf("用户 %s 登录成功处理完成", event.Name)
}

// 处理用户登出事件
func (c *HoneypointClient) handleUserLogout(event UserEvent) {
	log.Printf("处理用户登出: DID=%s, 用户名=%s, 时间戳=%d", event.DID, event.Name, event.Timestamp)
	
	// 记录用户登出事件
	c.riskManager.UserLogout(event.DID)
	
	log.Printf("用户 %s 登出成功处理完成", event.Name)
}

// 处理风险评分更新事件
func (c *HoneypointClient) handleRiskUpdate(event UserEvent) {
	log.Printf("处理风险评分更新: DID=%s, 用户名=%s, 风险评分=%d, 时间戳=%d", 
		event.DID, event.Name, event.RiskScore, event.Timestamp)
	
	// 检查是否超过风险阈值，如果超过则记录警告
	if event.RiskScore >= RiskScoreThreshold {
		log.Printf("警告：用户 %s 风险评分 %d 已超过阈值 %d", 
			event.Name, event.RiskScore, RiskScoreThreshold)
	}
	
	log.Printf("用户 %s 风险评分更新处理完成", event.Name)
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