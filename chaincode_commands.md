# 电网身份认证链码命令指南

本文档提供了与电网身份认证链码交互的常用命令。

## 环境设置

在使用以下命令前，请先设置环境变量：

```bash
source env.sh
```

## 用户管理命令

### 注册新用户

```bash
peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile ${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem -C mychannel -n powercc --peerAddresses localhost:7051 --tlsRootCertFiles ${PWD}/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt --peerAddresses localhost:9051 --tlsRootCertFiles ${PWD}/organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt -c '{"function":"RegisterUser","Args":["did:example:123", "张三", "110101199001011234", "公钥内容..."]}'
```

### 获取用户信息

```bash
peer chaincode query -C mychannel -n powercc -c '{"function":"GetUser","Args":["did:example:123"]}'
```

### 验证用户身份

```bash
peer chaincode query -C mychannel -n powercc -c '{"function":"VerifyUser","Args":["did:example:123"]}'
```

### 更改用户状态

```bash
peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile ${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem -C mychannel -n powercc --peerAddresses localhost:7051 --tlsRootCertFiles ${PWD}/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt --peerAddresses localhost:9051 --tlsRootCertFiles ${PWD}/organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt -c '{"function":"ChangeUserStatus","Args":["did:example:123", "suspended"]}'
```

### 获取所有用户

```bash
peer chaincode query -C mychannel -n powercc -c '{"function":"GetAllUsers","Args":[]}'
```

## 风险管理命令

### 记录攻击行为

```bash
peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile ${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem -C mychannel -n powercc --peerAddresses localhost:7051 --tlsRootCertFiles ${PWD}/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt --peerAddresses localhost:9051 --tlsRootCertFiles ${PWD}/organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt -c '{"function":"RecordAttack","Args":["did:example:123", "honeypot001", "SQL注入", "尝试通过SQL注入获取数据库访问权限", "8"]}'
```

### 更新风险评分

```bash
peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile ${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem -C mychannel -n powercc --peerAddresses localhost:7051 --tlsRootCertFiles ${PWD}/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt --peerAddresses localhost:9051 --tlsRootCertFiles ${PWD}/organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt -c '{"function":"UpdateRiskScore","Args":["did:example:123", "50"]}'
```

### 获取高风险用户

```bash
peer chaincode query -C mychannel -n powercc -c '{"function":"GetHighRiskUsers","Args":[]}'
```

### 获取特定风险评分范围内的用户

```bash
peer chaincode query -C mychannel -n powercc -c '{"function":"GetUsersByRiskScore","Args":["40", "70"]}'
```

## 关闭网络

当测试完成后，可以使用以下命令关闭网络：

```bash
cd $FABRIC_SAMPLES_PATH/test-network
./network.sh down
```
