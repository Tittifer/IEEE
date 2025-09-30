# DID风险评估侧链链码命令指南

本文档提供了与DID风险评估侧链链码交互的基本命令指南，包括如何部署、调用和查询链码。

## 环境准备

在执行任何命令前，请先确保已经加载了正确的环境变量：

```bash
cd ../sidechain_docker
source ./env.sh
cd ../sidechain
```

## 链码部署命令

使用deploy.sh脚本部署链码：

```bash
./deploy.sh -n    # 启动网络
./deploy.sh -d    # 部署链码
```

## 链码测试命令

### 1. 初始化账本

```bash
docker exec cli_sidechain peer chaincode invoke \
  -o orderer.sidechain.com:7050 \
  --tls \
  --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/sidechain.com/orderers/orderer.sidechain.com/msp/tlscacerts/tlsca.sidechain.com-cert.pem \
  -C sidechannel \
  -n sidechaincc \
  --peerAddresses peer0.org1.sidechain.com:7051 \
  --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.sidechain.com/peers/peer0.org1.sidechain.com/tls/ca.crt \
  -c '{"function":"DIDContract:InitLedger","Args":[]}' \
  --waitForEvent
```

### 2. 创建DID记录

```bash
docker exec cli_sidechain peer chaincode invoke \
  -o orderer.sidechain.com:7050 \
  --tls \
  --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/sidechain.com/orderers/orderer.sidechain.com/msp/tlscacerts/tlsca.sidechain.com-cert.pem \
  -C sidechannel \
  -n sidechaincc \
  --peerAddresses peer0.org1.sidechain.com:7051 \
  --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.sidechain.com/peers/peer0.org1.sidechain.com/tls/ca.crt \
  -c '{"function":"DIDContract:CreateDIDRecord","Args":["did:example:1234567890abcdef"]}' \
  --waitForEvent
```

### 3. 获取DID记录

```bash
docker exec cli_sidechain peer chaincode query \
  -C sidechannel \
  -n sidechaincc \
  -c '{"function":"DIDContract:GetDIDRecord","Args":["did:example:1234567890abcdef"]}'
```

### 4. 更新用户状态（登录）

```bash
docker exec cli_sidechain peer chaincode invoke \
  -o orderer.sidechain.com:7050 \
  --tls \
  --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/sidechain.com/orderers/orderer.sidechain.com/msp/tlscacerts/tlsca.sidechain.com-cert.pem \
  -C sidechannel \
  -n sidechaincc \
  --peerAddresses peer0.org1.sidechain.com:7051 \
  --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.sidechain.com/peers/peer0.org1.sidechain.com/tls/ca.crt \
  -c '{"function":"SessionContract:UpdateUserStatus","Args":["did:example:1234567890abcdef", "online"]}' \
  --waitForEvent
```

### 5. 报告风险行为

```bash
docker exec cli_sidechain peer chaincode invoke \
  -o orderer.sidechain.com:7050 \
  --tls \
  --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/sidechain.com/orderers/orderer.sidechain.com/msp/tlscacerts/tlsca.sidechain.com-cert.pem \
  -C sidechannel \
  -n sidechaincc \
  --peerAddresses peer0.org1.sidechain.com:7051 \
  --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.sidechain.com/peers/peer0.org1.sidechain.com/tls/ca.crt \
  -c '{"function":"RiskContract:ReportRiskBehavior","Args":["did:example:1234567890abcdef", "A"]}' \
  --waitForEvent
```

### 6. 检查用户风险阈值

```bash
docker exec cli_sidechain peer chaincode query \
  -C sidechannel \
  -n sidechaincc \
  -c '{"function":"RiskContract:CheckRiskThreshold","Args":["did:example:1234567890abcdef"]}'
```

### 7. 更新用户状态（登出）

```bash
docker exec cli_sidechain peer chaincode invoke \
  -o orderer.sidechain.com:7050 \
  --tls \
  --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/sidechain.com/orderers/orderer.sidechain.com/msp/tlscacerts/tlsca.sidechain.com-cert.pem \
  -C sidechannel \
  -n sidechaincc \
  --peerAddresses peer0.org1.sidechain.com:7051 \
  --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.sidechain.com/peers/peer0.org1.sidechain.com/tls/ca.crt \
  -c '{"function":"SessionContract:UpdateUserStatus","Args":["did:example:1234567890abcdef", "offline"]}' \
  --waitForEvent
```

### 8. 获取所有DID记录

```bash
docker exec cli_sidechain peer chaincode query \
  -C sidechannel \
  -n sidechaincc \
  -c '{"function":"DIDContract:GetAllDIDRecords","Args":[]}'
```

### 9. 获取用户会话信息

```bash
docker exec cli_sidechain peer chaincode query \
  -C sidechannel \
  -n sidechaincc \
  -c '{"function":"SessionContract:GetUserSession","Args":["did:example:1234567890abcdef"]}'
```

## 风险评估工作流程

1. **创建DID记录**：首先创建用户的DID记录
2. **用户登录**：
   - 用户在主链登录后，transmit服务自动在侧链更新用户状态为"online"
   - 系统重置用户的严重程度值`SeverityLevel`为0
   - 系统重置用户的风险触发标记`HasTriggeredRisk`为false
3. **报告风险行为**：当用户触发风险行为时，调用`RiskContract:ReportRiskBehavior`函数
   - 系统会自动记录风险行为并增加严重程度
   - 系统会自动计算并更新风险评分
   - 在用户当前会话中首次触发风险行为时，系统会更新时间戳
   - 时间间隔是本次登录首次触发风险行为和上次登录首次触发风险行为的时间差
   - 系统设置用户的风险触发标记`HasTriggeredRisk`为true
4. **风险评分同步**：
   - transmit服务自动将侧链的风险评分同步到主链
   - 如果风险评分超过阈值，主链用户状态会被更新为"risky"
5. **用户登出**：
   - 用户在主链登出后，transmit服务自动在侧链更新用户状态为"offline"
   - 系统重置用户的严重程度值`SeverityLevel`为0
   - 系统重置用户的风险触发标记`HasTriggeredRisk`为false

## 常见错误及解决方法

1. **找不到证书文件**：确保证书路径正确，可以使用绝对路径
2. **链码调用失败**：检查函数名和参数是否正确，特别注意JSON格式的正确性和引号的转义
3. **网络连接问题**：确保网络已启动并且容器正在运行
4. **链码容器启动失败**：检查docker网络配置，特别是`CORE_VM_DOCKER_HOSTCONFIG_NETWORKMODE`环境变量是否正确设置
5. **链码超时错误**：可能需要增加链码启动超时时间，修改peer节点的core.yaml文件中的`startuptimeout`参数
6. **找不到函数错误**：确保在函数名前加上合约名称前缀，例如`DIDContract:CreateDIDRecord`
7. **用户未登录错误**：确保在报告风险行为前，用户已经通过`SessionContract:UpdateUserStatus`函数登录

## 调试技巧

查看链码容器日志：

```bash
# 获取链码容器ID
CHAINCODE_CONTAINER=$(docker ps -a | grep sidechaincc | head -n 1 | awk '{print $1}')

# 查看日志
docker logs $CHAINCODE_CONTAINER
```

查看peer节点日志：

```bash
# 查看peer0节点日志
docker logs peer0.org1.sidechain.com

# 查看peer0节点错误日志
docker logs peer0.org1.sidechain.com 2>&1 | grep -i error | tail -n 20
```

检查网络配置：

```bash
# 查看docker网络列表
docker network ls | grep sidechain

# 检查网络详情
docker network inspect sidechain_docker_sidechain_network
```

检查链码容器网络配置：

```bash
# 获取链码容器ID
CHAINCODE_CONTAINER=$(docker ps -a | grep sidechaincc | head -n 1 | awk '{print $1}')

# 检查容器网络配置
docker inspect $CHAINCODE_CONTAINER | grep -A 20 "NetworkSettings"
```