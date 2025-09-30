package main

import (
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/Tittifer/IEEE/sidechain/contracts"
)

func main() {
	// 创建DID合约
	didContract := new(contracts.DIDContract)
	
	// 创建会话管理合约
	sessionContract := new(contracts.SessionContract)
	
	// 创建风险评估合约
	riskContract := new(contracts.RiskContract)
	
	// 创建链码
	chaincode, err := contractapi.NewChaincode(didContract, sessionContract, riskContract)
	if err != nil {
		fmt.Printf("创建链码时出错: %s", err.Error())
		return
	}

	// 启动链码
	if err := chaincode.Start(); err != nil {
		fmt.Printf("启动链码时出错: %s", err.Error())
	}
}