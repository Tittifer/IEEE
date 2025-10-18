package main

import (
	"log"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/Tittifer/IEEE/chain/contracts"
)

func main() {
	// 创建链码
	identityContract := new(contracts.IdentityContract)
	riskContract := new(contracts.RiskContract)

	// 创建链码容器
	cc, err := contractapi.NewChaincode(identityContract, riskContract)
	if err != nil {
		log.Panicf("创建IEEE链码失败: %v", err)
	}

	// 启动链码
	if err := cc.Start(); err != nil {
		log.Panicf("启动IEEE链码失败: %v", err)
	}
}