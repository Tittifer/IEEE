package contracts

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/Tittifer/IEEE/sidechain/models"
	"github.com/Tittifer/IEEE/sidechain/utils"
)

// DIDContract DID管理合约
type DIDContract struct {
	contractapi.Contract
}

// InitLedger 初始化账本
func (c *DIDContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	fmt.Println("DID管理链码初始化")
	return nil
}

// CreateDIDRecord 创建新的DID记录，初始风险评分和时间戳为0
func (c *DIDContract) CreateDIDRecord(ctx contractapi.TransactionContextInterface, did string) error {
	// 验证DID格式
	if !utils.ValidateDID(did) {
		return fmt.Errorf("无效的DID格式: %s", did)
	}

	// 检查DID是否已存在
	exists, err := c.DIDExists(ctx, did)
	if err != nil {
		return fmt.Errorf("检查DID是否存在时出错: %v", err)
	}
	if exists {
		return fmt.Errorf("DID %s 已存在", did)
	}

	// 创建新的DID记录，初始风险评分和时间戳为0
	record := models.DIDRiskRecord{
		DID:       did,
		RiskScore: models.InitialRiskScore,
		Timestamp: models.InitialTimestamp,
	}

	recordJSON, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("序列化DID记录时出错: %v", err)
	}

	// 将DID记录存储到账本中
	return ctx.GetStub().PutState(did, recordJSON)
}

// GetDIDRecord 根据DID获取记录
func (c *DIDContract) GetDIDRecord(ctx contractapi.TransactionContextInterface, did string) (*models.DIDRiskRecord, error) {
	// 验证DID格式
	if !utils.ValidateDID(did) {
		return nil, fmt.Errorf("无效的DID格式: %s", did)
	}

	recordJSON, err := ctx.GetStub().GetState(did)
	if err != nil {
		return nil, fmt.Errorf("获取DID记录时出错: %v", err)
	}
	if recordJSON == nil {
		return nil, fmt.Errorf("DID %s 不存在", did)
	}

	var record models.DIDRiskRecord
	err = json.Unmarshal(recordJSON, &record)
	if err != nil {
		return nil, fmt.Errorf("反序列化DID记录时出错: %v", err)
	}

	return &record, nil
}

// UpdateRiskScore 更新DID的风险评分和时间戳
func (c *DIDContract) UpdateRiskScore(ctx contractapi.TransactionContextInterface, did string, riskScore int, timestamp int64) error {
	// 验证DID格式
	if !utils.ValidateDID(did) {
		return fmt.Errorf("无效的DID格式: %s", did)
	}

	record, err := c.GetDIDRecord(ctx, did)
	if err != nil {
		return err
	}

	record.RiskScore = riskScore
	record.Timestamp = timestamp

	recordJSON, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("序列化更新后的DID记录时出错: %v", err)
	}

	return ctx.GetStub().PutState(did, recordJSON)
}

// DIDExists 检查DID是否存在
func (c *DIDContract) DIDExists(ctx contractapi.TransactionContextInterface, did string) (bool, error) {
	recordJSON, err := ctx.GetStub().GetState(did)
	if err != nil {
		return false, fmt.Errorf("查询账本时出错: %v", err)
	}

	return recordJSON != nil, nil
}

// GetAllDIDRecords 获取所有DID记录
func (c *DIDContract) GetAllDIDRecords(ctx contractapi.TransactionContextInterface) ([]*models.DIDRiskRecord, error) {
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, fmt.Errorf("获取所有DID记录时出错: %v", err)
	}
	defer resultsIterator.Close()

	var records []*models.DIDRiskRecord
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, fmt.Errorf("迭代DID记录时出错: %v", err)
		}

		var record models.DIDRiskRecord
		err = json.Unmarshal(queryResponse.Value, &record)
		if err != nil {
			// 跳过非DID记录的状态
			continue
		}
		records = append(records, &record)
	}

	return records, nil
}
