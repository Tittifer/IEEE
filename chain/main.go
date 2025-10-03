package main

import (
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/Tittifer/IEEE/chain/contracts"
)

func main() {
	// 创建身份认证合约
	identityContract := new(contracts.IdentityContract)
	
	// 创建风险管理合约
	riskContract := new(contracts.RiskContract)
	
	// 创建链码
	chaincode, err := contractapi.NewChaincode(identityContract, riskContract)
	if err != nil {
		fmt.Printf("创建链码时出错: %s", err.Error())
		return
	}

	// 启动链码
	if err := chaincode.Start(); err != nil {
		fmt.Printf("启动链码时出错: %s", err.Error())
	}
}