package db

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// DBManager 数据库管理器
type DBManager struct {
	db *sql.DB
	chainClient ChainClient
}

// Device 设备结构体
type Device struct {
	DID           string    `json:"did"`
	Name          string    `json:"name"`
	Model         string    `json:"model"`
	Vendor        string    `json:"vendor"`
	RiskScore     float64   `json:"riskScore"`
	AttackIndexI  float64   `json:"attackIndexI"`
	AttackProfile []string  `json:"attackProfile"`
	LastEventTime time.Time `json:"lastEventTime"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"createdAt"`
	LastUpdatedAt time.Time `json:"lastUpdatedAt"`
}

// RiskRule 风险规则结构体
type RiskRule struct {
	ID           int     `json:"id"`
	BehaviorType string  `json:"behaviorType"`
	Category     string  `json:"category"`
	Score        float64 `json:"score"`
	Weight       float64 `json:"weight"`
	Description  string  `json:"description"`
}

// ChainClient 区块链客户端接口
type ChainClient interface {
	GetDeviceInfo(did string) (*Device, error)
	UpdateDeviceRiskScore(did string, riskScore float64, attackIndexI float64, attackProfile []string) error
}

// NewDBManager 创建新的数据库管理器
func NewDBManager(host string, port int, user, password, dbName string, chainClient ChainClient) (*DBManager, error) {
	// 构建DSN (Data Source Name)
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=true&loc=Local",
		user, password, host, port, dbName)
	
	log.Printf("尝试连接到数据库: %s:%d, 用户: %s, 数据库: %s", host, port, user, dbName)

	// 打开数据库连接
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %w", err)
	}

	// 设置连接池参数
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Hour)

	// 测试数据库连接
	if err := db.Ping(); err != nil {
		db.Close()
		// 尝试使用127.0.0.1替代localhost
		if host == "localhost" {
			log.Println("尝试使用127.0.0.1替代localhost连接数据库...")
			altDsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=true&loc=Local",
				user, password, "127.0.0.1", port, dbName)
			
			altDb, altErr := sql.Open("mysql", altDsn)
			if altErr != nil {
				return nil, fmt.Errorf("连接数据库失败(使用127.0.0.1): %w", altErr)
			}
			
			if altErr := altDb.Ping(); altErr != nil {
				altDb.Close()
				return nil, fmt.Errorf("测试数据库连接失败(使用127.0.0.1): %w, 原始错误: %w", altErr, err)
			}
			
			log.Println("使用127.0.0.1连接数据库成功")
			db = altDb
		} else {
			return nil, fmt.Errorf("测试数据库连接失败: %w", err)
		}
	}

	// 确保必要的表存在
	manager := &DBManager{
		db: db,
		chainClient: chainClient,
	}
	if err := manager.initTables(); err != nil {
		db.Close()
		return nil, fmt.Errorf("初始化数据库表失败: %w", err)
	}

	// 初始化风险规则
	if err := manager.initRiskRules(); err != nil {
		log.Printf("初始化风险规则失败: %v", err)
	}

	log.Println("数据库连接和初始化成功")
	return manager, nil
}

// Close 关闭数据库连接
func (m *DBManager) Close() error {
	return m.db.Close()
}

// initTables 初始化数据库表
func (m *DBManager) initTables() error {
	// 创建风险规则表
	_, err := m.db.Exec(`
		CREATE TABLE IF NOT EXISTS risk_rules (
			id INT AUTO_INCREMENT PRIMARY KEY,
			behavior_type VARCHAR(50) NOT NULL UNIQUE,
			category VARCHAR(50) NOT NULL,
			score DECIMAL(10,2) DEFAULT 0.00,
			weight DECIMAL(10,2) DEFAULT 0.00,
			description VARCHAR(255)
		)
	`)
	if err != nil {
		return fmt.Errorf("创建风险规则表失败: %w", err)
	}

	return nil
}

// initRiskRules 初始化风险规则
func (m *DBManager) initRiskRules() error {
	// 检查风险规则表是否为空
	var count int
	err := m.db.QueryRow("SELECT COUNT(*) FROM risk_rules").Scan(&count)
	if err != nil {
		return fmt.Errorf("查询风险规则表失败: %w", err)
	}

	// 如果表不为空，则不需要初始化
	if count > 0 {
		log.Printf("风险规则表中已有 %d 条规则，跳过初始化", count)
		return nil
	}

	// 定义风险规则，按照大纲中的风险评估规则表配置
	rules := []struct {
		behaviorType string
		category     string
		score        float64
		weight       float64
		description  string
	}{
		// 侦察阶段
		{"visit_trap_ip", "Recon", 10.0, 0.2, "访问陷阱IP"},
		{"connect_bait_wifi", "Recon", 15.0, 0.2, "连接诱饵WiFi"},
		{"port_scan_honeypot", "Recon", 20.0, 0.2, "对蜜点进行端口扫描"},
		
		// 初始接入阶段
		{"weak_password_login", "InitialAccess", 40.0, 0.5, "尝试弱口令登录"},
		{"exploit_known_vulnerability", "InitialAccess", 80.0, 0.8, "利用已知漏洞攻击"},
		
		// 执行阶段
		{"execute_info_gathering", "Execution", 30.0, 0.4, "执行信息收集命令"},
		{"upload_script", "Execution", 100.0, 1.0, "上传脚本文件"},
		{"upload_known_backdoor", "Execution", 1000.0, 0.0, "上传已知后门程序"},
		{"modify_config_file", "Execution", 150.0, 1.2, "修改系统配置文件"},
		
		// 持久化阶段
		{"create_scheduled_task", "Persistence", 120.0, 1.2, "创建定时任务"},
		{"modify_system_service", "Persistence", 150.0, 1.2, "修改系统服务"},
		
		// 防御规避阶段
		{"clear_stop_log_service", "DefenseEvasion", 100.0, 0.8, "清空或停止日志服务"},
		{"use_rootkit", "DefenseEvasion", 1000.0, 0.0, "使用Rootkit技术"},
		
		// 凭证访问阶段
		{"read_fake_credential", "CredentialAccess", 200.0, 1.5, "读取伪造的凭证文件"},
		{"attempt_memory_credential", "CredentialAccess", 250.0, 1.5, "尝试内存抓取凭证"},
		
		// 横向移动阶段
		{"login_with_stolen_credential", "LateralMovement", 1000.0, 0.0, "使用窃取的凭证登录"},
		
		// 数据收集阶段
		{"compress_sensitive_files", "Collection", 180.0, 1.0, "打包压缩敏感文件"},
		
		// 渗出阶段
		{"transfer_data_outside", "Exfiltration", 300.0, 1.8, "向外网传输数据"},
		{"trigger_bait_file_callback", "Exfiltration", 1000.0, 0.0, "触发诱饵文件回调"},
	}

	// 插入风险规则
	for _, rule := range rules {
		_, err := m.db.Exec(
			"INSERT INTO risk_rules (behavior_type, category, score, weight, description) VALUES (?, ?, ?, ?, ?)",
			rule.behaviorType, rule.category, rule.score, rule.weight, rule.description,
		)
		if err != nil {
			return fmt.Errorf("插入风险规则失败: %w", err)
		}
	}

	log.Println("成功初始化风险规则")
	return nil
}

// GetDeviceFromChain 从区块链获取设备信息
func (m *DBManager) GetDeviceFromChain(did string) (*Device, error) {
	return m.chainClient.GetDeviceInfo(did)
}

// UpdateDeviceAttackIndex 更新设备攻击画像指数
func (m *DBManager) UpdateDeviceAttackIndex(did string, attackIndexI float64) error {
	// 先从链上获取设备信息
	device, err := m.GetDeviceFromChain(did)
	if err != nil {
		return fmt.Errorf("获取设备信息失败: %w", err)
	}
	
	// 更新攻击画像指数
	return m.chainClient.UpdateDeviceRiskScore(did, device.RiskScore, attackIndexI, device.AttackProfile)
}

// GetRiskRuleByType 根据行为类型获取风险规则
func (m *DBManager) GetRiskRuleByType(behaviorType string) (*RiskRule, error) {
	var rule RiskRule
	err := m.db.QueryRow(
		"SELECT id, behavior_type, category, score, weight, description FROM risk_rules WHERE behavior_type = ?",
		behaviorType,
	).Scan(&rule.ID, &rule.BehaviorType, &rule.Category, &rule.Score, &rule.Weight, &rule.Description)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("获取风险规则失败: %w", err)
	}

	return &rule, nil
}

// GetAllRiskRules 获取所有风险规则
func (m *DBManager) GetAllRiskRules() ([]RiskRule, error) {
	rows, err := m.db.Query(
		"SELECT id, behavior_type, category, score, weight, description FROM risk_rules ORDER BY id",
	)
	if err != nil {
		return nil, fmt.Errorf("获取风险规则失败: %w", err)
	}
	defer rows.Close()

	var rules []RiskRule
	for rows.Next() {
		var rule RiskRule
		if err := rows.Scan(&rule.ID, &rule.BehaviorType, &rule.Category, &rule.Score, &rule.Weight, &rule.Description); err != nil {
			return nil, fmt.Errorf("扫描风险规则记录失败: %w", err)
		}
		rules = append(rules, rule)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("迭代风险规则记录失败: %w", err)
	}

	return rules, nil
}

// ResetDeviceRiskData 重置设备风险数据
func (m *DBManager) ResetDeviceRiskData(did string) error {
	// 重置链上设备风险评分和攻击画像
	emptyProfile := make([]string, 0)
	err := m.chainClient.UpdateDeviceRiskScore(did, 0.0, 0.0, emptyProfile)
	if err != nil {
		return fmt.Errorf("重置设备风险评分失败: %w", err)
	}
	
	log.Printf("已成功重置设备 %s 的风险数据", did)
	return nil
}