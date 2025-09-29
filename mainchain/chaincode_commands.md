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
docker exec peer0.org1.mainchain.com peer chaincode invoke \
  -o localhost:8051 \
  --tls \
  --cafile /etc/hyperledger/fabric/msp/tlscacerts/tlsca.mainchain.com-cert.pem \
  -C mainchannel \
  -n mainchaincc \
  --peerAddresses peer0.org1.mainchain.com:8051 \
  --tlsRootCertFiles /etc/hyperledger/fabric/tls/ca.crt \
  -c '{"function":"InitLedger","Args":[]}' \
  --waitForEvent
```

### 2. 注册新用户

```bash
docker exec cli_mainchain peer chaincode invoke \
  -o localhost:8050 \
  --tls \
  --cafile /etc/hyperledger/fabric/msp/tlscacerts/tlsca.org1.mainchain.com-cert.pem\
  -C mainchannel \
  -n mainchaincc \
  --peerAddresses peer0.org1.mainchain.com:8051 \
  --tlsRootCertFiles /etc/hyperledger/fabric/tls/ca.crt \
  -c '{"function":"RegisterUser","Args":["张三", "110101199001011234", "13800138000", "京A12345"]}' \
  --waitForEvent
```

### 3. 获取用户DID

```bash
docker exec cli_mainchain peer chaincode query \
  -C mainchannel \
  -n mainchaincc \
  -c '{"function":"GetDIDByInfo","Args":["张三", "110101199001011234", "13800138000", "京A12345"]}'
```

### 4. 查询用户信息

```bash
# 先获取DID
DID=$(docker exec cli_mainchain peer chaincode query \
  -C mainchannel \
  -n mainchaincc \
  -c '{"function":"GetDIDByInfo","Args":["张三", "110101199001011234", "13800138000", "京A12345"]}' 2>/dev/null)

# 使用DID查询用户信息
docker exec cli_mainchain peer chaincode query \
  -C mainchannel \
  -n mainchaincc \
  -c "{\"function\":\"GetUser\",\"Args\":[\"$DID\"]}"
```

### 5. 用户登录

```bash
# 使用DID和用户名进行登录
docker exec cli_mainchain peer chaincode query \
  -C mainchannel \
  -n mainchaincc \
  -c "{\"function\":\"UserLogin\",\"Args\":[\"$DID\", \"张三\"]}"
```

### 6. 验证用户身份

```bash
docker exec cli_mainchain peer chaincode query \
  -C mainchannel \
  -n mainchaincc \
  -c "{\"function\":\"VerifyIdentity\",\"Args\":[\"$DID\", \"张三\"]}"
```

### 7. 获取所有用户

```bash
docker exec cli_mainchain peer chaincode query \
  -C mainchannel \
  -n mainchaincc \
  -c '{"function":"GetAllUsers","Args":[]}'
```

## 使用mainchain_cli.sh简化命令

mainchain_docker目录下的mainchain_cli.sh脚本可以简化链码调用：

```bash
cd ../mainchain_docker

# 查询命令
./mainchain_cli.sh query GetDIDByInfo 张三 110101199001011234 13800138000 京A12345

# 调用命令
./mainchain_cli.sh invoke RegisterUser 李四 110101199001012345 13900139000 京B12345
```

## 常见错误及解决方法

1. **找不到证书文件**：确保证书路径正确，可以使用绝对路径
2. **链码调用失败**：检查函数名和参数是否正确
3. **网络连接问题**：确保网络已启动并且容器正在运行

## 调试技巧

查看链码容器日志：

```bash
# 获取链码容器ID
CHAINCODE_CONTAINER=$(docker ps -a | grep mainchaincc | head -n 1 | awk '{print $1}')

# 查看日志
docker logs $CHAINCODE_CONTAINER
```

查看peer日志：

```bash
docker logs peer0.org1.mainchain.com
```


