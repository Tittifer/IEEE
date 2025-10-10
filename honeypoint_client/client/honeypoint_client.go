package client

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"path"
	"strconv"
	"time"

	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/Tittifer/IEEE/honeypoint_client/db"
	"github.com/Tittifer/IEEE/honeypoint_client/risk"
)

// HoneypointClient 蜜点后台客户端结构体
type HoneypointClient struct {
	contract     *client.Contract
	gateway      *client.Gateway
	conn         *grpc.ClientConn
	config       *ConnectionConfig
	dbManager    *db.DBManager
	riskAssessor *risk.RiskAssessor
	network      *client.Network
	stopChan     chan struct{}
	isRunning    bool
	ctx          context.Context
	cancel       context.CancelFunc
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
	identityContract   = "IdentityContract"
	configPath         = "config.json" // 修改为相对于执行目录的路径
	riskScoreThreshold = 50            // 风险评分阈值
)

// NewHoneypointClient 创建新的蜜点后台客户端
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

	// 创建数据库管理器
	dbManager, err := db.NewDBManager(config.DBHost, config.DBPort, config.DBUser, config.DBPassword, config.DBName)
	if err != nil {
		gw.Close()
		conn.Close()
		return nil, fmt.Errorf("创建数据库管理器失败: %w", err)
	}

	// 创建风险评估器
	riskAssessor := risk.NewRiskAssessor(dbManager)

	// 创建上下文，用于取消事件监听
	ctx, cancel := context.WithCancel(context.Background())

	// 创建蜜点后台客户端
	honeypointClient := &HoneypointClient{
		contract:     contract,
		gateway:      gw,
		conn:         conn,
		config:       config,
		dbManager:    dbManager,
		riskAssessor: riskAssessor,
		network:      network,
		stopChan:     make(chan struct{}),
		ctx:          ctx,
		cancel:       cancel,
	}

	return honeypointClient, nil
}

// Close 关闭连接
func (c *HoneypointClient) Close() {
	// 停止事件监听
	if c.isRunning {
		c.StopEventListener()
	}

	// 关闭数据库连接
	if c.dbManager != nil {
		c.dbManager.Close()
	}

	// 关闭区块链连接
	if c.gateway != nil {
		c.gateway.Close()
	}

	if c.conn != nil {
		c.conn.Close()
	}

	// 取消上下文
	if c.cancel != nil {
		c.cancel()
	}
}

// StartEventListener 启动事件监听
func (c *HoneypointClient) StartEventListener() error {
	if c.isRunning {
		return fmt.Errorf("事件监听器已经在运行")
	}

	c.isRunning = true

	// 监听用户注册事件
	go c.listenForUserRegistered()

	// 监听用户登录事件
	go c.listenForUserLoggedIn()

	// 监听用户登出事件
	go c.listenForUserLoggedOut()

	// 监听风险评分更新事件
	go c.listenForRiskScoreUpdated()

	log.Println("事件监听器已启动")
	return nil
}

// StopEventListener 停止事件监听
func (c *HoneypointClient) StopEventListener() {
	if !c.isRunning {
		return
	}

	close(c.stopChan)
	c.cancel() // 取消上下文，停止所有事件监听
	c.isRunning = false
	log.Println("事件监听器已停止")
}

// listenForUserRegistered 监听用户注册事件
func (c *HoneypointClient) listenForUserRegistered() {
	log.Println("开始监听用户注册事件...")

	// 使用新的API监听链码事件
	events, err := c.network.ChaincodeEvents(c.ctx, c.config.ChaincodeName)
	if err != nil {
		log.Printf("注册用户注册事件监听失败: %v", err)
		return
	}

	for {
		select {
		case <-c.stopChan:
			return
		case event, ok := <-events:
			if !ok {
				return
			}

			// 只处理UserRegistered事件
			if event.EventName != "UserRegistered" {
				continue
			}

			// 解析事件数据
			var userEvent UserEvent
			if err := json.Unmarshal(event.Payload, &userEvent); err != nil {
				log.Printf("解析用户注册事件数据失败: %v", err)
				continue
			}

			log.Printf("收到用户注册事件: DID=%s, 姓名=%s", userEvent.DID, userEvent.Name)

			// 在数据库中创建用户
			_, err := c.dbManager.CreateUser(userEvent.DID, userEvent.Name)
			if err != nil {
				log.Printf("创建用户失败: %v", err)
				continue
			}

			log.Printf("用户 %s 已添加到数据库", userEvent.DID)
		}
	}
}

// listenForUserLoggedIn 监听用户登录事件
func (c *HoneypointClient) listenForUserLoggedIn() {
	log.Println("开始监听用户登录事件...")

	// 使用新的API监听链码事件
	events, err := c.network.ChaincodeEvents(c.ctx, c.config.ChaincodeName)
	if err != nil {
		log.Printf("注册用户登录事件监听失败: %v", err)
		return
	}

	for {
		select {
		case <-c.stopChan:
			return
		case event, ok := <-events:
			if !ok {
				return
			}

			// 只处理UserLoggedIn事件
			if event.EventName != "UserLoggedIn" {
				continue
			}

			// 解析事件数据
			var userEvent UserEvent
			if err := json.Unmarshal(event.Payload, &userEvent); err != nil {
				log.Printf("解析用户登录事件数据失败: %v", err)
				continue
			}

			log.Printf("收到用户登录事件: DID=%s, 姓名=%s", userEvent.DID, userEvent.Name)

			// 检查用户是否存在
			user, err := c.dbManager.GetUserByDID(userEvent.DID)
			if err != nil {
				log.Printf("获取用户信息失败: %v", err)
				continue
			}

			if user == nil {
				// 用户不存在，创建用户
				_, err := c.dbManager.CreateUser(userEvent.DID, userEvent.Name)
				if err != nil {
					log.Printf("创建用户失败: %v", err)
					continue
				}
				log.Printf("用户 %s 已添加到数据库", userEvent.DID)
			} else {
				log.Printf("用户 %s 已登录，当前风险评分: %d", userEvent.DID, user.CurrentScore)
			}

			// 启动风险监控
			go c.startRiskMonitoring(userEvent.DID)
		}
	}
}

// listenForUserLoggedOut 监听用户登出事件
func (c *HoneypointClient) listenForUserLoggedOut() {
	log.Println("开始监听用户登出事件...")

	// 使用新的API监听链码事件
	events, err := c.network.ChaincodeEvents(c.ctx, c.config.ChaincodeName)
	if err != nil {
		log.Printf("注册用户登出事件监听失败: %v", err)
		return
	}

	for {
		select {
		case <-c.stopChan:
			return
		case event, ok := <-events:
			if !ok {
				return
			}

			// 只处理UserLoggedOut事件
			if event.EventName != "UserLoggedOut" {
				continue
			}

			// 解析事件数据
			var userEvent UserEvent
			if err := json.Unmarshal(event.Payload, &userEvent); err != nil {
				log.Printf("解析用户登出事件数据失败: %v", err)
				continue
			}

			log.Printf("收到用户登出事件: DID=%s, 姓名=%s", userEvent.DID, userEvent.Name)

			// 用户登出，不需要特别处理，只记录日志
			log.Printf("用户 %s 已登出", userEvent.DID)
		}
	}
}

// listenForRiskScoreUpdated 监听风险评分更新事件
func (c *HoneypointClient) listenForRiskScoreUpdated() {
	log.Println("开始监听风险评分更新事件...")

	// 使用新的API监听链码事件
	events, err := c.network.ChaincodeEvents(c.ctx, c.config.ChaincodeName)
	if err != nil {
		log.Printf("注册风险评分更新事件监听失败: %v", err)
		return
	}

	for {
		select {
		case <-c.stopChan:
			return
		case event, ok := <-events:
			if !ok {
				return
			}

			// 只处理RiskScoreUpdated事件
			if event.EventName != "RiskScoreUpdated" {
				continue
			}

			// 解析事件数据
			var userEvent UserEvent
			if err := json.Unmarshal(event.Payload, &userEvent); err != nil {
				log.Printf("解析风险评分更新事件数据失败: %v", err)
				continue
			}

			log.Printf("收到风险评分更新事件: DID=%s, 姓名=%s, 风险评分=%d", userEvent.DID, userEvent.Name, userEvent.RiskScore)

			// 获取用户信息
			user, err := c.dbManager.GetUserByDID(userEvent.DID)
			if err != nil {
				log.Printf("获取用户信息失败: %v", err)
				continue
			}

			if user == nil {
				log.Printf("用户 %s 不存在", userEvent.DID)
				continue
			}

			// 更新用户风险评分
			if err := c.dbManager.UpdateUserRiskScore(user.ID, userEvent.RiskScore); err != nil {
				log.Printf("更新用户风险评分失败: %v", err)
				continue
			}

			log.Printf("用户 %s 的风险评分已更新为 %d", userEvent.DID, userEvent.RiskScore)
		}
	}
}

// startRiskMonitoring 启动风险监控
func (c *HoneypointClient) startRiskMonitoring(did string) {
	log.Printf("开始对用户 %s 进行风险监控", did)

	// 获取用户信息
	user, err := c.dbManager.GetUserByDID(did)
	if err != nil {
		log.Printf("获取用户信息失败: %v", err)
		return
	}

	if user == nil {
		log.Printf("用户 %s 不存在", did)
		return
	}

	// 模拟接收风险行为
	go func() {
		// 列出可用的风险行为
		rules, err := c.riskAssessor.ListAvailableRiskBehaviors()
		if err != nil {
			log.Printf("获取风险行为列表失败: %v", err)
			return
		}

		log.Printf("可用的风险行为类型:")
		for i, rule := range rules {
			log.Printf("%d. %s - %s (得分: %d)", i+1, rule.BehaviorType, rule.Description, rule.Score)
		}

		log.Printf("请通过命令行输入风险行为类型来模拟用户风险行为")
	}()
}

// ProcessRiskBehavior 处理用户风险行为
func (c *HoneypointClient) ProcessRiskBehavior(did string, behaviorType string) error {
	// 评估风险
	newScore, err := c.riskAssessor.AssessRisk(did, behaviorType)
	if err != nil {
		return fmt.Errorf("风险评估失败: %w", err)
	}

	// 如果风险评分超过阈值，向链上报告
	if newScore >= riskScoreThreshold {
		log.Printf("用户 %s 的风险评分 %d 超过阈值 %d，向链上报告", did, newScore, riskScoreThreshold)

		// 向链上报告风险评分
		_, err := c.contract.SubmitTransaction(identityContract+":UpdateRiskScore", did, strconv.Itoa(newScore))
		if err != nil {
			return fmt.Errorf("向链上报告风险评分失败: %w", err)
		}

		log.Printf("已向链上报告用户 %s 的风险评分", did)
	} else {
		log.Printf("用户 %s 的风险评分 %d 未超过阈值 %d", did, newScore, riskScoreThreshold)
	}

	return nil
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