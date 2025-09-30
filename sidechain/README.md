# 电网用户DID风险评估侧链链码

这是一个基于Hyperledger Fabric的智能合约，用于电网场景下的用户DID风险评估侧链。该链码维护用户DID与风险评分的映射表，并提供风险评估和管理功能。

## 项目结构

```
sidechain/
├── main.go                 # 主程序入口
├── go.mod                  # Go模块定义
├── models/                 # 数据模型
│   ├── did_record.go       # DID记录模型
│   ├── risk_behavior.go    # 风险行为模型
│   └── user_session.go     # 用户会话模型
├── contracts/              # 智能合约
│   ├── did_contract.go     # DID管理合约
│   ├── session_contract.go # 会话管理合约
│   └── risk_contract.go    # 风险评估合约
└── utils/                  # 工具函数
    ├── did_utils.go        # DID相关工具函数
    └── risk_utils.go       # 风险评估工具函数
```

## 功能特点

- **DID管理**：创建和管理用户DID记录
- **风险评分系统**：基于风险行为动态调整风险评分
- **用户会话管理**：跟踪用户登录状态和风险行为
- **风险行为识别**：识别和记录不同类型的风险行为
- **风险评分衰减**：基于时间的风险评分自动衰减机制

## 数据结构

### DID风险记录 (DIDRiskRecord)

```go
type DIDRiskRecord struct {
    DID       string `json:"did"`       // 用户的分布式身份标识符
    RiskScore int    `json:"riskScore"` // 风险评分
    Timestamp int64  `json:"timestamp"` // 时间戳
}
```

### 用户会话 (UserSession)

```go
type UserSession struct {
    DID              string           `json:"did"`              // 用户的分布式身份标识符
    Status           UserSessionStatus `json:"status"`          // 会话状态：online/offline
    SeverityLevel    int              `json:"severityLevel"`    // 当前会话中的风险行为严重程度累计值s
    Timestamp        int64            `json:"timestamp"`        // 最后更新时间戳
    HasTriggeredRisk bool             `json:"hasTriggeredRisk"` // 当前会话中是否已触发过风险行为
}
```

### 风险行为 (RiskBehavior)

```go
type RiskBehavior struct {
    Type            RiskBehaviorType `json:"type"`            // 风险行为类型
    SeverityLevel   int              `json:"severityLevel"`   // 风险行为严重程度s
    OccurrenceTime  int64            `json:"occurrenceTime"`  // 风险行为发生时间
}
```

## 主要合约

### DIDContract

DID管理合约，处理DID记录的创建和查询。

- **InitLedger**: 初始化账本
- **CreateDIDRecord**: 创建新的DID记录
- **GetDIDRecord**: 获取DID记录
- **DIDExists**: 检查DID是否存在
- **GetAllDIDRecords**: 获取所有DID记录

### SessionContract

会话管理合约，处理用户登录状态和风险行为记录。

- **UpdateUserStatus**: 更新用户状态（登录/登出）
- **GetUserSession**: 获取用户会话信息
- **RecordRiskBehavior**: 记录风险行为

### RiskContract

风险评估合约，处理风险评分的计算和更新。

- **UpdateRiskScore**: 更新用户风险评分
- **EvaluateRiskScore**: 评估用户当前风险评分
- **ReportRiskBehavior**: 报告风险行为并更新评分
- **CheckRiskThreshold**: 检查用户风险评分是否超过阈值

## 风险评估算法

侧链实现了两个关键的风险评估算法：

1. **分数更新公式**：
   ```
   S_t = min(S_max, S_{t-1}^{'} + K*s)
   ```
   其中：
   - S_t：本次计算分数
   - S_{t-1}^{'}：衰减后旧得分
   - K：用于把严重程度换算为得分的尺度系数
   - s：用户恶意行为严重程度

2. **降温曲线公式**：
   ```
   S_{t-1}^{'} = max(0, S_{t-1} - (δ*Δt)/(1+α*S_{t-1}))
   ```
   其中：
   - S_{t-1}：上次计算得分
   - δ：影响低分时降温速度参数
   - Δt：时间间隔
   - α：影响高分时降温速度参数

## 使用示例

### 1. 创建DID记录

```
peer chaincode invoke -C sidechannel -n sidechaincc -c '{"function":"CreateDIDRecord","Args":["did:example:1234567890abcdef"]}'
```

### 2. 更新用户状态（登录）

```
peer chaincode invoke -C sidechannel -n sidechaincc -c '{"function":"UpdateUserStatus","Args":["did:example:1234567890abcdef", "online"]}'
```

### 3. 报告风险行为

```
peer chaincode invoke -C sidechannel -n sidechaincc -c '{"function":"ReportRiskBehavior","Args":["did:example:1234567890abcdef", "A"]}'
```

### 4. 评估用户风险评分

```
peer chaincode query -C sidechannel -n sidechaincc -c '{"function":"EvaluateRiskScore","Args":["did:example:1234567890abcdef"]}'
```

### 5. 更新用户状态（登出）

```
peer chaincode invoke -C sidechannel -n sidechaincc -c '{"function":"UpdateUserStatus","Args":["did:example:1234567890abcdef", "offline"]}'
```

## 部署说明

1. 安装依赖：
   ```
   go mod tidy
   ```

2. 打包链码：
   ```
   peer lifecycle chaincode package sidechaincc.tar.gz --path /path/to/chaincode --lang golang --label sidechaincc_1.0
   ```

3. 安装链码：
   ```
   peer lifecycle chaincode install sidechaincc.tar.gz
   ```

4. 批准链码定义：
   ```
   peer lifecycle chaincode approveformyorg -o orderer.sidechain.com:8050 --channelID sidechannel --name sidechaincc --version 1.0 --package-id $PACKAGE_ID --sequence 1 --tls --cafile /path/to/orderer/tls/ca.crt
   ```

5. 提交链码定义：
   ```
   peer lifecycle chaincode commit -o orderer.sidechain.com:8050 --channelID sidechannel --name sidechaincc --version 1.0 --sequence 1 --tls --cafile /path/to/orderer/tls/ca.crt
   ```
