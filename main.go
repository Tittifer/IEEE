package main

import (
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/Tittifer/IEEE/contracts"
)

func main() {
	// 创建智能合约
	powerGridContract := new(contracts.PowerGridContract)
	
	// 创建链码
	chaincode, err := contractapi.NewChaincode(powerGridContract)
	if err != nil {
		fmt.Printf("创建电网身份认证链码时出错: %s", err.Error())
		return
	}

	// 启动链码
	if err := chaincode.Start(); err != nil {
		fmt.Printf("启动电网身份认证链码时出错: %s", err.Error())
	}
}