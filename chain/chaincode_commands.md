# 身份认证链码命令指南

本文档提供了与身份认证链码交互的基本命令指南，包括如何部署、调用和查询链码。

## 环境准备

在执行任何命令前，请先确保已经加载了正确的环境变量：

```bash
cd ../chain_docker
source ./env.sh
cd ../chain
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
docker exec cli_chain peer chaincode invoke \
  -o orderer.chain.com:8050 \
  --tls \
  --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/chain.com/orderers/orderer.chain.com/msp/tlscacerts/tlsca.chain.com-cert.pem \
  -C mainchannel \
  -n chaincc \
  --peerAddresses peer0.org1.chain.com:8051 \
  --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.chain.com/peers/peer0.org1.chain.com/tls/ca.crt \
  -c '{"function":"InitLedger","Args":[]}' \
  --waitForEvent
```

### 2. 注册新设备

```bash
docker exec cli_chain peer chaincode invoke \
  -o orderer.chain.com:8050 \
  --tls \
  --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/chain.com/orderers/orderer.chain.com/msp/tlscacerts/tlsca.chain.com-cert.pem \
  -C mainchannel \
  -n chaincc \
  --peerAddresses peer0.org1.chain.com:8051 \
  --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.chain.com/peers/peer0.org1.chain.com/tls/ca.crt \
  -c '{"function":"RegisterDevice","Args":["智能电表", "XM100", "国家电网", "SN12345678"]}' \
  --waitForEvent
```

### 3. 获取设备DID

```bash
docker exec cli_chain peer chaincode query \
  -C mainchannel \
  -n chaincc \
  -c "{\"function\":\"GetDIDByInfo\",\"Args\":[\"智能电表\", \"XM100\", \"国家电网\", \"SN12345678\"]}"
```

### 4. 查询设备信息

```bash
# 先获取DID
DID=$(docker exec cli_chain peer chaincode query \
  -C mainchannel \
  -n chaincc \
  -c "{\"function\":\"GetDIDByInfo\",\"Args\":[\"智能电表\", \"XM100\", \"国家电网\", \"SN12345678\"]}" 2>/dev/null)

# 使用DID查询设备信息
docker exec cli_chain peer chaincode query \
  -C mainchannel \
  -n chaincc \
  -c "{\"function\":\"GetDevice\",\"Args\":[\"$DID\"]}"
```

### 5. 验证设备身份

```bash
docker exec cli_chain peer chaincode query \
  -C mainchannel \
  -n chaincc \
  -c "{\"function\":\"VerifyDeviceIdentity\",\"Args\":[\"$DID\", \"智能电表\", \"XM100\"]}"
```

### 6. 更新设备风险评分

```bash
# 准备攻击画像JSON数据
ATTACK_PROFILE='[\"Recon\", \"InitialAccess\"]'

docker exec cli_chain peer chaincode invoke \
  -o orderer.chain.com:8050 \
  --tls \
  --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/chain.com/orderers/orderer.chain.com/msp/tlscacerts/tlsca.chain.com-cert.pem \
  -C mainchannel \
  -n chaincc \
  --peerAddresses peer0.org1.chain.com:8051 \
  --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.chain.com/peers/peer0.org1.chain.com/tls/ca.crt \
  -c "{\"function\":\"UpdateDeviceRiskScore\",\"Args\":[\"$DID\", \"30.5\", \"1.2\", \"$ATTACK_PROFILE\"]}" \
  --waitForEvent
```

### 7. 重置设备风险评分

```bash
docker exec cli_chain peer chaincode invoke \
  -o orderer.chain.com:8050 \
  --tls \
  --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/chain.com/orderers/orderer.chain.com/msp/tlscacerts/tlsca.chain.com-cert.pem \
  -C mainchannel \
  -n chaincc \
  --peerAddresses peer0.org1.chain.com:8051 \
  --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.chain.com/peers/peer0.org1.chain.com/tls/ca.crt \
  -c "{\"function\":\"ResetDeviceRiskScore\",\"Args\":[\"$DID\"]}" \
  --waitForEvent
```

### 8. 获取所有设备信息

```bash
docker exec cli_chain peer chaincode query \
  -C mainchannel \
  -n chaincc \
  -c "{\"function\":\"GetAllDevices\",\"Args\":[]}"
```

### 9. 获取设备风险响应策略

```bash
docker exec cli_chain peer chaincode query \
  -C mainchannel \
  -n chaincc \
  -c "{\"function\":\"GetDeviceRiskResponse\",\"Args\":[\"$DID\"]}"
```

## 使用chain_cli.sh简化命令

chain_docker目录下的chain_cli.sh脚本可以简化链码调用：

```bash
cd ../chain_docker

# 查询命令
./chain_cli.sh query GetDIDByInfo 智能电表 XM100 国家电网 SN12345678

# 调用命令
./chain_cli.sh invoke RegisterDevice 智能插座 SP200 国家电网 SN87654321

# 获取设备信息
./chain_cli.sh query GetDevice $DID

# 获取所有设备
./chain_cli.sh query GetAllDevices

# 重置设备风险评分
./chain_cli.sh invoke ResetDeviceRiskScore $DID
```

注意：如果使用chain_cli.sh脚本，确保脚本中的函数调用使用双引号和转义，如下所示：
```bash
docker exec cli_chain peer chaincode query -C $CHANNEL_NAME -n $CHAINCODE_NAME -c "{\"function\":\"$FUNC_NAME\",\"Args\":$ARGS}"
```

## 链间通信工作流程

1. **设备注册**：在主链上注册设备，生成DID
2. **风险行为报告**：
   - 蜜点客户端检测到设备的风险行为，计算风险评分
   - 将风险评分、攻击画像指数和攻击画像更新到链上
3. **风险响应**：
   - 根据设备的风险评分，系统自动执行相应的响应策略
   - 高风险设备（700-1000分）：硬性阻断
   - 警戒风险设备（200-699分）：主动欺骗与隔离引导
   - 关注风险设备（1-199分）：增强监控，主动引诱
   - 常规风险设备（0分）：标准化信任与监控

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
CHAINCODE_CONTAINER=$(docker ps -a | grep chaincc | head -n 1 | awk '{print $1}')

# 查看日志
docker logs $CHAINCODE_CONTAINER
```

查看peer节点日志：

```bash
# 查看peer0节点日志
docker logs peer0.org1.chain.com

# 查看peer0节点错误日志
docker logs peer0.org1.chain.com 2>&1 | grep -i error | tail -n 20
```

检查网络配置：

```bash
# 查看docker网络列表
docker network ls | grep chain

# 检查网络详情
docker network inspect chain_docker_chain_network
```

检查链码容器网络配置：

```bash
# 获取链码容器ID
CHAINCODE_CONTAINER=$(docker ps -a | grep chaincc | head -n 1 | awk '{print $1}')

# 检查容器网络配置
docker inspect $CHAINCODE_CONTAINER | grep -A 20 "NetworkSettings"
```