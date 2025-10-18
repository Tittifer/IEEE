package contracts

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/Tittifer/IEEE/chain/models"
	"github.com/Tittifer/IEEE/chain/utils"
)

// RiskContract 风险管理合约
type RiskContract struct {
	contractapi.Contract
}

// InitRiskLedger 初始化风险管理账本
func (c *RiskContract) InitRiskLedger(ctx contractapi.TransactionContextInterface) error {
	fmt.Println("风险管理链码初始化")
	return nil
}

// UpdateRiskScore 更新设备风险评分
func (c *RiskContract) UpdateRiskScore(ctx contractapi.TransactionContextInterface, did string, newScoreStr string, attackIndexStr string, attackProfileJSON string) error {
	// 验证DID格式
	if !utils.ValidateDID(did) {
		return fmt.Errorf("无效的DID格式: %s", did)
	}
	
	// 将字符串转换为浮点数
	newScore, err := strconv.ParseFloat(newScoreStr, 64)
	if err != nil {
		return fmt.Errorf("风险评分格式无效: %v", err)
	}
	
	// 验证风险评分范围
	if newScore < 0 {
		return fmt.Errorf("风险评分必须大于等于0")
	}
	
	attackIndex, err := strconv.ParseFloat(attackIndexStr, 64)
	if err != nil {
		return fmt.Errorf("攻击画像指数格式无效: %v", err)
	}
	
	// 验证攻击画像指数范围
	if attackIndex < 0 {
		return fmt.Errorf("攻击画像指数必须大于等于0")
	}
	
	// 解析攻击画像JSON
	var attackProfile []string
	err = json.Unmarshal([]byte(attackProfileJSON), &attackProfile)
	if err != nil {
		return fmt.Errorf("攻击画像JSON解析失败: %v", err)
	}
	
	// 从账本中读取设备信息
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
	deviceInfo.RiskScore = newScore
	deviceInfo.AttackIndexI = attackIndex
	deviceInfo.AttackProfile = attackProfile
	
	// 根据风险评分更新设备状态
	if newScore >= 700 {
		// 高危：硬性阻断
		deviceInfo.Status = models.StatusRisky
	} else if deviceInfo.Status == models.StatusRisky && newScore < models.RiskScoreThreshold {
		// 如果设备之前是风险状态，但现在风险评分降低，恢复为活跃状态
		deviceInfo.Status = models.StatusActive
	}
	
	// 将设备信息转换为JSON并存储
	updatedDeviceInfoJSON, err := json.Marshal(deviceInfo)
	if err != nil {
		return fmt.Errorf("设备信息序列化失败: %v", err)
	}
	
	// 将更新后的设备信息写入账本
	err = ctx.GetStub().PutState(did, updatedDeviceInfoJSON)
	if err != nil {
		return fmt.Errorf("存储设备信息时出错: %v", err)
	}
	
	return nil
}

// GetRiskScore 获取设备风险评分
func (c *RiskContract) GetRiskScore(ctx contractapi.TransactionContextInterface, did string) (float64, error) {
	// 验证DID格式
	if !utils.ValidateDID(did) {
		return 0.0, fmt.Errorf("无效的DID格式: %s", did)
	}
	
	// 从账本中读取设备信息
	deviceInfoJSON, err := ctx.GetStub().GetState(did)
	if err != nil {
		return 0.0, fmt.Errorf("读取设备信息时出错: %v", err)
	}
	if deviceInfoJSON == nil {
		return 0.0, fmt.Errorf("设备DID %s 不存在", did)
	}
	
	// 反序列化设备信息
	var deviceInfo models.DeviceInfo
	err = json.Unmarshal(deviceInfoJSON, &deviceInfo)
	if err != nil {
		return 0.0, fmt.Errorf("设备信息反序列化失败: %v", err)
	}
	
	return deviceInfo.RiskScore, nil
}

// GetAttackProfile 获取设备攻击画像
func (c *RiskContract) GetAttackProfile(ctx contractapi.TransactionContextInterface, did string) (string, error) {
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
	
	// 反序列化设备信息
	var deviceInfo models.DeviceInfo
	err = json.Unmarshal(deviceInfoJSON, &deviceInfo)
	if err != nil {
		return "", fmt.Errorf("设备信息反序列化失败: %v", err)
	}
	
	// 序列化攻击画像
	attackProfileJSON, err := json.Marshal(deviceInfo.AttackProfile)
	if err != nil {
		return "", fmt.Errorf("攻击画像序列化失败: %v", err)
	}
	
	return string(attackProfileJSON), nil
}

// CheckDeviceConnectionEligibility 检查设备是否有资格连接
func (c *RiskContract) CheckDeviceConnectionEligibility(ctx contractapi.TransactionContextInterface, did string) (string, error) {
	// 获取设备风险评分
	riskScore, err := c.GetRiskScore(ctx, did)
	if err != nil {
		return "", err
	}
	
	// 根据风险评分返回响应策略
	if riskScore >= 700 {
		// 高危（700-1000）：硬性阻断
		return "高危风险，禁止设备连接", nil
	} else if riskScore >= 200 && riskScore < 700 {
		// 警戒（200-699）：主动欺骗与隔离引导
		return "警戒风险，将设备流量重定向至隔离的蜜网环境", nil
	} else if riskScore >= 1 && riskScore < 200 {
		// 关注（1-199）：增强监控，主动引诱
		return "关注风险，增强监控，主动引诱", nil
	} else {
		// 常规（0）：标准化信任与监控
		return "常规风险，设备可以正常连接", nil
	}
}

// GetHighRiskDevices 获取高风险设备
func (c *RiskContract) GetHighRiskDevices(ctx contractapi.TransactionContextInterface) ([]*models.DeviceInfo, error) {
	// 获取所有设备的迭代器
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, fmt.Errorf("获取状态范围时出错: %v", err)
	}
	defer resultsIterator.Close()
	
	// 存储高风险设备的切片
	var highRiskDevices []*models.DeviceInfo
	
	// 遍历所有状态
	for resultsIterator.HasNext() {
		// 获取下一个键值对
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, fmt.Errorf("获取下一个状态时出错: %v", err)
		}
		
		// 反序列化设备信息
		var deviceInfo models.DeviceInfo
		err = json.Unmarshal(queryResponse.Value, &deviceInfo)
		if err != nil {
			// 跳过非设备信息的状态
			continue
		}
		
		// 检查设备是否为高风险设备
		if deviceInfo.RiskScore >= 700 {
			highRiskDevices = append(highRiskDevices, &deviceInfo)
		}
	}
	
	return highRiskDevices, nil
}

// GetDevicesByRiskScoreRange 获取特定风险评分范围内的设备
func (c *RiskContract) GetDevicesByRiskScoreRange(ctx contractapi.TransactionContextInterface, minScoreStr, maxScoreStr string) ([]*models.DeviceInfo, error) {
	// 将字符串转换为浮点数
	minScore, err := strconv.ParseFloat(minScoreStr, 64)
	if err != nil {
		return nil, fmt.Errorf("最小风险评分格式无效: %v", err)
	}
	
	maxScore, err := strconv.ParseFloat(maxScoreStr, 64)
	if err != nil {
		return nil, fmt.Errorf("最大风险评分格式无效: %v", err)
	}
	
	// 验证风险评分范围
	if minScore < 0 || minScore > maxScore {
		return nil, fmt.Errorf("风险评分范围无效")
	}
	
	// 获取所有设备的迭代器
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, fmt.Errorf("获取状态范围时出错: %v", err)
	}
	defer resultsIterator.Close()
	
	// 存储符合条件的设备的切片
	var devices []*models.DeviceInfo
	
	// 遍历所有状态
	for resultsIterator.HasNext() {
		// 获取下一个键值对
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, fmt.Errorf("获取下一个状态时出错: %v", err)
		}
		
		// 反序列化设备信息
		var deviceInfo models.DeviceInfo
		err = json.Unmarshal(queryResponse.Value, &deviceInfo)
		if err != nil {
			// 跳过非设备信息的状态
			continue
		}
		
		// 检查设备是否在指定风险评分范围内
		if deviceInfo.RiskScore >= minScore && deviceInfo.RiskScore <= maxScore {
			devices = append(devices, &deviceInfo)
		}
	}
	
	return devices, nil
}

// GetDeviceRiskResponse 获取设备风险响应策略
func (c *RiskContract) GetDeviceRiskResponse(ctx contractapi.TransactionContextInterface, did string) (string, error) {
	// 获取设备风险评分
	riskScore, err := c.GetRiskScore(ctx, did)
	if err != nil {
		return "", err
	}
	
	// 根据风险评分返回响应策略
	var response string
	if riskScore >= 700 {
		// 高危（700-1000）：硬性阻断
		response = `{
			"riskLevel": "高危",
			"riskScore": ` + fmt.Sprintf("%.2f", riskScore) + `,
			"strategy": "硬性阻断",
			"measures": [
				"立即强制中断该设备所有已建立的网络连接",
				"临时锁定设备账户",
				"对其交互过的蜜点进行快照存证",
				"进入人工介入和深度溯源"
			]
		}`
	} else if riskScore >= 200 && riskScore < 700 {
		// 警戒（200-699）：主动欺骗与隔离引导
		response = `{
			"riskLevel": "警戒",
			"riskScore": ` + fmt.Sprintf("%.2f", riskScore) + `,
			"strategy": "主动欺骗与隔离引导",
			"measures": [
				"将该设备的所有内部域名解析请求指向对应的伪造服务",
				"在网络层将其流量重定向至一个隔离的蜜网环境中",
				"在其所处的蜜网环境中，动态植入伪造的凭证文件、数据库连接字符串、API密钥等蜜点"
			]
		}`
	} else if riskScore >= 1 && riskScore < 200 {
		// 关注（1-199）：增强监控，主动引诱
		response = `{
			"riskLevel": "关注",
			"riskScore": ` + fmt.Sprintf("%.2f", riskScore) + `,
			"strategy": "增强监控，主动引诱，无感知的情报收集",
			"measures": [
				"针对该设备的源IP，自动启用全数据包捕获",
				"详细记录其在蜜点中的所有操作行为",
				"实施轻微的服务质量策略，对其扫描或连接行为进行速率限制",
				"在其当前互动的蜜点环境中，主动暴露更具吸引力的诱饵"
			]
		}`
	} else {
		// 常规（0）：标准化信任与监控
		response = `{
			"riskLevel": "常规",
			"riskScore": ` + fmt.Sprintf("%.2f", riskScore) + `,
			"strategy": "标准化信任与监控",
			"measures": [
				"维持默认的日志记录",
				"仅在会话建立或关键操作时通过智能合约验证其DID身份的有效性"
			]
		}`
	}
	
	return response, nil
}