# Honeypoint Client 蜜点后台客户端

蜜点后台客户端是一个监听区块链事件的服务，用于处理用户身份注册和登录事件，并执行风险评估功能。

## 功能特点

1. 监听区块链上的用户注册、登录和登出事件
2. 将用户信息存储到MySQL数据库中
3. 实现风险评估算法，计算用户风险评分
4. 支持命令行输入风险行为，模拟用户风险行为
5. 当用户风险评分超过阈值时，向区块链报告

## 目录结构

```
honeypoint_client/
├── client/           # 客户端代码
│   ├── config.go     # 配置文件
│   └── honeypoint_client.go # 主客户端代码
├── db/               # 数据库相关代码
│   └── database.go   # 数据库操作
├── risk/             # 风险评估相关代码
│   └── assessment.go # 风险评估算法
├── go.mod            # Go模块文件
├── main.go           # 主程序入口
└── README.md         # 说明文档
```

## 数据库结构

蜜点后台客户端使用MySQL数据库，包含以下表：

1. `users` - 存储用户信息
   - `id` - 主键
   - `did` - 用户DID
   - `name` - 用户名称
   - `current_score` - 当前风险评分
   - `last_update` - 最后更新时间
   - `created_at` - 创建时间

2. `risk_rules` - 存储风险规则
   - `id` - 主键
   - `behavior_type` - 行为类型
   - `score` - 分数
   - `description` - 描述
   - `created_at` - 创建时间

3. `risk_behaviors` - 存储风险行为记录
   - `id` - 主键
   - `user_id` - 用户ID
   - `behavior_type` - 行为类型
   - `score` - 分数
   - `timestamp` - 时间戳

## 风险评估算法

风险评估算法基于以下公式：

1. 分数更新：
   ```
   S_t = min(S_max, Score · λ)
   ```
   - S_t：本次计算分数
   - S_max：最高得分
   - Score：当前行为得分
   - λ：历史风险行为影响因子

2. 历史风险行为影响因子：
   ```
   λ = 1 + (∑(e^(-β(t-t_j)) · S_p,j)) / T
   ```
   - N：用户历史风险行为记录总数
   - T：风险分数归一化基数
   - β：衰减速率系数
   - S_p,j：第j次风险行为的分数
   - t_j：第j次风险行为发生时间

## 使用方法

1. 确保MySQL数据库已启动，并创建了`ieee_honeypoint`数据库
2. 运行程序：
   ```
   go run main.go
   ```

3. 模拟风险行为：
   ```
   risk <DID> <风险行为类型>
   ```

4. 查看帮助：
   ```
   help
   ```

5. 退出程序：
   ```
   exit
   ```
   或
   ```
   quit
   ```

## 风险行为类型

风险行为类型包括但不限于：

- `edge_honeypot` - 访问非关键蜜点 (2分)
- `business_honeypot` - 访问业务处理类蜜点 (6分)
- `core_honeypot` - 访问模拟核心控制系统的蜜点 (9分)
- `linear_access` - 按预设路径顺序访问蜜点 (1分)
- `multi_branch_access` - 在多个蜜点分支间跳跃访问 (4分)
- `cross_domain_access` - 跨越不同安全域访问蜜点 (7分)
- 更多类型请参考程序输出的风险行为列表
