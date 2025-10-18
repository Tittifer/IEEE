package contracts

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/Tittifer/IEEE/chain/models"
	"github.com/Tittifer/IEEE/chain/utils"
)

// IdentityContract 身份管理合约
type IdentityContract struct {
	contractapi.Contract
}

// InitLedger 初始化账本
func (c *IdentityContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	fmt.Println("身份认证链码初始化")
	return nil
}

// RegisterDevice 注册新设备
func (c *IdentityContract) RegisterDevice(ctx contractapi.TransactionContextInterface, name, model, vendor, deviceID string) error {
	// 根据设备信息生成DID
	did := utils.GenerateDID(name, model, vendor, deviceID)
	
	// 检查设备是否已存在
	exists, err := c.DeviceExists(ctx, did)
	if err != nil {
		return fmt.Errorf("检查设备是否存在时出错: %v", err)
	}
	if exists {
		return fmt.Errorf("设备DID %s 已存在", did)
	}

	// 使用交易时间戳确保确定性
	timestamp, err := ctx.GetStub().GetTxTimestamp()
	if err != nil {
		return fmt.Errorf("获取交易时间戳失败: %v", err)
	}
	txTime := time.Unix(timestamp.Seconds, int64(timestamp.Nanos))
	
	// 创建新设备，初始化风险相关参数
	deviceInfo := models.DeviceInfo{
		DID:           did,
		Name:          name,
		Model:         model,
		Vendor:        vendor,
		RiskScore:     0.0,               // 初始风险评分为0
		AttackIndexI:  0.0,               // 初始攻击画像指数为0
		AttackProfile: []string{},        // 初始攻击画像为空
		LastEventTime: txTime,            // 初始事件时间为当前时间
		Status:        models.StatusActive,
		CreatedAt:     txTime,
		LastUpdatedAt: txTime,
	}

	// 将设备信息转换为JSON并存储
	deviceInfoJSON, err := json.Marshal(deviceInfo)
	if err != nil {
		return fmt.Errorf("设备信息序列化失败: %v", err)
	}

	// 将设备信息写入账本
	err = ctx.GetStub().PutState(did, deviceInfoJSON)
	if err != nil {
		return fmt.Errorf("存储设备信息时出错: %v", err)
	}

	// 创建设备注册事件
	registerEvent := models.DeviceEvent{
		EventType: models.EventTypeRegister,
		DID:       did,
		Name:      name,
		Timestamp: txTime.Unix(),
	}

	// 序列化事件数据
	eventJSON, err := json.Marshal(registerEvent)
	if err != nil {
		return fmt.Errorf("事件数据序列化失败: %v", err)
	}

	// 发送设备注册事件
	err = ctx.GetStub().SetEvent("DeviceRegistered", eventJSON)
	if err != nil {
		return fmt.Errorf("发送设备注册事件失败: %v", err)
	}

	return nil
}

// GetDevice 获取设备信息
func (c *IdentityContract) GetDevice(ctx contractapi.TransactionContextInterface, did string) (string, error) {
	// 验证DID格式
	if !utils.ValidateDID(did) {
		return "", fmt.Errorf("无效的DID格式: %s", did)
	}
	
	// 从账本中读取设备信息
	deviceInfoJSON, err := ctx.GetStub().GetState(did)
	if err != nil {
		return "", fmt.Errorf("读取设备信息时出错: %v", err)
	}
	if deviceInfoJSON == nil {
		return "", fmt.Errorf("设备DID %s 不存在", did)
	}

	// 直接返回JSON字符串，无需再次序列化
	return string(deviceInfoJSON), nil
}

// DeviceExists 检查设备是否存在
func (c *IdentityContract) DeviceExists(ctx contractapi.TransactionContextInterface, did string) (bool, error) {
	// 从账本中读取设备信息
	deviceInfoJSON, err := ctx.GetStub().GetState(did)
	if err != nil {
		return false, fmt.Errorf("读取设备信息时出错: %v", err)
	}
	
	// 如果设备信息不为空，则设备存在
	return deviceInfoJSON != nil, nil
}

// GetDIDByInfo 根据设备信息生成DID
func (c *IdentityContract) GetDIDByInfo(ctx contractapi.TransactionContextInterface, name, model, vendor, deviceID string) (string, error) {
	// 根据设备信息生成DID
	did := utils.GenerateDID(name, model, vendor, deviceID)
	return did, nil
}

// VerifyDeviceIdentity 验证设备身份
func (c *IdentityContract) VerifyDeviceIdentity(ctx contractapi.TransactionContextInterface, did, name, model string) (bool, error) {
	// 验证DID格式
	if !utils.ValidateDID(did) {
		return false, fmt.Errorf("无效的DID格式: %s", did)
	}
	
	// 从账本中读取设备信息
	deviceInfoJSON, err := ctx.GetStub().GetState(did)
	if err != nil {
		return false, fmt.Errorf("读取设备信息时出错: %v", err)
	}
	if deviceInfoJSON == nil {
		return false, fmt.Errorf("设备DID %s 不存在", did)
	}

	// 反序列化设备信息
	var deviceInfo models.DeviceInfo
	err = json.Unmarshal(deviceInfoJSON, &deviceInfo)
	if err != nil {
		return false, fmt.Errorf("设备信息反序列化失败: %v", err)
	}
	
	// 验证设备名称和型号是否匹配
	if deviceInfo.Name != name || deviceInfo.Model != model {
		return false, nil
	}
	
	// 验证设备状态是否为活跃
	if deviceInfo.Status != models.StatusActive {
		return false, fmt.Errorf("设备状态为 %s, 非活跃状态", deviceInfo.Status)
	}
	
	return true, nil
}

// UpdateDeviceRiskScore 更新设备风险评分
func (c *IdentityContract) UpdateDeviceRiskScore(ctx contractapi.TransactionContextInterface, did string, riskScoreStr string, attackIndexStr string, attackProfileJSON string) error {
	// 验证DID格式
	if !utils.ValidateDID(did) {
		return fmt.Errorf("无效的DID格式: %s", did)
	}
	
	// 解析风险评分和攻击画像指数
	riskScore, err := strconv.ParseFloat(riskScoreStr, 64)
	if err != nil {
		return fmt.Errorf("无效的风险评分格式: %s", riskScoreStr)
	}
	
	attackIndex, err := strconv.ParseFloat(attackIndexStr, 64)
	if err != nil {
		return fmt.Errorf("无效的攻击画像指数格式: %s", attackIndexStr)
	}
	
	// 解析攻击画像JSON
	var attackProfile []string
	err = json.Unmarshal([]byte(attackProfileJSON), &attackProfile)
	if err != nil {
		return fmt.Errorf("攻击画像JSON解析失败: %v", err)
	}
	
	// 获取设备信息
	deviceInfoJSON, err := ctx.GetStub().GetState(did)
	if err != nil {
		return fmt.Errorf("读取设备信息时出错: %v", err)
	}
	if deviceInfoJSON == nil {
		return fmt.Errorf("设备DID %s 不存在", did)
	}
	
	// 反序列化设备信息
	var deviceInfo models.DeviceInfo
	err = json.Unmarshal(deviceInfoJSON, &deviceInfo)
	if err != nil {
		return fmt.Errorf("设备信息反序列化失败: %v", err)
	}
	
	// 更新风险评分和攻击画像
	deviceInfo.RiskScore = riskScore
	deviceInfo.AttackIndexI = attackIndex
	deviceInfo.AttackProfile = attackProfile
	
	// 如果风险评分超过阈值，更新设备状态为风险状态
	if riskScore > models.RiskScoreThreshold && deviceInfo.Status != models.StatusRisky {
		deviceInfo.Status = models.StatusRisky
	}
	
	// 使用交易时间戳更新最后更新时间和最后事件时间
	timestamp, err := ctx.GetStub().GetTxTimestamp()
	if err != nil {
		return fmt.Errorf("获取交易时间戳失败: %v", err)
	}
	txTime := time.Unix(timestamp.Seconds, int64(timestamp.Nanos))
	deviceInfo.LastUpdatedAt = txTime
	deviceInfo.LastEventTime = txTime
	
	// 将更新后的设备信息保存到账本
	deviceInfoJSON, err = json.Marshal(deviceInfo)
	if err != nil {
		return fmt.Errorf("设备信息序列化失败: %v", err)
	}
	
	err = ctx.GetStub().PutState(did, deviceInfoJSON)
	if err != nil {
		return fmt.Errorf("更新设备信息时出错: %v", err)
	}
	
	// 创建风险评分更新事件
	riskScoreEvent := models.DeviceEvent{
		EventType: models.EventTypeRiskUpdate,
		DID:       did,
		Name:      deviceInfo.Name,
		Timestamp: txTime.Unix(),
		RiskScore: riskScore,
	}

	// 序列化事件数据
	eventJSON, err := json.Marshal(riskScoreEvent)
	if err != nil {
		return fmt.Errorf("事件数据序列化失败: %v", err)
	}

	// 发送风险评分更新事件
	err = ctx.GetStub().SetEvent("RiskScoreUpdated", eventJSON)
	if err != nil {
		return fmt.Errorf("发送风险评分更新事件失败: %v", err)
	}
	
	return nil
}

// ResetDeviceRiskScore 重置设备风险评分
func (c *IdentityContract) ResetDeviceRiskScore(ctx contractapi.TransactionContextInterface, did string) error {
	// 验证DID格式
	if !utils.ValidateDID(did) {
		return fmt.Errorf("无效的DID格式: %s", did)
	}
	
	// 获取设备信息
	deviceInfoJSON, err := ctx.GetStub().GetState(did)
	if err != nil {
		return fmt.Errorf("读取设备信息时出错: %v", err)
	}
	if deviceInfoJSON == nil {
		return fmt.Errorf("设备DID %s 不存在", did)
	}
	
	// 反序列化设备信息
	var deviceInfo models.DeviceInfo
	err = json.Unmarshal(deviceInfoJSON, &deviceInfo)
	if err != nil {
		return fmt.Errorf("设备信息反序列化失败: %v", err)
	}
	
	// 重置风险评分和攻击画像
	deviceInfo.RiskScore = 0.0
	deviceInfo.AttackIndexI = 0.0
	deviceInfo.AttackProfile = []string{}
	
	// 如果设备状态为风险状态，恢复为活跃状态
	if deviceInfo.Status == models.StatusRisky {
		deviceInfo.Status = models.StatusActive
	}
	
	// 使用交易时间戳更新最后更新时间
	timestamp, err := ctx.GetStub().GetTxTimestamp()
	if err != nil {
		return fmt.Errorf("获取交易时间戳失败: %v", err)
	}
	txTime := time.Unix(timestamp.Seconds, int64(timestamp.Nanos))
	deviceInfo.LastUpdatedAt = txTime
	
	// 将更新后的设备信息保存到账本
	deviceInfoJSON, err = json.Marshal(deviceInfo)
	if err != nil {
		return fmt.Errorf("设备信息序列化失败: %v", err)
	}
	
	err = ctx.GetStub().PutState(did, deviceInfoJSON)
	if err != nil {
		return fmt.Errorf("更新设备信息时出错: %v", err)
	}
	
	// 创建风险评分重置事件
	resetEvent := models.DeviceEvent{
		EventType: models.EventTypeRiskReset,
		DID:       did,
		Name:      deviceInfo.Name,
		Timestamp: txTime.Unix(),
		RiskScore: 0.0,
	}

	// 序列化事件数据
	eventJSON, err := json.Marshal(resetEvent)
	if err != nil {
		return fmt.Errorf("事件数据序列化失败: %v", err)
	}

	// 发送风险评分重置事件
	err = ctx.GetStub().SetEvent("RiskScoreReset", eventJSON)
	if err != nil {
		return fmt.Errorf("发送风险评分重置事件失败: %v", err)
	}
	
	return nil
}

// GetAllDevices 获取所有设备
func (c *IdentityContract) GetAllDevices(ctx contractapi.TransactionContextInterface) (string, error) {
	// 获取所有设备的迭代器
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return "", fmt.Errorf("获取状态范围时出错: %v", err)
	}
	defer resultsIterator.Close()

	// 存储所有设备的切片
	var devices []*models.DeviceInfo
	
	// 遍历所有状态
	for resultsIterator.HasNext() {
		// 获取下一个键值对
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return "", fmt.Errorf("获取下一个状态时出错: %v", err)
		}

		// 反序列化设备信息
		var deviceInfo models.DeviceInfo
		err = json.Unmarshal(queryResponse.Value, &deviceInfo)
		if err != nil {
			// 跳过非设备信息的状态
			continue
		}

		// 将设备信息添加到切片中
		devices = append(devices, &deviceInfo)
	}

	// 将设备列表转换为JSON字符串
	devicesJSON, err := json.Marshal(devices)
	if err != nil {
		return "", fmt.Errorf("序列化设备列表失败: %v", err)
	}

	return string(devicesJSON), nil
}