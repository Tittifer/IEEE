# 电网设备身份认证与风险评估链码

这是一个基于Hyperledger Fabric的智能合约，用于电网场景下的设备身份认证和风险评估。该链码维护设备信息表，包括设备的DID、风险评分、攻击画像指数和攻击画像，并提供身份验证和风险管理功能。

## 项目结构

```
chain/
├── main.go                 # 主程序入口
├── go.mod                  # Go模块定义
├── models/                 # 数据模型
│   └── device.go           # 设备相关模型
├── contracts/              # 智能合约
│   ├── identity_contract.go  # 身份管理合约
│   └── risk_contract.go      # 风险评估合约
└── utils/                  # 工具函数
    └── identity_utils.go     # 身份相关工具函数
```

## 功能特点

- **设备身份管理**：注册、验证和管理设备身份
- **风险评分系统**：基于设备行为动态调整风险评分
- **攻击画像管理**：维护设备的攻击画像和攻击画像指数
- **设备状态管理**：管理设备的活跃状态
- **DID生成与验证**：生成和验证分布式身份标识符
- **风险响应策略**：根据风险评分提供不同的响应策略

## 数据结构

### 设备信息 (DeviceInfo)

```go
type DeviceInfo struct {
    DID           string    `json:"did"`           // 设备的分布式身份标识符
    Name          string    `json:"name"`          // 设备名称
    Model         string    `json:"model"`         // 设备型号
    Vendor        string    `json:"vendor"`        // 设备供应商
    RiskScore     float64   `json:"riskScore"`     // 设备历史风险分数 (S_{t-1})，范围 [0, S_{max}]
    AttackIndexI  float64   `json:"attackIndexI"`  // 攻击画像指数 (I)，范围 [0, ∞)
    AttackProfile []string  `json:"attackProfile"` // 攻击画像，存储设备已触发过的不重复的行为类别
    LastEventTime time.Time `json:"lastEventTime"` // 上次事件时间 (t_{last})
    Status        string    `json:"status"`        // 设备状态: active/inactive/risky
    CreatedAt     time.Time `json:"createdAt"`     // 创建时间
    LastUpdatedAt time.Time `json:"lastUpdatedAt"` // 最后更新时间
}
```

## 主要合约

### IdentityContract

身份管理合约，处理设备的注册、查询和身份验证。

- **InitLedger**: 初始化账本
- **RegisterDevice**: 注册新设备
- **GetDevice**: 获取设备信息
- **DeviceExists**: 检查设备是否存在
- **GetDIDByInfo**: 根据设备信息生成DID
- **VerifyDeviceIdentity**: 验证设备身份
- **UpdateDeviceRiskScore**: 更新设备风险评分
- **ResetDeviceRiskScore**: 重置设备风险评分
- **GetAllDevices**: 获取所有设备

### RiskContract

风险评估合约，处理设备的风险评分管理。

- **InitRiskLedger**: 初始化风险管理账本
- **UpdateRiskScore**: 更新设备风险评分
- **GetRiskScore**: 获取设备风险评分
- **GetAttackProfile**: 获取设备攻击画像
- **CheckDeviceConnectionEligibility**: 检查设备是否有资格连接
- **GetHighRiskDevices**: 获取高风险设备
- **GetDevicesByRiskScoreRange**: 获取特定风险评分范围内的设备
- **GetDeviceRiskResponse**: 获取设备风险响应策略

## 风险评分

风险评分范围从0到1000，根据风险等级划分为以下几个区间：
- **常规（0分）**：标准化信任与监控
- **关注（1-199分）**：增强监控，主动引诱
- **警戒（200-699分）**：主动欺骗与隔离引导
- **高危（700-1000分）**：硬性阻断

## 设备状态

设备状态包括以下几种：
- **active**: 设备活跃状态
- **inactive**: 设备非活跃状态
- **risky**: 设备风险状态（风险评分超过阈值）

## 风险评估算法

系统实现了基于历史行为和时间衰减的风险评分算法：

1. **更新攻击画像指数 (I)**：
   - 如果当前行为类别不在设备的攻击画像集合中（意图升级）：
     ```
     ΔI = W
     ```
   - 如果已存在（持续试探）：
     ```
     ΔI = 0
     ```
   - 完成I的累加：
     ```
     I_new = I_old + ΔI
     ```

2. **计算实时风险分 (S_t)**：
   - 对历史分数进行降温：
     ```
     S'_{t-1} = max(0, S_{t-1} - (δ * Δt) / (1 + α * S_{t-1}))
     ```
   - 计算最终得分：
     ```
     S_t = min(S_max, S_base * (1+I) + S'_{t-1})
     ```
   
3. **后台状态维护**：
   - I慢速衰减：
     ```
     I_new = I_old * e^(-λ*Δt)
     ```

## 使用示例

### 1. 注册新设备

```
peer chaincode invoke -C mainchannel -n chaincc -c '{"function":"RegisterDevice","Args":["智能电表", "XM100", "国家电网", "SN12345678"]}'
```

### 2. 验证设备身份

```
peer chaincode query -C mainchannel -n chaincc -c '{"function":"VerifyDeviceIdentity","Args":["did:ieee:device:1234567890abcdef", "智能电表", "XM100"]}'
```

### 3. 更新设备风险评分

```
peer chaincode invoke -C mainchannel -n chaincc -c '{"function":"UpdateDeviceRiskScore","Args":["did:ieee:device:1234567890abcdef", "30.5", "1.2", "[\"Recon\",\"InitialAccess\"]"]}'
```

### 4. 获取设备风险响应策略

```
peer chaincode query -C mainchannel -n chaincc -c '{"function":"GetDeviceRiskResponse","Args":["did:ieee:device:1234567890abcdef"]}'
```

### 5. 获取高风险设备

```
peer chaincode query -C mainchannel -n chaincc -c '{"function":"GetHighRiskDevices","Args":[]}'
```

### 6. 重置设备风险评分

```
peer chaincode invoke -C mainchannel -n chaincc -c '{"function":"ResetDeviceRiskScore","Args":["did:ieee:device:1234567890abcdef"]}'
```

### 7. 获取所有设备

```
peer chaincode query -C mainchannel -n chaincc -c '{"function":"GetAllDevices","Args":[]}'
```

## 部署说明

1. 安装依赖：
   ```
   go mod tidy
   ```

2. 打包链码：
   ```
   peer lifecycle chaincode package chaincc.tar.gz --path /path/to/chaincode --lang golang --label chaincc_1.0
   ```

3. 安装链码：
   ```
   peer lifecycle chaincode install chaincc.tar.gz
   ```

4. 批准链码定义：
   ```
   peer lifecycle chaincode approveformyorg -o orderer.chain.com:8050 --channelID mainchannel --name chaincc --version 1.0 --package-id $PACKAGE_ID --sequence 1 --tls --cafile /path/to/orderer/tls/ca.crt
   ```

5. 提交链码定义：
   ```
   peer lifecycle chaincode commit -o orderer.chain.com:8050 --channelID mainchannel --name chaincc --version 1.0 --sequence 1 --tls --cafile /path/to/orderer/tls/ca.crt
   ```