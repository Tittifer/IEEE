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

// 引用风险合约中的常量

// InitLedger 初始化账本
func (c *IdentityContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	fmt.Println("身份认证链码初始化")
	return nil
}

// RegisterUser 注册新用户
func (c *IdentityContract) RegisterUser(ctx contractapi.TransactionContextInterface, name, idNumber, phoneNumber, vehicleID string) error {
	// 根据用户信息生成DID
	did := utils.GenerateDID(name, idNumber, phoneNumber, vehicleID)
	
	// 检查用户是否已存在
	exists, err := c.UserExists(ctx, did)
	if err != nil {
		return fmt.Errorf("检查用户是否存在时出错: %v", err)
	}
	if exists {
		return fmt.Errorf("用户DID %s 已存在", did)
	}

	// 使用交易时间戳确保确定性
	timestamp, err := ctx.GetStub().GetTxTimestamp()
	if err != nil {
		return fmt.Errorf("获取交易时间戳失败: %v", err)
	}
	txTime := time.Unix(timestamp.Seconds, int64(timestamp.Nanos))
	timeStr := txTime.Format(time.RFC3339)
	
	// 计算初始风险值
	initialRiskScore := utils.CalculateInitialRiskScore()
	
	// 创建新用户
	userInfo := models.UserInfo{
		DID:           did,
		Name:          name,
		RiskScore:     initialRiskScore,
		Status:        models.StatusActive,
		CreatedAt:     timeStr,
		LastUpdatedAt: timeStr,
	}

	// 将用户信息转换为JSON并存储
	userInfoJSON, err := json.Marshal(userInfo)
	if err != nil {
		return fmt.Errorf("用户信息序列化失败: %v", err)
	}

	// 将用户信息写入账本
	err = ctx.GetStub().PutState(did, userInfoJSON)
	if err != nil {
		return fmt.Errorf("存储用户信息时出错: %v", err)
	}

	// 创建用户注册事件
	registerEvent := models.UserEvent{
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

	// 发送用户注册事件
	err = ctx.GetStub().SetEvent("UserRegistered", eventJSON)
	if err != nil {
		return fmt.Errorf("发送用户注册事件失败: %v", err)
	}

	return nil
}

// GetUser 获取用户信息
func (c *IdentityContract) GetUser(ctx contractapi.TransactionContextInterface, did string) (string, error) {
	// 验证DID格式
	if !utils.ValidateDID(did) {
		return "", fmt.Errorf("无效的DID格式: %s", did)
	}
	
	// 从账本中读取用户信息
	userInfoJSON, err := ctx.GetStub().GetState(did)
	if err != nil {
		return "", fmt.Errorf("读取用户信息时出错: %v", err)
	}
	if userInfoJSON == nil {
		return "", fmt.Errorf("用户DID %s 不存在", did)
	}

	// 直接返回JSON字符串，无需再次序列化
	return string(userInfoJSON), nil
}

// UserExists 检查用户是否存在
func (c *IdentityContract) UserExists(ctx contractapi.TransactionContextInterface, did string) (bool, error) {
	// 从账本中读取用户信息
	userInfoJSON, err := ctx.GetStub().GetState(did)
	if err != nil {
		return false, fmt.Errorf("读取用户信息时出错: %v", err)
	}
	
	// 如果用户信息不为空，则用户存在
	return userInfoJSON != nil, nil
}

// GetDIDByInfo 根据用户信息生成DID
func (c *IdentityContract) GetDIDByInfo(ctx contractapi.TransactionContextInterface, name, idNumber, phoneNumber, vehicleID string) (string, error) {
	// 根据用户信息生成DID
	did := utils.GenerateDID(name, idNumber, phoneNumber, vehicleID)
	return did, nil
}

// VerifyIdentity 验证用户身份
func (c *IdentityContract) VerifyIdentity(ctx contractapi.TransactionContextInterface, did, name string) (bool, error) {
	// 验证DID格式
	if !utils.ValidateDID(did) {
		return false, fmt.Errorf("无效的DID格式: %s", did)
	}
	
	// 从账本中读取用户信息
	userInfoJSON, err := ctx.GetStub().GetState(did)
	if err != nil {
		return false, fmt.Errorf("读取用户信息时出错: %v", err)
	}
	if userInfoJSON == nil {
		return false, fmt.Errorf("用户DID %s 不存在", did)
	}

	// 反序列化用户信息
	var userInfo models.UserInfo
	err = json.Unmarshal(userInfoJSON, &userInfo)
	if err != nil {
		return false, fmt.Errorf("用户信息反序列化失败: %v", err)
	}
	
	// 验证用户姓名是否匹配
	if userInfo.Name != name {
		return false, nil
	}
	
	// 验证用户状态是否为活跃
	if userInfo.Status != models.StatusActive {
		return false, fmt.Errorf("用户状态为 %s, 非活跃状态", userInfo.Status)
	}
	
	return true, nil
}

// UserLogin 用户登录
func (c *IdentityContract) UserLogin(ctx contractapi.TransactionContextInterface, did, name string) (string, error) {
	// 验证DID格式
	if !utils.ValidateDID(did) {
		return "", fmt.Errorf("无效的DID格式: %s", did)
	}
	
	// 获取用户信息
	userInfoJSON, err := ctx.GetStub().GetState(did)
	if err != nil {
		return "", fmt.Errorf("读取用户信息时出错: %v", err)
	}
	if userInfoJSON == nil {
		// 用户不存在，返回登录失败信息
		return "登录失败：用户不存在", nil
	}
	
	// 反序列化用户信息
	var userInfo models.UserInfo
	err = json.Unmarshal(userInfoJSON, &userInfo)
	if err != nil {
		return "", fmt.Errorf("用户信息反序列化失败: %v", err)
	}
	
	// 验证用户姓名是否匹配
	if userInfo.Name != name {
		// 姓名不匹配，返回登录失败信息
		return "登录失败：用户名不匹配", nil
	}
	
	// 验证用户状态是否为活跃
	if userInfo.Status != models.StatusActive {
		// 用户状态不活跃，返回登录失败信息
		return "登录失败：用户状态不活跃", nil
	}
	
	// 检查用户风险评分
	if userInfo.RiskScore > models.RiskScoreThreshold {
		// 风险评分过高，禁止登录
		return "禁止用户登录", nil
	}
	
	// 更新用户状态为在线
	userInfo.Status = models.StatusOnline
	
	// 使用交易时间戳更新最后更新时间
	timestamp, err := ctx.GetStub().GetTxTimestamp()
	if err != nil {
		return "", fmt.Errorf("获取交易时间戳失败: %v", err)
	}
	txTime := time.Unix(timestamp.Seconds, int64(timestamp.Nanos))
	userInfo.LastUpdatedAt = txTime.Format(time.RFC3339)
	
	// 将更新后的用户信息保存到账本
	userInfoJSON, err = json.Marshal(userInfo)
	if err != nil {
		return "", fmt.Errorf("用户信息序列化失败: %v", err)
	}
	
	err = ctx.GetStub().PutState(did, userInfoJSON)
	if err != nil {
		return "", fmt.Errorf("更新用户信息时出错: %v", err)
	}
	
	// 创建用户登录事件
	loginEvent := models.UserEvent{
		EventType: models.EventTypeLogin,
		DID:       did,
		Name:      name,
		Timestamp: txTime.Unix(),
	}

	// 序列化事件数据
	eventJSON, err := json.Marshal(loginEvent)
	if err != nil {
		return "", fmt.Errorf("事件数据序列化失败: %v", err)
	}

	// 发送用户登录事件
	err = ctx.GetStub().SetEvent("UserLoggedIn", eventJSON)
	if err != nil {
		return "", fmt.Errorf("发送用户登录事件失败: %v", err)
	}
	
	// 登录成功
	return "登录成功", nil
}

// UserLogout 用户登出
func (c *IdentityContract) UserLogout(ctx contractapi.TransactionContextInterface, did string) (string, error) {
	// 验证DID格式
	if !utils.ValidateDID(did) {
		return "", fmt.Errorf("无效的DID格式: %s", did)
	}
	
	// 获取用户信息
	userInfoJSON, err := ctx.GetStub().GetState(did)
	if err != nil {
		return "", fmt.Errorf("读取用户信息时出错: %v", err)
	}
	if userInfoJSON == nil {
		return "", fmt.Errorf("用户DID %s 不存在", did)
	}
	
	// 反序列化用户信息
	var userInfo models.UserInfo
	err = json.Unmarshal(userInfoJSON, &userInfo)
	if err != nil {
		return "", fmt.Errorf("用户信息反序列化失败: %v", err)
	}
	
	// 检查用户是否已经登出
	if userInfo.Status != models.StatusOnline {
		return "用户未登录，无需登出", nil
	}
	
	// 更新用户状态为离线
	userInfo.Status = models.StatusOffline
	
	// 使用交易时间戳更新最后更新时间
	timestamp, err := ctx.GetStub().GetTxTimestamp()
	if err != nil {
		return "", fmt.Errorf("获取交易时间戳失败: %v", err)
	}
	txTime := time.Unix(timestamp.Seconds, int64(timestamp.Nanos))
	userInfo.LastUpdatedAt = txTime.Format(time.RFC3339)
	
	// 将更新后的用户信息保存到账本
	userInfoJSON, err = json.Marshal(userInfo)
	if err != nil {
		return "", fmt.Errorf("用户信息序列化失败: %v", err)
	}
	
	err = ctx.GetStub().PutState(did, userInfoJSON)
	if err != nil {
		return "", fmt.Errorf("更新用户信息时出错: %v", err)
	}
	
	// 创建用户登出事件
	logoutEvent := models.UserEvent{
		EventType: models.EventTypeLogout,
		DID:       did,
		Name:      userInfo.Name,
		Timestamp: txTime.Unix(),
	}

	// 序列化事件数据
	eventJSON, err := json.Marshal(logoutEvent)
	if err != nil {
		return "", fmt.Errorf("事件数据序列化失败: %v", err)
	}

	// 发送用户登出事件
	err = ctx.GetStub().SetEvent("UserLoggedOut", eventJSON)
	if err != nil {
		return "", fmt.Errorf("发送用户登出事件失败: %v", err)
	}
	
	// 登出成功
	return "登出成功", nil
}

// UpdateRiskScore 更新用户风险评分
func (c *IdentityContract) UpdateRiskScore(ctx contractapi.TransactionContextInterface, did string, riskScoreStr string) error {
	// 验证DID格式
	if !utils.ValidateDID(did) {
		return fmt.Errorf("无效的DID格式: %s", did)
	}
	
	// 解析风险评分
	riskScore, err := strconv.Atoi(riskScoreStr)
	if err != nil {
		return fmt.Errorf("无效的风险评分格式: %s", riskScoreStr)
	}
	
	// 获取用户信息
	userInfoJSON, err := ctx.GetStub().GetState(did)
	if err != nil {
		return fmt.Errorf("读取用户信息时出错: %v", err)
	}
	if userInfoJSON == nil {
		return fmt.Errorf("用户DID %s 不存在", did)
	}
	
	// 反序列化用户信息
	var userInfo models.UserInfo
	err = json.Unmarshal(userInfoJSON, &userInfo)
	if err != nil {
		return fmt.Errorf("用户信息反序列化失败: %v", err)
	}
	
	// 更新风险评分
	userInfo.RiskScore = riskScore
	
	// 如果风险评分超过阈值，更新用户状态为风险状态
	if riskScore > models.RiskScoreThreshold && userInfo.Status != models.StatusRisky {
		userInfo.Status = models.StatusRisky
	}
	
	// 使用交易时间戳更新最后更新时间
	timestamp, err := ctx.GetStub().GetTxTimestamp()
	if err != nil {
		return fmt.Errorf("获取交易时间戳失败: %v", err)
	}
	txTime := time.Unix(timestamp.Seconds, int64(timestamp.Nanos))
	userInfo.LastUpdatedAt = txTime.Format(time.RFC3339)
	
	// 将更新后的用户信息保存到账本
	userInfoJSON, err = json.Marshal(userInfo)
	if err != nil {
		return fmt.Errorf("用户信息序列化失败: %v", err)
	}
	
	err = ctx.GetStub().PutState(did, userInfoJSON)
	if err != nil {
		return fmt.Errorf("更新用户信息时出错: %v", err)
	}
	
	// 创建风险评分更新事件
	riskScoreEvent := models.UserEvent{
		EventType: models.EventTypeRiskUpdate,
		DID:       did,
		Name:      userInfo.Name,
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

// GetAllUsers 获取所有用户
func (c *IdentityContract) GetAllUsers(ctx contractapi.TransactionContextInterface) (string, error) {
	// 获取所有用户的迭代器
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return "", fmt.Errorf("获取状态范围时出错: %v", err)
	}
	defer resultsIterator.Close()

	// 存储所有用户的切片
	var users []*models.UserInfo

	// 遍历所有状态
	for resultsIterator.HasNext() {
		// 获取下一个键值对
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return "", fmt.Errorf("获取下一个状态时出错: %v", err)
		}

		// 反序列化用户信息
		var userInfo models.UserInfo
		err = json.Unmarshal(queryResponse.Value, &userInfo)
		if err != nil {
			// 跳过非用户信息的状态
			continue
		}

		// 将用户信息添加到切片中
		users = append(users, &userInfo)
	}

	// 将用户列表转换为JSON字符串
	usersJSON, err := json.Marshal(users)
	if err != nil {
		return "", fmt.Errorf("序列化用户列表失败: %v", err)
	}

	return string(usersJSON), nil
}
