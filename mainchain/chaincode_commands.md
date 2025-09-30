# 身份认证链码命令指南

本文档提供了与身份认证链码交互的基本命令指南，包括如何部署、调用和查询链码。

## 环境准备

在执行任何命令前，请先确保已经加载了正确的环境变量：

```bash
cd ../mainchain_docker
source ./env.sh
cd ../mainchain
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
docker exec cli_mainchain peer chaincode invoke \
  -o orderer.mainchain.com:8050 \
  --tls \
  --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/mainchain.com/orderers/orderer.mainchain.com/msp/tlscacerts/tlsca.mainchain.com-cert.pem \
  -C mainchannel \
  -n mainchaincc \
  --peerAddresses peer0.org1.mainchain.com:8051 \
  --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.mainchain.com/peers/peer0.org1.mainchain.com/tls/ca.crt \
  -c '{"function":"InitLedger","Args":[]}' \
  --waitForEvent
```

### 2. 注册新用户

```bash
docker exec cli_mainchain peer chaincode invoke \
  -o orderer.mainchain.com:8050 \
  --tls \
  --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/mainchain.com/orderers/orderer.mainchain.com/msp/tlscacerts/tlsca.mainchain.com-cert.pem \
  -C mainchannel \
  -n mainchaincc \
  --peerAddresses peer0.org1.mainchain.com:8051 \
  --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.mainchain.com/peers/peer0.org1.mainchain.com/tls/ca.crt \
  -c '{"function":"RegisterUser","Args":["张三", "110101199001011234", "13800138000", "京A12345"]}' \
  --waitForEvent
```

### 3. 获取用户DID

```bash
docker exec cli_mainchain peer chaincode query \
  -C mainchannel \
  -n mainchaincc \
  -c "{\"function\":\"GetDIDByInfo\",\"Args\":[\"张三\", \"110101199001011234\", \"13800138000\", \"京A12345\"]}"
```

### 4. 查询用户信息

```bash
# 先获取DID
DID=$(docker exec cli_mainchain peer chaincode query \
  -C mainchannel \
  -n mainchaincc \
  -c "{\"function\":\"GetDIDByInfo\",\"Args\":[\"张三\", \"110101199001011234\", \"13800138000\", \"京A12345\"]}" 2>/dev/null)

# 使用DID查询用户信息
docker exec cli_mainchain peer chaincode query \
  -C mainchannel \
  -n mainchaincc \
  -c "{\"function\":\"GetUser\",\"Args\":[\"$DID\"]}"
```

### 5. 用户登录

```bash
# 使用DID和用户名进行登录
docker exec cli_mainchain peer chaincode invoke \
  -o orderer.mainchain.com:8050 \
  --tls \
  --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/mainchain.com/orderers/orderer.mainchain.com/msp/tlscacerts/tlsca.mainchain.com-cert.pem \
  -C mainchannel \
  -n mainchaincc \
  --peerAddresses peer0.org1.mainchain.com:8051 \
  --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.mainchain.com/peers/peer0.org1.mainchain.com/tls/ca.crt \
  -c "{\"function\":\"UserLogin\",\"Args\":[\"$DID\", \"张三\"]}" \
  --waitForEvent
```

### 6. 用户登出

```bash
# 使用DID进行登出
docker exec cli_mainchain peer chaincode invoke \
  -o orderer.mainchain.com:8050 \
  --tls \
  --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/mainchain.com/orderers/orderer.mainchain.com/msp/tlscacerts/tlsca.mainchain.com-cert.pem \
  -C mainchannel \
  -n mainchaincc \
  --peerAddresses peer0.org1.mainchain.com:8051 \
  --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.mainchain.com/peers/peer0.org1.mainchain.com/tls/ca.crt \
  -c "{\"function\":\"UserLogout\",\"Args\":[\"$DID\"]}" \
  --waitForEvent
```

### 7. 验证用户身份

```bash
docker exec cli_mainchain peer chaincode query \
  -C mainchannel \
  -n mainchaincc \
  -c "{\"function\":\"VerifyIdentity\",\"Args\":[\"$DID\", \"张三\"]}"
```

### 8. 更新用户风险评分

```bash
docker exec cli_mainchain peer chaincode invoke \
  -o orderer.mainchain.com:8050 \
  --tls \
  --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/mainchain.com/orderers/orderer.mainchain.com/msp/tlscacerts/tlsca.mainchain.com-cert.pem \
  -C mainchannel \
  -n mainchaincc \
  --peerAddresses peer0.org1.mainchain.com:8051 \
  --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.mainchain.com/peers/peer0.org1.mainchain.com/tls/ca.crt \
  -c "{\"function\":\"UpdateRiskScore\",\"Args\":[\"$DID\", \"30\"]}" \
  --waitForEvent
```

### 9. 获取所有用户

```bash
docker exec cli_mainchain peer chaincode query \
  -C mainchannel \
  -n mainchaincc \
  -c "{\"function\":\"GetAllUsers\",\"Args\":[]}"
```

## 使用mainchain_cli.sh简化命令

mainchain_docker目录下的mainchain_cli.sh脚本可以简化链码调用：

```bash
cd ../mainchain_docker

# 查询命令
./mainchain_cli.sh query GetDIDByInfo 张三 110101199001011234 13800138000 京A12345

# 调用命令
./mainchain_cli.sh invoke RegisterUser 李四 110101199001012345 13900139000 京B12345

# 用户登录
./mainchain_cli.sh invoke UserLogin $DID 张三

# 用户登出
./mainchain_cli.sh invoke UserLogout $DID
```

注意：如果使用mainchain_cli.sh脚本，确保脚本中的函数调用使用双引号和转义，如下所示：
```bash
docker exec cli_mainchain peer chaincode query -C $CHANNEL_NAME -n $CHAINCODE_NAME -c "{\"function\":\"$FUNC_NAME\",\"Args\":$ARGS}"
```

## 链间通信工作流程

1. **用户注册**：在主链上注册用户，生成DID
2. **用户登录**：
   - 用户在主链上登录，状态变为"online"
   - transmit服务自动将用户状态同步到侧链，侧链状态也更新为"online"
3. **风险行为报告**：
   - 侧链记录用户的风险行为，计算风险评分
   - transmit服务自动将侧链的风险评分同步到主链
4. **用户登出**：
   - 用户在主链上登出，状态变为"offline"
   - transmit服务自动将用户状态同步到侧链，侧链状态也更新为"offline"

## 常见错误及解决方法

1. **找不到证书文件**：确保证书路径正确，可以使用绝对路径
2. **链码调用失败**：检查函数名和参数是否正确，特别注意JSON格式的正确性和引号的转义
3. **网络连接问题**：确保网络已启动并且容器正在运行
4. **链码容器启动失败**：检查docker网络配置，特别是`CORE_VM_DOCKER_HOSTCONFIG_NETWORKMODE`环境变量是否正确设置
5. **链码超时错误**：可能需要增加链码启动超时时间，修改peer节点的core.yaml文件中的`startuptimeout`参数

## 调试技巧

查看链码容器日志：

```bash
# 获取链码容器ID
CHAINCODE_CONTAINER=$(docker ps -a | grep mainchaincc | head -n 1 | awk '{print $1}')

# 查看日志
docker logs $CHAINCODE_CONTAINER
```

查看peer节点日志：

```bash
# 查看peer0节点日志
docker logs peer0.org1.mainchain.com

# 查看peer0节点错误日志
docker logs peer0.org1.mainchain.com 2>&1 | grep -i error | tail -n 20
```

检查网络配置：

```bash
# 查看docker网络列表
docker network ls | grep mainchain

# 检查网络详情
docker network inspect mainchain_docker_mainchain_network
```

检查链码容器网络配置：

```bash
# 获取链码容器ID
CHAINCODE_CONTAINER=$(docker ps -a | grep mainchaincc | head -n 1 | awk '{print $1}')

# 检查容器网络配置
docker inspect $CHAINCODE_CONTAINER | grep -A 20 "NetworkSettings"
```
