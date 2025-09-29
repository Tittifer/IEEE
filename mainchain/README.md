# 电网用户身份认证与风险评估链码

这是一个基于Hyperledger Fabric的智能合约，用于电网场景下的用户身份认证和风险评估。该链码维护用户信息表，包括用户的DID和风险评分，并通过蜜点记录恶意攻击行为。

## 项目结构

```
IEEE/
├── main.go                 # 主程序入口
├── go.mod                  # Go模块定义
├── models/                 # 数据模型
│   └── user.go             # 用户和攻击记录相关模型
└── contracts/              # 智能合约
    ├── powergrid_contract.go  # 主合约
    ├── user_contract.go       # 用户管理合约
    └── risk_contract.go       # 风险评估合约
```

## 功能特点

- **用户身份管理**：注册、验证和管理用户身份
- **风险评分系统**：基于用户行为动态调整风险评分
- **攻击行为记录**：记录用户在蜜点上的攻击行为
- **访问控制**：根据风险评分自动调整用户访问级别
- **用户状态管理**：管理用户的活跃、暂停和阻止状态

## 数据结构

### 用户信息 (UserInfo)

```go
type UserInfo struct {
    DID           string    // 用户的分布式身份标识符
    Name          string    // 用户姓名
    IDNumber      string    // 身份证号
    PublicKey     string    // 用户公钥
    RiskScore     int       // 用户风险评分 (0-100)
    AccessLevel   int       // 用户访问级别 (1-5)
    Status        string    // 用户状态: active, suspended, blocked
    AttackHistory []Attack  // 攻击历史记录
    CreatedAt     time.Time // 创建时间
    LastUpdatedAt time.Time // 最后更新时间
}
```

### 攻击记录 (Attack)

```go
type Attack struct {
    Timestamp   time.Time // 攻击时间
    HoneypotID  string    // 蜜点ID
    AttackType  string    // 攻击类型
    Description string    // 攻击描述
    Severity    int       // 攻击严重程度 (1-10)
}
```

## 主要合约

### PowerGridContract

主合约，整合了用户管理和风险评估功能。

- **InitLedger**: 初始化账本

### UserContract

用户管理合约，处理用户的注册、查询和状态管理。

- **RegisterUser**: 注册新用户
- **GetUser**: 获取用户信息
- **VerifyUser**: 验证用户身份
- **ChangeUserStatus**: 更改用户状态
- **UserExists**: 检查用户是否存在
- **GetAllUsers**: 获取所有用户

### RiskContract

风险评估合约，处理用户的风险评分和攻击记录。

- **RecordAttack**: 记录用户攻击行为
- **UpdateRiskScore**: 手动更新用户风险评分
- **GetUsersByRiskScore**: 获取特定风险评分范围内的用户
- **GetHighRiskUsers**: 获取高风险用户（风险评分>=60）

## 风险评分与访问级别

风险评分范围从0到100，对应的访问级别和状态如下：

| 风险评分 | 访问级别 | 状态      | 说明                 |
|---------|---------|-----------|---------------------|
| 0-19    | 1       | active    | 最高权限，完全访问     |
| 20-39   | 2       | active    | 高权限，大部分功能可用 |
| 40-59   | 3       | active    | 中等权限，部分功能受限 |
| 60-79   | 4       | suspended | 低权限，大部分功能受限 |
| 80-100  | 5       | blocked   | 最低权限，几乎无法访问 |

## 使用示例

### 1. 注册新用户

```
peer chaincode invoke -C mychannel -n powercc -c '{"function":"RegisterUser","Args":["did:example:123", "张三", "110101199001011234", "公钥内容..."]}'
```

### 2. 验证用户身份

```
peer chaincode query -C mychannel -n powercc -c '{"function":"VerifyUser","Args":["did:example:123"]}'
```

### 3. 记录攻击行为

```
peer chaincode invoke -C mychannel -n powercc -c '{"function":"RecordAttack","Args":["did:example:123", "honeypot001", "SQL注入", "尝试通过SQL注入获取数据库访问权限", "8"]}'
```

### 4. 获取高风险用户

```
peer chaincode query -C mychannel -n powercc -c '{"function":"GetHighRiskUsers","Args":[]}'
```

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
   peer lifecycle chaincode approveformyorg -o localhost:7050 --channelID mychannel --name powercc --version 1.0 --package-id $PACKAGE_ID --sequence 1
   ```

5. 提交链码定义：
   ```
   peer lifecycle chaincode commit -o localhost:7050 --channelID mychannel --name powercc --version 1.0 --sequence 1
   ```