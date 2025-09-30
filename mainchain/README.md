# 电网用户身份认证与风险评估链码

这是一个基于Hyperledger Fabric的智能合约，用于电网场景下的用户身份认证和风险评估。该链码维护用户信息表，包括用户的DID和风险评分，并提供身份验证和风险管理功能。

## 项目结构

```
mainchain/
├── main.go                 # 主程序入口
├── go.mod                  # Go模块定义
├── models/                 # 数据模型
│   └── user.go             # 用户相关模型
├── contracts/              # 智能合约
│   ├── identity_contract.go  # 身份管理合约
│   └── risk_contract.go      # 风险评估合约
└── utils/                  # 工具函数
    └── identity_utils.go     # 身份相关工具函数
```

## 功能特点

- **用户身份管理**：注册、验证和管理用户身份
- **风险评分系统**：基于用户行为动态调整风险评分
- **用户状态管理**：管理用户的活跃状态、在线/离线状态
- **DID生成与验证**：生成和验证分布式身份标识符
- **链间数据同步**：与侧链进行数据同步，包括风险评分和用户状态

## 数据结构

### 用户信息 (UserInfo)

```go
type UserInfo struct {
    DID           string `json:"did"`           // 用户的分布式身份标识符
    Name          string `json:"name"`          // 用户姓名
    RiskScore     int    `json:"riskScore"`     // 用户风险评分 (0-100)
    Status        string `json:"status"`        // 用户状态: active/inactive/risky/online/offline
    CreatedAt     string `json:"createdAt"`     // 创建时间
    LastUpdatedAt string `json:"lastUpdatedAt"` // 最后更新时间
}
```

## 主要合约

### IdentityContract

身份管理合约，处理用户的注册、查询和身份验证。

- **InitLedger**: 初始化账本
- **RegisterUser**: 注册新用户
- **GetUser**: 获取用户信息
- **UserExists**: 检查用户是否存在
- **GetDIDByInfo**: 根据用户信息生成DID
- **VerifyIdentity**: 验证用户身份
- **UserLogin**: 用户登录
- **UserLogout**: 用户登出
- **UpdateRiskScore**: 更新用户风险评分
- **GetAllUsers**: 获取所有用户

### RiskContract

风险评估合约，处理用户的风险评分管理。

- **InitRiskLedger**: 初始化风险管理账本
- **UpdateRiskScore**: 更新用户风险评分
- **GetRiskScore**: 获取用户风险评分
- **CheckUserLoginEligibility**: 检查用户是否有资格登录
- **GetHighRiskUsers**: 获取高风险用户
- **GetUsersByRiskScoreRange**: 获取特定风险评分范围内的用户

## 风险评分

风险评分范围从0到100，风险评分阈值为50，超过此值将禁止用户登录。

## 用户状态

用户状态包括以下几种：
- **active**: 用户活跃状态
- **inactive**: 用户非活跃状态
- **risky**: 用户风险状态（风险评分超过阈值）
- **online**: 用户在线状态
- **offline**: 用户离线状态

## 使用示例

### 1. 注册新用户

```
peer chaincode invoke -C mainchannel -n powercc -c '{"function":"RegisterUser","Args":["张三", "110101199001011234", "13800138000", "京A12345"]}'
```

### 2. 验证用户身份

```
peer chaincode query -C mainchannel -n powercc -c '{"function":"VerifyIdentity","Args":["did:example:1234567890abcdef", "张三"]}'
```

### 3. 用户登录

```
peer chaincode invoke -C mainchannel -n powercc -c '{"function":"UserLogin","Args":["did:example:1234567890abcdef", "张三"]}'
```

### 4. 用户登出

```
peer chaincode invoke -C mainchannel -n powercc -c '{"function":"UserLogout","Args":["did:example:1234567890abcdef"]}'
```

### 5. 获取高风险用户

```
peer chaincode query -C mainchannel -n powercc -c '{"function":"GetHighRiskUsers","Args":[]}'
```

### 6. 更新用户风险评分

```
peer chaincode invoke -C mainchannel -n powercc -c '{"function":"UpdateRiskScore","Args":["did:example:1234567890abcdef", "30"]}'
```

## 链间通信

主链与侧链之间通过transmit模块进行数据同步，实现以下功能：
- 用户登录时，同步更新侧链用户状态为登录状态
- 用户登出时，同步更新侧链用户状态为登出状态
- 侧链风险评分变化时，同步更新主链用户的风险评分

## 部署说明

1. 安装依赖：
   ```
   go mod tidy
   ```

2. 打包链码：
   ```
   peer lifecycle chaincode package powercc.tar.gz --path /path/to/chaincode --lang golang --label powercc_1.0
   ```

3. 安装链码：
   ```
   peer lifecycle chaincode install powercc.tar.gz
   ```

4. 批准链码定义：
   ```
   peer lifecycle chaincode approveformyorg -o orderer.mainchain.com:8050 --channelID mainchannel --name powercc --version 1.0 --package-id $PACKAGE_ID --sequence 1 --tls --cafile /path/to/orderer/tls/ca.crt
   ```

5. 提交链码定义：
   ```
   peer lifecycle chaincode commit -o orderer.mainchain.com:8050 --channelID mainchannel --name powercc --version 1.0 --sequence 1 --tls --cafile /path/to/orderer/tls/ca.crt
   ```