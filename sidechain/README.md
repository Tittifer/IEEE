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
- **链间数据同步**：与主链进行数据同步，包括用户登录/登出状态

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
   - Δt：时间间隔（当前登录首次触发风险行为与上次登录首次触发风险行为的时间差）
   - α：影响高分时降温速度参数

## 用户会话和风险行为流程

1. **用户登录**：
   - 主链用户登录后，通过transmit服务同步到侧链
   - 侧链调用`SessionContract:UpdateUserStatus`更新用户状态为"online"
   - 重置用户的严重程度值`SeverityLevel`为0
   - 重置用户的风险触发标记`HasTriggeredRisk`为false

2. **风险行为报告**：
   - 当用户触发风险行为时，调用`RiskContract:ReportRiskBehavior`
   - 系统增加用户的严重程度值`SeverityLevel`
   - 设置用户的风险触发标记`HasTriggeredRisk`为true
   - 如果是当前会话中首次触发风险行为，更新时间戳`Timestamp`
   - 计算风险评分并更新到账本

3. **用户登出**：
   - 主链用户登出后，通过transmit服务同步到侧链
   - 侧链调用`SessionContract:UpdateUserStatus`更新用户状态为"offline"
   - 重置用户的严重程度值`SeverityLevel`为0
   - 重置用户的风险触发标记`HasTriggeredRisk`为false

## 使用示例

### 1. 创建DID记录

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

### 2. 更新用户状态（登录）

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

### 3. 报告风险行为

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

### 4. 检查用户风险阈值

```bash
docker exec cli_sidechain peer chaincode query \
  -C sidechannel \
  -n sidechaincc \
  -c '{"function":"RiskContract:CheckRiskThreshold","Args":["did:example:1234567890abcdef"]}'
```

### 5. 更新用户状态（登出）

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

## 链间通信

侧链与主链之间通过transmit模块进行数据同步，实现以下功能：
- 主链用户登录时，同步更新侧链用户状态为登录状态
- 主链用户登出时，同步更新侧链用户状态为登出状态
- 侧链风险评分变化时，同步更新主链用户的风险评分

## 部署说明

1. 安装依赖：
   ```bash
   go mod tidy
   ```

2. 使用deploy.sh脚本部署链码：
   ```bash
   ./deploy.sh -n    # 启动网络
   ./deploy.sh -d    # 部署链码
   ./deploy.sh -t    # 测试链码
   ```