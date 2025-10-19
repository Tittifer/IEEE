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

	"github.com/Tittifer/IEEE/honeypoint_client/chain"
	"github.com/Tittifer/IEEE/honeypoint_client/risk"
)

// HoneypointClient 蜜点后台客户端结构体
type HoneypointClient struct {
	contract     *client.Contract
	gateway      *client.Gateway
	conn         *grpc.ClientConn
	config       *ConnectionConfig
	chainManager *chain.ChainManager
	riskAssessor *risk.RiskAssessor
	chainClient  *ChainClient
	network      *client.Network
	stopChan     chan struct{}
	isRunning    bool
	ctx          context.Context
	cancel       context.CancelFunc
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
	identityContract   = "IdentityContract"
	riskContract       = "RiskContract"
	configPath         = "config.json"
	riskScoreThreshold = 50.00 // 风险评分阈值
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

	// 创建上下文，用于取消事件监听
	ctx, cancel := context.WithCancel(context.Background())

	// 创建蜜点后台客户端
	honeypointClient := &HoneypointClient{
		contract:  contract,
		gateway:   gw,
		conn:      conn,
		config:    config,
		network:   network,
		stopChan:  make(chan struct{}),
		ctx:       ctx,
		cancel:    cancel,
	}

	// 创建区块链客户端
	chainClient := NewChainClient(honeypointClient)
	honeypointClient.chainClient = chainClient

	// 创建区块链管理器
	chainManager := chain.NewChainManager(chainClient)
	honeypointClient.chainManager = chainManager

	// 创建风险评估器
	riskAssessor := risk.NewRiskAssessor(chainManager)
	honeypointClient.riskAssessor = riskAssessor

	return honeypointClient, nil
}

// Close 关闭连接
func (c *HoneypointClient) Close() {
	// 停止事件监听
	if c.isRunning {
		c.StopEventListener()
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

	// 监听设备注册事件
	go c.listenForDeviceRegistered()

	// 监听风险评分更新事件
	go c.listenForRiskScoreUpdated()
	
	// 监听风险评分重置事件
	go c.listenForRiskScoreReset()

	// 启动周期性维护任务
	go c.startPeriodicMaintenance()

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

// listenForDeviceRegistered 监听设备注册事件
func (c *HoneypointClient) listenForDeviceRegistered() {
	log.Println("开始监听设备注册事件...")

	// 使用新的API监听链码事件
	events, err := c.network.ChaincodeEvents(c.ctx, c.config.ChaincodeName)
	if err != nil {
		log.Printf("注册设备注册事件监听失败: %v", err)
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

			// 只处理DeviceRegistered事件
			if event.EventName != "DeviceRegistered" {
				continue
			}

			// 解析事件数据
			var deviceEvent DeviceEvent
			if err := json.Unmarshal(event.Payload, &deviceEvent); err != nil {
				log.Printf("解析设备注册事件数据失败: %v", err)
				continue
			}

			log.Printf("收到设备注册事件: DID=%s, 名称=%s", deviceEvent.DID, deviceEvent.Name)
			
			// 设备注册后立即启动风险监控
			go c.startRiskMonitoring(deviceEvent.DID)
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
			var deviceEvent DeviceEvent
			if err := json.Unmarshal(event.Payload, &deviceEvent); err != nil {
				log.Printf("解析风险评分更新事件数据失败: %v", err)
				continue
			}

			log.Printf("收到风险评分更新事件: DID=%s, 名称=%s, 风险评分=%.2f", deviceEvent.DID, deviceEvent.Name, deviceEvent.RiskScore)
		}
	}
}

// startRiskMonitoring 启动风险监控
func (c *HoneypointClient) startRiskMonitoring(did string) {
	log.Printf("开始对设备 %s 进行风险监控", did)

	// 检查设备是否存在
	_, err := c.chainClient.GetDeviceInfo(did)
	if err != nil {
		log.Printf("获取设备信息失败: %v", err)
		return
	}

	// 列出可用的风险行为（仅用于验证规则加载成功）
	_ = c.riskAssessor.ListAvailableRiskBehaviors()

	// 日志已禁用，避免输出过多信息
	// log.Printf("可用的风险行为类型:")
	// for i, rule := range rules {
	// 	log.Printf("%d. %s - %s (得分: %.2f, 类别: %s, 权重: %.2f)", i+1, rule.BehaviorType, rule.Description, rule.Score, rule.Category, rule.Weight)
	// }

	log.Printf("设备 %s 已开始风险监控", did)
}

// ProcessRiskBehavior 处理设备风险行为
func (c *HoneypointClient) ProcessRiskBehavior(did string, behaviorType string) error {
	// 评估风险
	newScore, newAttackIndex, updatedProfile, err := c.riskAssessor.AssessRisk(did, behaviorType)
	if err != nil {
		return fmt.Errorf("风险评估失败: %w", err)
	}

	// 无论风险评分是否超过阈值，都立即向链上报告
	log.Printf("设备 %s 的风险评分为 %.2f，攻击画像指数为 %.2f，立即向链上报告", did, newScore, newAttackIndex)

	// 更新链上设备风险评分
	err = c.chainClient.UpdateDeviceRiskScore(did, newScore, newAttackIndex, updatedProfile)
	if err != nil {
		return fmt.Errorf("向链上报告风险评分失败: %w", err)
	}

	log.Printf("已向链上报告设备 %s 的风险评分 %.2f", did, newScore)

	// 检查风险评分是否超过阈值
	if newScore >= riskScoreThreshold {
		log.Printf("设备 %s 的风险评分 %.2f 超过阈值 %.2f，可能会被限制访问", did, newScore, riskScoreThreshold)
	}

	return nil
}

// listenForRiskScoreReset 监听风险评分重置事件
func (c *HoneypointClient) listenForRiskScoreReset() {
	log.Println("开始监听风险评分重置事件...")

	// 使用新的API监听链码事件
	events, err := c.network.ChaincodeEvents(c.ctx, c.config.ChaincodeName)
	if err != nil {
		log.Printf("注册风险评分重置事件监听失败: %v", err)
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

			// 只处理RiskScoreReset事件
			if event.EventName != "RiskScoreReset" {
				continue
			}

			// 解析事件数据
			var deviceEvent DeviceEvent
			if err := json.Unmarshal(event.Payload, &deviceEvent); err != nil {
				log.Printf("解析风险评分重置事件数据失败: %v", err)
				continue
			}

			log.Printf("收到风险评分重置事件: DID=%s, 名称=%s", deviceEvent.DID, deviceEvent.Name)

			// 重置设备风险数据
			if err := c.chainManager.ResetDeviceRiskData(deviceEvent.DID); err != nil {
				log.Printf("重置设备风险数据失败: %v", err)
				continue
			}

			log.Printf("设备 %s 的风险数据已重置", deviceEvent.DID)
		}
	}
}

// startPeriodicMaintenance 启动周期性维护任务
func (c *HoneypointClient) startPeriodicMaintenance() {
	// 每天执行一次维护任务
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	log.Println("启动周期性维护任务")

	for {
		select {
		case <-c.stopChan:
			return
		case <-ticker.C:
			// 获取所有设备
			devices, err := c.getAllDevices()
			if err != nil {
				log.Printf("获取所有设备失败: %v", err)
				continue
			}

			// 对每个设备执行维护任务
			for _, device := range devices {
				// 执行攻击画像指数的慢速衰减
				err := c.riskAssessor.PerformBackgroundMaintenance(device.DID)
				if err != nil {
					log.Printf("执行设备 %s 的维护任务失败: %v", device.DID, err)
				} else {
					log.Printf("已完成设备 %s 的维护任务", device.DID)
				}
			}
		}
	}
}

// getAllDevices 获取所有设备
func (c *HoneypointClient) getAllDevices() ([]*chain.Device, error) {
	// 调用链码获取所有设备
	devicesJSON, err := c.contract.EvaluateTransaction("IdentityContract:GetAllDevices")
	if err != nil {
		return nil, fmt.Errorf("评估交易失败: %w", err)
	}

	// 解析设备列表
	var devices []*chain.Device
	if err := json.Unmarshal(devicesJSON, &devices); err != nil {
		return nil, fmt.Errorf("设备列表解析失败: %w", err)
	}

	return devices, nil
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