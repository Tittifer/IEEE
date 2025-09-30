# 电网用户身份与风险评估链间数据传输服务

这是一个用于连接和转发Hyperledger Fabric主链和侧链之间数据的服务。主链负责用户身份管理，侧链负责风险评估，本服务自动在两条链之间同步和转发数据。

## 项目结构

```
transmit/
├── main.go                 # 主程序入口
├── go.mod                  # Go模块定义
├── config/                 # 配置
│   └── config.go           # 配置定义
├── connection/             # 区块链连接
│   ├── mainchain.go        # 主链连接
│   └── sidechain.go        # 侧链连接
├── handlers/               # 链上操作处理
│   ├── mainchain_handler.go # 主链处理器
│   └── sidechain_handler.go # 侧链处理器
└── models/                 # 数据模型
    └── user_data.go        # 用户数据模型
```

## 功能特点

- **主链交互**：与主链进行交互，获取和更新用户身份信息
- **侧链交互**：与侧链进行交互，获取和更新风险评估信息
- **数据同步**：定期在主链和侧链之间同步数据
- **自动处理**：自动处理链间数据转发，无需人工干预
- **用户状态同步**：同步用户登录/登出状态
- **风险评分同步**：同步风险评分结果

## 主要功能

### 1. 用户DID同步

- 从主链获取所有用户信息
- 检查每个用户在侧链是否有对应的DID记录
- 如果没有，则在侧链创建DID记录

### 2. 用户登录/登出状态同步

- 当用户在主链登录时，自动更新侧链用户状态为"online"
- 当用户在主链登出时，自动更新侧链用户状态为"offline"
- 当侧链用户状态变化时，同步更新主链用户状态

### 3. 风险评分同步

- 从侧链获取用户风险评分
- 将侧链风险评分同步到主链
- 如果风险评分超过阈值，更新主链用户状态为"risky"

### 4. 数据一致性检查

- 定期检查主链和侧链数据的一致性
- 确保两条链之间的数据同步和一致

## 工作流程

1. 服务启动，连接主链和侧链
2. 定期（默认30秒）执行数据同步操作：
   - **用户信息同步**：
     - 从主链获取所有用户
     - 检查每个用户在侧链是否有DID记录，如无则创建
   - **风险评分同步**：
     - 从侧链获取用户风险评分
     - 将风险评分同步到主链
     - 如果风险评分超过阈值，更新主链用户状态为"risky"
   - **用户状态同步**：
     - 如果主链用户状态为"online"，确保侧链用户状态也为"online"
     - 如果主链用户状态为"offline"，确保侧链用户状态也为"offline"
     - 如果侧链用户状态为"online"，确保主链用户状态也为"online"
     - 如果侧链用户状态为"offline"，确保主链用户状态也为"offline"
   - **数据一致性检查**：
     - 从侧链获取所有DID记录
     - 检查每个DID在主链是否有对应用户

## 链间通信场景

### 场景1：用户登录流程

1. 用户在主链上登录，状态变为"online"
2. transmit服务检测到主链用户状态变化
3. transmit服务自动将用户状态同步到侧链，侧链状态也更新为"online"
4. 侧链重置用户的严重程度值和风险触发标记

### 场景2：风险行为报告流程

1. 侧链记录用户的风险行为，计算风险评分
2. transmit服务检测到侧链风险评分变化
3. transmit服务自动将侧链的风险评分同步到主链
4. 如果风险评分超过阈值，主链用户状态会被更新为"risky"

### 场景3：用户登出流程

1. 用户在主链上登出，状态变为"offline"
2. transmit服务检测到主链用户状态变化
3. transmit服务自动将用户状态同步到侧链，侧链状态也更新为"offline"
4. 侧链重置用户的严重程度值和风险触发标记

## 部署说明

1. 安装依赖：
   ```bash
   go mod tidy
   ```

2. 编译：
   ```bash
   go build -o transmit
   ```

3. 运行：
   ```bash
   ./transmit
   ```

## 配置说明

配置文件位于`config/config.go`，包含主链和侧链的连接信息。您可以根据实际环境修改配置：

```go
// 获取默认配置
func GetDefaultConfig() *Config {
    return &Config{
        MainchainConfig: NetworkConfig{
            ChannelName:   "mainchannel",
            ChaincodeName: "mainchaincc",
            MspID:         "Org1MSP",
            CertPath:      "/path/to/cert.pem",
            KeyPath:       "/path/to/key.pem",
            TlsCertPath:   "/path/to/tls/ca.crt",
            PeerEndpoint:  "localhost:8051",
            GatewayPeer:   "peer0.org1.mainchain.com:8051",
        },
        SidechainConfig: NetworkConfig{
            ChannelName:   "sidechannel",
            ChaincodeName: "sidechaincc",
            MspID:         "Org1MSP",
            CertPath:      "/path/to/cert.pem",
            KeyPath:       "/path/to/key.pem",
            TlsCertPath:   "/path/to/tls/ca.crt",
            PeerEndpoint:  "localhost:7051",
            GatewayPeer:   "peer0.org1.sidechain.com:7051",
        },
        SyncInterval: 30, // 默认30秒同步一次
    }
}
```

## 注意事项

1. 确保主链和侧链已经正确部署并启动
2. 确保证书和私钥路径正确
3. 确保网络连接正常
4. 服务启动后会自动连接主链和侧链并开始数据同步
5. 如果需要更改同步间隔，可以修改`SyncInterval`配置