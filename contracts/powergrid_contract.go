package contracts

import (
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// PowerGridContract 电网场景下的主合约，整合了用户管理和风险评估功能
type PowerGridContract struct {
	contractapi.Contract
	UserContract
	RiskContract
}

// InitLedger 初始化账本
func (c *PowerGridContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	fmt.Println("电网身份认证与风险评估链码初始化")
	return nil
}
