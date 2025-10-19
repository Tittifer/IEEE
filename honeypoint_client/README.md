# Honeypoint Client 蜜点后台客户端

蜜点后台客户端是一个监听区块链事件的服务，用于处理设备注册事件，并执行风险评估功能。

## 功能特点

1. 监听区块链上的设备注册和风险评分更新事件
2. 实现风险评估算法，计算设备风险评分
3. 支持命令行输入风险行为，模拟设备风险行为
4. 当设备风险评分发生变化时，向区块链报告
5. 维护风险规则

## 目录结构

```
honeypoint_client/
├── client/           # 客户端代码
│   ├── config.go     # 配置文件
│   ├── chain_client.go # 区块链客户端
│   └── honeypoint_client.go # 主客户端代码
├── chain/            # 区块链相关代码
│   └── chain_manager.go # 区块链管理器
├── risk/             # 风险评估相关代码
│   ├── assessment.go # 风险评估算法
│   └── rules.go      # 风险规则定义
├── go.mod            # Go模块文件
├── main.go           # 主程序入口
└── README.md         # 说明文档
```

## 风险规则结构

风险规则直接定义在代码中，包含以下字段：

1. `RiskRule` - 风险规则结构体
   - `BehaviorType` - 行为类型
   - `Category` - 行为类别
   - `Score` - 基础分数
   - `Weight` - 权重
   - `Description` - 描述

## 区块链存储

设备信息和风险数据存储在区块链上，包括：

- `did` - 设备DID
- `name` - 设备名称
- `model` - 设备型号
- `vendor` - 设备供应商
- `riskScore` - 历史风险分数
- `attackIndexI` - 攻击画像指数
- `attackProfile` - 攻击画像（行为类别集合）
- `lastEventTime` - 上次事件时间
- `status` - 设备状态
- `createdAt` - 创建时间
- `lastUpdatedAt` - 最后更新时间

## 风险评估算法

风险评估算法基于以下步骤：

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

其中：
- S_t：本次计算分数
- S_max：最高得分（1000分）
- S_base：当前行为基础得分
- I：攻击画像指数
- δ：影响低分时降温速度参数（0.05）
- Δt：上次风险行为和本次风险行为的时间间隔
- α：影响高分时降温速度参数（0.02）
- λ：攻击画像指数衰减系数（0.01）
- W：行为权重

## 使用方法

1. 运行程序：
   ```
   go run main.go
   ```

3. 模拟风险行为：
   ```
   risk <设备DID> <风险行为类型>
   ```

4. 查看可用的风险行为类型：
   ```
   list
   ```

5. 查看帮助：
   ```
   help
   ```

6. 退出程序：
   ```
   exit
   ```

## 风险行为类型

风险行为类型包括但不限于：

### 侦察阶段
- `visit_trap_ip` - 访问陷阱IP (10分)
- `connect_bait_wifi` - 连接诱饵WiFi (15分)
- `port_scan_honeypot` - 对蜜点进行端口扫描 (20分)

### 初始接入阶段
- `weak_password_login` - 尝试弱口令登录 (40分)
- `exploit_known_vulnerability` - 利用已知漏洞攻击 (80分)

### 执行阶段
- `execute_info_gathering` - 执行信息收集命令 (30分)
- `upload_script` - 上传脚本文件 (100分)
- `upload_known_backdoor` - 上传已知后门程序 (1000分)
- `modify_config_file` - 修改系统配置文件 (150分)

### 持久化阶段
- `create_scheduled_task` - 创建定时任务 (120分)
- `modify_system_service` - 修改系统服务 (150分)

### 防御规避阶段
- `clear_stop_log_service` - 清空或停止日志服务 (100分)
- `use_rootkit` - 使用Rootkit技术 (1000分)

### 凭证访问阶段
- `read_fake_credential` - 读取伪造的凭证文件 (200分)
- `attempt_memory_credential` - 尝试内存抓取凭证 (250分)

### 横向移动阶段
- `login_with_stolen_credential` - 使用窃取的凭证登录 (1000分)

### 数据收集阶段
- `compress_sensitive_files` - 打包压缩敏感文件 (180分)

### 渗出阶段
- `transfer_data_outside` - 向外网传输数据 (300分)
- `trigger_bait_file_callback` - 触发诱饵文件回调 (1000分)

## 风险响应策略

系统根据设备风险评分实施不同的响应策略：

1. **常规（0分）**：标准化信任与监控
2. **关注（1-199分）**：增强监控，主动引诱
3. **警戒（200-699分）**：主动欺骗与隔离引导
4. **高危（700-1000分）**：硬性阻断