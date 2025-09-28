# 电网身份认证链码命令指南

本文档提供了与电网身份认证链码交互的常用命令。

## 环境设置

在使用以下命令前，请先设置环境变量：

```bash
source env.sh
```

## 链码基本命令

### 初始化账本

```bash
peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile ${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem -C mychannel -n powercc --peerAddresses localhost:7051 --tlsRootCertFiles ${PWD}/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt --peerAddresses localhost:9051 --tlsRootCertFiles ${PWD}/organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt -c '{"function":"IdentityContract:InitLedger","Args":[]}'
```

## 用户身份管理命令 (IdentityContract)

### 注册新用户 (RegisterUser)

```bash
peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile ${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem -C mychannel -n powercc --peerAddresses localhost:7051 --tlsRootCertFiles ${PWD}/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt --peerAddresses localhost:9051 --tlsRootCertFiles ${PWD}/organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt -c '{"function":"IdentityContract:RegisterUser","Args":["张三", "110101199001011234", "13800138000", "京A12345"]}'
```

### 获取用户信息 (GetUser)

```bash
peer chaincode query -C mychannel -n powercc -c '{"function":"IdentityContract:GetUser","Args":["did:example:123456789abcdef0"]}'
```

### 根据用户信息获取DID (GetDIDByInfo)

```bash
peer chaincode query -C mychannel -n powercc -c '{"function":"IdentityContract:GetDIDByInfo","Args":["张三", "110101199001011234", "13800138000", "京A12345"]}'
```

### 验证用户身份 (VerifyIdentity)

```bash
peer chaincode query -C mychannel -n powercc -c '{"function":"IdentityContract:VerifyIdentity","Args":["did:example:123456789abcdef0", "张三"]}'
```

### 用户登录 (UserLogin)

```bash
peer chaincode query -C mychannel -n powercc -c '{"function":"IdentityContract:UserLogin","Args":["did:example:123456789abcdef0", "张三"]}'
```

返回值说明：
- 如果返回 `"登录成功"` - 登录成功
- 如果返回 `"登录失败：用户不存在"` - 用户DID不存在
- 如果返回 `"登录失败：用户名不匹配"` - DID和姓名不匹配
- 如果返回 `"登录失败：用户状态不活跃"` - 用户状态非活跃
- 如果返回 `"禁止用户登录"` - 登录失败（风险评分超过阈值）

### 检查用户是否存在 (UserExists)

```bash
peer chaincode query -C mychannel -n powercc -c '{"function":"IdentityContract:UserExists","Args":["did:example:123456789abcdef0"]}'
```

### 获取所有用户 (GetAllUsers)

```bash
peer chaincode query -C mychannel -n powercc -c '{"function":"IdentityContract:GetAllUsers","Args":[]}'
```

## 风险管理命令 (RiskContract)

### 更新用户风险评分 (UpdateRiskScore)

```bash
peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile ${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem -C mychannel -n powercc --peerAddresses localhost:7051 --tlsRootCertFiles ${PWD}/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt --peerAddresses localhost:9051 --tlsRootCertFiles ${PWD}/organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt -c '{"function":"RiskContract:UpdateRiskScore","Args":["did:example:123456789abcdef0", "60"]}'
```

### 获取用户风险评分 (GetRiskScore)

```bash
peer chaincode query -C mychannel -n powercc -c '{"function":"RiskContract:GetRiskScore","Args":["did:example:123456789abcdef0"]}'
```

### 检查用户登录资格 (CheckUserLoginEligibility)

```bash
peer chaincode query -C mychannel -n powercc -c '{"function":"RiskContract:CheckUserLoginEligibility","Args":["did:example:123456789abcdef0"]}'
```

返回值说明：
- 如果返回 `"用户可以登录"` - 用户风险评分在允许范围内
- 如果返回 `"禁止用户登录"` - 用户风险评分超过阈值

### 获取高风险用户 (GetHighRiskUsers)

```bash
peer chaincode query -C mychannel -n powercc -c '{"function":"RiskContract:GetHighRiskUsers","Args":[]}'
```

### 获取特定风险评分范围内的用户 (GetUsersByRiskScoreRange)

```bash
peer chaincode query -C mychannel -n powercc -c '{"function":"RiskContract:GetUsersByRiskScoreRange","Args":["40", "70"]}'
```

## 工具函数相关命令

以下命令展示了如何使用链码中的工具函数：

### 生成DID (GenerateDID)

这是链码内部使用的函数，通过以下参数生成：

```
did:example:<用户信息的SHA-256哈希前16位>
```

例如，用户信息 "张三", "110101199001011234", "13800138000", "京A12345" 将生成唯一的DID。

### 验证DID格式 (ValidateDID)

DID格式必须符合 `did:example:<至少16位标识符>` 的格式。

## 测试与部署

### 部署链码

使用项目根目录下的deploy.sh脚本部署链码：

```bash
./deploy.sh
```

### 关闭网络

当测试完成后，可以使用以下命令关闭网络：

```bash
cd $FABRIC_SAMPLES_PATH/test-network
./network.sh down
```