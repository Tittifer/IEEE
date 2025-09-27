package contracts

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/Tittifer/IEEE/models"
	"github.com/Tittifer/IEEE/utils"
)

// UserContract 用户管理合约
type UserContract struct {
	contractapi.Contract
}

// RegisterUser 注册新用户
func (c *UserContract) RegisterUser(ctx contractapi.TransactionContextInterface, did, name, idNumber, publicKey string) error {
	// 检查用户是否已存在
	exists, err := c.UserExists(ctx, did)
	if err != nil {
		return fmt.Errorf("检查用户是否存在时出错: %v", err)
	}
	if exists {
		return fmt.Errorf("用户DID %s 已存在", did)
	}

	// 使用交易时间戳而不是当前时间，确保确定性
	timestamp, err := ctx.GetStub().GetTxTimestamp()
	if err != nil {
		return fmt.Errorf("获取交易时间戳失败: %v", err)
	}
	txTime := time.Unix(timestamp.Seconds, int64(timestamp.Nanos))
	
	// 创建新用户
	userInfo := models.UserInfo{
		DID:           did,
		Name:          name,
		IDNumber:      idNumber,
		PublicKey:     publicKey,
		RiskScore:     0,                // 初始风险评分为0
		AccessLevel:   models.AccessLevelHighest, // 初始访问级别为1（最高权限）
		Status:        models.StatusActive,       // 初始状态为活跃
		AttackHistory: []models.Attack{},         // 初始攻击历史为空
		CreatedAt:     txTime,
		LastUpdatedAt: txTime,
	}

	// 创建可验证凭证
	credential, err := utils.CreateVerifiableCredential(userInfo)
	if err != nil {
		return fmt.Errorf("创建可验证凭证失败: %v", err)
	}
	
	// 将凭证序列化为JSON
	credentialJSON, err := json.Marshal(credential)
	if err != nil {
		return fmt.Errorf("凭证序列化失败: %v", err)
	}
	
	// 将凭证存储到用户信息中
	userInfo.Credential = string(credentialJSON)

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

	return nil
}

// RegisterUserWithVehicle 注册新用户并生成可验证凭证
func (c *UserContract) RegisterUserWithVehicle(ctx contractapi.TransactionContextInterface, name, idNumber, phoneNumber, vehiclePlate, vehicleModel, vehicleColor, vehicleVIN, publicKey string) error {
	// 根据用户信息生成DID
	did := utils.GenerateDID(name, idNumber)
	
	// 检查用户是否已存在
	exists, err := c.UserExists(ctx, did)
	if err != nil {
		return fmt.Errorf("检查用户是否存在时出错: %v", err)
	}
	if exists {
		return fmt.Errorf("用户DID %s 已存在", did)
	}

	// 使用交易时间戳而不是当前时间，确保确定性
	timestamp, err := ctx.GetStub().GetTxTimestamp()
	if err != nil {
		return fmt.Errorf("获取交易时间戳失败: %v", err)
	}
	txTime := time.Unix(timestamp.Seconds, int64(timestamp.Nanos))
	
	// 创建新用户
	userInfo := models.UserInfo{
		DID:           did,
		Name:          name,
		IDNumber:      idNumber,
		PhoneNumber:   phoneNumber,
		VehiclePlate:  vehiclePlate,
		VehicleModel:  vehicleModel,
		VehicleColor:  vehicleColor,
		VehicleVIN:    vehicleVIN,
		PublicKey:     publicKey,
		RiskScore:     0,                // 初始风险评分为0
		AccessLevel:   models.AccessLevelHighest, // 初始访问级别为1（最高权限）
		Status:        models.StatusActive,       // 初始状态为活跃
		AttackHistory: []models.Attack{},         // 初始攻击历史为空
		CreatedAt:     txTime,
		LastUpdatedAt: txTime,
	}

	// 创建可验证凭证
	credential, err := utils.CreateVerifiableCredential(userInfo)
	if err != nil {
		return fmt.Errorf("创建可验证凭证失败: %v", err)
	}
	
	// 将凭证序列化为JSON
	credentialJSON, err := json.Marshal(credential)
	if err != nil {
		return fmt.Errorf("凭证序列化失败: %v", err)
	}
	
	// 将凭证存储到用户信息中
	userInfo.Credential = string(credentialJSON)

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

	return nil
}

// GetUser 获取用户信息
func (c *UserContract) GetUser(ctx contractapi.TransactionContextInterface, did string) (*models.UserInfo, error) {
	// 从账本中读取用户信息
	userInfoJSON, err := ctx.GetStub().GetState(did)
	if err != nil {
		return nil, fmt.Errorf("读取用户信息时出错: %v", err)
	}
	if userInfoJSON == nil {
		return nil, fmt.Errorf("用户DID %s 不存在", did)
	}

	// 反序列化用户信息
	var userInfo models.UserInfo
	err = json.Unmarshal(userInfoJSON, &userInfo)
	if err != nil {
		return nil, fmt.Errorf("用户信息反序列化失败: %v", err)
	}

	return &userInfo, nil
}

// VerifyUser 验证用户身份
func (c *UserContract) VerifyUser(ctx contractapi.TransactionContextInterface, did string) (bool, error) {
	// 获取用户信息
	userInfo, err := c.GetUser(ctx, did)
	if err != nil {
		return false, err
	}

	// 检查用户状态是否为活跃
	if userInfo.Status != models.StatusActive {
		return false, fmt.Errorf("用户状态为 %s, 非活跃状态", userInfo.Status)
	}

	// 检查用户风险评分是否超过阈值
	if userInfo.RiskScore >= models.RiskThresholdHigh {
		return false, fmt.Errorf("用户风险评分为 %d, 超过安全阈值", userInfo.RiskScore)
	}

	return true, nil
}

// ChangeUserStatus 更改用户状态
func (c *UserContract) ChangeUserStatus(ctx contractapi.TransactionContextInterface, did, newStatus string) error {
	// 检查状态值是否有效
	if newStatus != models.StatusActive && newStatus != models.StatusSuspended && newStatus != models.StatusBlocked {
		return fmt.Errorf("无效的状态值: %s, 必须是 active、suspended 或 blocked", newStatus)
	}

	// 获取用户信息
	userInfo, err := c.GetUser(ctx, did)
	if err != nil {
		return err
	}

	// 更新状态
	userInfo.Status = newStatus
	
	// 使用交易时间戳而不是当前时间，确保确定性
	timestamp, err := ctx.GetStub().GetTxTimestamp()
	if err != nil {
		return fmt.Errorf("获取交易时间戳失败: %v", err)
	}
	userInfo.LastUpdatedAt = time.Unix(timestamp.Seconds, int64(timestamp.Nanos))

	// 将更新后的用户信息转换为JSON并存储
	userInfoJSON, err := json.Marshal(userInfo)
	if err != nil {
		return fmt.Errorf("用户信息序列化失败: %v", err)
	}

	// 将更新后的用户信息写入账本
	err = ctx.GetStub().PutState(did, userInfoJSON)
	if err != nil {
		return fmt.Errorf("更新用户状态时出错: %v", err)
	}

	return nil
}

// UserExists 检查用户是否存在
func (c *UserContract) UserExists(ctx contractapi.TransactionContextInterface, did string) (bool, error) {
	userInfoJSON, err := ctx.GetStub().GetState(did)
	if err != nil {
		return false, fmt.Errorf("读取用户信息时出错: %v", err)
	}
	return userInfoJSON != nil, nil
}

// GetAllUsers 获取所有用户
func (c *UserContract) GetAllUsers(ctx contractapi.TransactionContextInterface) ([]*models.UserInfo, error) {
	// 获取所有用户的迭代器
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, fmt.Errorf("获取用户迭代器时出错: %v", err)
	}
	defer resultsIterator.Close()

	var users []*models.UserInfo
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, fmt.Errorf("迭代用户时出错: %v", err)
		}

		var user models.UserInfo
		err = json.Unmarshal(queryResponse.Value, &user)
		if err != nil {
			return nil, fmt.Errorf("用户信息反序列化失败: %v", err)
		}
		users = append(users, &user)
	}

	return users, nil
}

// GetUserCredential 获取用户的可验证凭证
func (c *UserContract) GetUserCredential(ctx contractapi.TransactionContextInterface, did string) (*models.VerifiableCredential, error) {
	// 获取用户信息
	userInfo, err := c.GetUser(ctx, did)
	if err != nil {
		return nil, err
	}
	
	// 检查用户是否有凭证
	if userInfo.Credential == "" {
		return nil, fmt.Errorf("用户 %s 没有可验证凭证", did)
	}
	
	// 反序列化凭证
	var credential models.VerifiableCredential
	err = json.Unmarshal([]byte(userInfo.Credential), &credential)
	if err != nil {
		return nil, fmt.Errorf("凭证反序列化失败: %v", err)
	}
	
	return &credential, nil
}

// RevokeCredential 吊销用户凭证
func (c *UserContract) RevokeCredential(ctx contractapi.TransactionContextInterface, did string) error {
	// 获取用户信息
	userInfo, err := c.GetUser(ctx, did)
	if err != nil {
		return err
	}
	
	// 检查用户是否有凭证
	if userInfo.Credential == "" {
		return fmt.Errorf("用户 %s 没有可验证凭证", did)
	}
	
	// 反序列化凭证
	var credential models.VerifiableCredential
	err = json.Unmarshal([]byte(userInfo.Credential), &credential)
	if err != nil {
		return fmt.Errorf("凭证反序列化失败: %v", err)
	}
	
	// 更新凭证状态
	credential.CredentialStatus.Type = models.CredentialStatusRevoked
	
	// 使用交易时间戳
	timestamp, err := ctx.GetStub().GetTxTimestamp()
	if err != nil {
		return fmt.Errorf("获取交易时间戳失败: %v", err)
	}
	txTime := time.Unix(timestamp.Seconds, int64(timestamp.Nanos))
	
	// 重新签名凭证
	proof, err := utils.SignCredential(&credential)
	if err != nil {
		return fmt.Errorf("签名凭证失败: %v", err)
	}
	credential.Proof = *proof
	
	// 将凭证序列化为JSON
	credentialJSON, err := json.Marshal(credential)
	if err != nil {
		return fmt.Errorf("凭证序列化失败: %v", err)
	}
	
	// 更新用户信息
	userInfo.Credential = string(credentialJSON)
	userInfo.LastUpdatedAt = txTime
	
	// 将用户信息转换为JSON并存储
	userInfoJSON, err := json.Marshal(userInfo)
	if err != nil {
		return fmt.Errorf("用户信息序列化失败: %v", err)
	}
	
	// 将用户信息写入账本
	err = ctx.GetStub().PutState(did, userInfoJSON)
	if err != nil {
		return fmt.Errorf("更新用户信息时出错: %v", err)
	}
	
	return nil
}
