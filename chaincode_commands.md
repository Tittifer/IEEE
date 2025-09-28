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
peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile ${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem -C mychannel -n powercc --peerAddresses localhost:7051 --tlsRootCertFiles ${PWD}/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt --peerAddresses localhost:9051 --tlsRootCertFiles ${PWD}/organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt -c '{"function":"InitLedger","Args":[]}'
```

## 用户身份管理命令 (IdentityContract)

### 注册新用户 (RegisterUser)

```bash
peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile ${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem -C mychannel -n powercc --peerAddresses localhost:7051 --tlsRootCertFiles ${PWD}/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt --peerAddresses localhost:9051 --tlsRootCertFiles ${PWD}/organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt -c '{"function":"RegisterUser","Args":["张三", "110101199001011234", "13800138000", "京A12345"]}'
```

### 获取用户信息 (GetUser)

```bash
peer chaincode query -C mychannel -n powercc -c '{"function":"GetUser","Args":["did:example:123456789abcdef0"]}'
```

### 根据用户信息获取用户 (GetUserByInfo)

```bash
peer chaincode query -C mychannel -n powercc -c '{"function":"GetUserByInfo","Args":["张三", "110101199001011234", "13800138000", "京A12345"]}'
```

### 验证用户身份 (VerifyIdentity)

```bash
peer chaincode query -C mychannel -n powercc -c '{"function":"VerifyIdentity","Args":["did:example:123456789abcdef0", "张三", "110101199001011234", "13800138000", "京A12345"]}'
```

### 检查用户是否存在 (UserExists)

```bash
peer chaincode query -C mychannel -n powercc -c '{"function":"UserExists","Args":["did:example:123456789abcdef0"]}'
```

### 获取所有用户 (GetAllUsers)

```bash
peer chaincode query -C mychannel -n powercc -c '{"function":"GetAllUsers","Args":[]}'
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