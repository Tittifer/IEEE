package contracts

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/Tittifer/IEEE/models"
	"github.com/Tittifer/IEEE/utils"
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

	// 使用交易时间戳而不是当前时间，确保确定性
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
		IDNumber:      idNumber,
		PhoneNumber:   phoneNumber,
		VehicleID:     vehicleID,
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

	return nil
}

// GetUser 获取用户信息
func (c *IdentityContract) GetUser(ctx contractapi.TransactionContextInterface, did string) (*models.UserInfo, error) {
	// 验证DID格式
	if !utils.ValidateDID(did) {
		return nil, fmt.Errorf("无效的DID格式: %s", did)
	}
	
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

// GetUserByInfo 根据用户信息获取用户
func (c *IdentityContract) GetUserByInfo(ctx contractapi.TransactionContextInterface, name, idNumber, phoneNumber, vehicleID string) (*models.UserInfo, error) {
	// 根据用户信息生成DID
	did := utils.GenerateDID(name, idNumber, phoneNumber, vehicleID)
	
	// 调用GetUser函数获取用户信息
	return c.GetUser(ctx, did)
}

// VerifyIdentity 验证用户身份
func (c *IdentityContract) VerifyIdentity(ctx contractapi.TransactionContextInterface, did, name, idNumber, phoneNumber, vehicleID string) (bool, error) {
	// 验证DID格式
	if !utils.ValidateDID(did) {
		return false, fmt.Errorf("无效的DID格式: %s", did)
	}
	
	// 获取用户信息
	userInfo, err := c.GetUser(ctx, did)
	if err != nil {
		return false, err
	}
	
	// 验证用户信息是否匹配
	if userInfo.Name != name || 
	   userInfo.IDNumber != idNumber || 
	   userInfo.PhoneNumber != phoneNumber || 
	   userInfo.VehicleID != vehicleID {
		return false, nil
	}
	
	// 验证用户状态是否为活跃
	if userInfo.Status != models.StatusActive {
		return false, fmt.Errorf("用户状态为 %s, 非活跃状态", userInfo.Status)
	}
	
	return true, nil
}

// GetAllUsers 获取所有用户
func (c *IdentityContract) GetAllUsers(ctx contractapi.TransactionContextInterface) ([]*models.UserInfo, error) {
	// 获取所有用户的迭代器
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, fmt.Errorf("获取状态范围时出错: %v", err)
	}
	defer resultsIterator.Close()

	// 存储所有用户的切片
	var users []*models.UserInfo

	// 遍历所有状态
	for resultsIterator.HasNext() {
		// 获取下一个键值对
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, fmt.Errorf("获取下一个状态时出错: %v", err)
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

	return users, nil
}
