# Device Client 设备客户端

设备客户端是一个用于与区块链网络交互的工具，用于电网设备的注册、信息查询和风险管理。该客户端允许设备在区块链上注册身份，查询风险评分和响应策略，以及重置风险评分。

## 功能特点

1. **设备注册**：将新设备注册到区块链网络，生成唯一的分布式身份标识符(DID)
2. **设备信息查询**：查询设备的详细信息，包括风险评分和攻击画像
3. **风险响应策略查询**：获取基于设备当前风险评分的响应策略
4. **风险评分重置**：重置设备的风险评分和攻击画像
5. **DID生成**：根据设备信息生成唯一的DID

## 目录结构

```
device_client/
├── client/           # 客户端代码
│   ├── config.go     # 配置文件
│   └── device_client.go # 设备客户端核心代码
├── go.mod            # Go模块文件
├── main.go           # 主程序入口
└── README.md         # 说明文档
```

## 设备信息结构

设备客户端与区块链交互，管理以下设备信息：

- **DID**：设备的分布式身份标识符
- **名称**：设备名称
- **型号**：设备型号
- **供应商**：设备供应商
- **风险评分**：当前风险评分（0-1000）
- **攻击画像指数**：反映设备历史攻击行为的广度和深度
- **攻击画像**：设备已触发过的不重复的行为类别集合
- **状态**：设备当前状态（活跃、非活跃、风险）

## 命令行界面

设备客户端提供了交互式命令行界面，支持以下命令：

### 基本命令

- `help` - 显示帮助信息
- `exit` - 退出程序

### 设备管理命令

- `register <设备名称> <设备型号> <设备供应商> <设备ID>` - 注册新设备
- `info <DID>` - 获取设备信息
- `did <设备名称> <设备型号> <设备供应商> <设备ID>` - 根据设备信息获取DID
- `reset <DID>` - 重置设备风险评分
- `risk <DID>` - 获取设备风险响应策略

## 使用示例

### 1. 启动设备客户端

```bash
cd device_client
go run main.go
```

### 2. 注册新设备

```
> register 智能电表 XM100 国家电网 SN12345678
设备注册成功! DID: did:ieee:device:1234567890abcdef
```

### 3. 查询设备信息

```
> info did:ieee:device:1234567890abcdef
{
  "did": "did:ieee:device:1234567890abcdef",
  "name": "智能电表",
  "model": "XM100",
  "vendor": "国家电网",
  "riskScore": 0.0,
  "attackIndexI": 0.0,
  "attackProfile": [],
  "lastEventTime": "2025-10-18T12:00:00Z",
  "status": "active",
  "createdAt": "2025-10-18T12:00:00Z",
  "lastUpdatedAt": "2025-10-18T12:00:00Z"
}
```

### 4. 查询设备风险响应策略

```
> risk did:ieee:device:1234567890abcdef
{
  "riskLevel": "常规",
  "riskScore": 0.00,
  "strategy": "标准化信任与监控",
  "measures": [
    "维持默认的日志记录",
    "仅在会话建立或关键操作时通过智能合约验证其DID身份的有效性"
  ]
}
```

### 5. 重置设备风险评分

```
> reset did:ieee:device:1234567890abcdef
设备风险评分已重置为0
```

### 6. 根据设备信息获取DID

```
> did 智能电表 XM100 国家电网 SN12345678
设备DID: did:ieee:device:1234567890abcdef
```

## 风险等级与响应策略

设备客户端可以查询设备的风险等级和相应的响应策略：

1. **常规（0分）**：标准化信任与监控
   - 维持默认的日志记录
   - 仅在会话建立或关键操作时通过智能合约验证其DID身份的有效性

2. **关注（1-199分）**：增强监控，主动引诱
   - 针对该设备的源IP，自动启用全数据包捕获
   - 详细记录其在蜜点中的所有操作行为
   - 实施轻微的服务质量策略，对其扫描或连接行为进行速率限制
   - 在其当前互动的蜜点环境中，主动暴露更具吸引力的诱饵

3. **警戒（200-699分）**：主动欺骗与隔离引导
   - 将该设备的所有内部域名解析请求指向对应的伪造服务
   - 在网络层将其流量重定向至一个隔离的蜜网环境中
   - 在其所处的蜜网环境中，动态植入伪造的凭证文件、数据库连接字符串、API密钥等蜜点

4. **高危（700-1000分）**：硬性阻断
   - 立即强制中断该设备所有已建立的网络连接
   - 临时锁定设备账户
   - 对其交互过的蜜点进行快照存证
   - 进入人工介入和深度溯源

## 配置文件

设备客户端使用`config.json`配置文件，需要包含以下信息：

```json
{
  "mspid": "Org1MSP",
  "certPath": "../crypto-config/peerOrganizations/org1.example.com/users/User1@org1.example.com/msp/signcerts/User1@org1.example.com-cert.pem",
  "keyPath": "../crypto-config/peerOrganizations/org1.example.com/users/User1@org1.example.com/msp/keystore",
  "tlsCertPath": "../crypto-config/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt",
  "peerEndpoint": "localhost:7051",
  "gatewayPeer": "peer0.org1.example.com",
  "channelName": "mychannel",
  "chaincodeName": "ieee"
}
```

## 环境要求

- Go 1.18+
- Hyperledger Fabric 2.2+
- 正确配置的Fabric网络和已部署的链码

## 开发说明

1. 确保已安装Go和所需依赖
2. 配置正确的证书路径和网络连接信息
3. 运行`go mod tidy`安装依赖
4. 使用`go run main.go`启动客户端

## 错误处理

设备客户端会处理以下常见错误：

- 网络连接错误：检查网络配置和证书路径
- 设备已存在：尝试使用不同的设备信息
- 设备不存在：检查DID是否正确
- 链码调用错误：检查链码是否正确部署

## 安全注意事项

- 保护好设备的私钥和证书
- 不要在公共环境中暴露DID信息
- 定期检查设备的风险评分和响应策略
