package db

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// DBManager 数据库管理器
type DBManager struct {
	db *sql.DB
}

// User 用户结构体
type User struct {
	ID           int       `json:"id"`
	DID          string    `json:"did"`
	Name         string    `json:"name"`
	CurrentScore float64   `json:"currentScore"` // 修改为float64类型
	LastUpdate   time.Time `json:"lastUpdate"`
	CreatedAt    time.Time `json:"createdAt"`
}

// RiskRule 风险规则结构体
type RiskRule struct {
	ID           int     `json:"id"`
	BehaviorType string  `json:"behaviorType"`
	Score        float64 `json:"score"` // 修改为float64类型
	Description  string  `json:"description"`
}

// RiskBehavior 风险行为记录结构体
type RiskBehavior struct {
	ID           int       `json:"id"`
	BehaviorType string    `json:"behaviorType"`
	Score        float64   `json:"score"` // 修改为float64类型
	Timestamp    time.Time `json:"timestamp"`
}

// NewDBManager 创建新的数据库管理器
func NewDBManager(host string, port int, user, password, dbName string) (*DBManager, error) {
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
	manager := &DBManager{db: db}
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
	// 创建用户表
	_, err := m.db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id INT AUTO_INCREMENT PRIMARY KEY,
			did VARCHAR(255) NOT NULL UNIQUE,
			name VARCHAR(255),
			current_score DECIMAL(10,2) DEFAULT 0.00,
			last_update TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("创建用户表失败: %w", err)
	}

	// 创建风险规则表
	_, err = m.db.Exec(`
		CREATE TABLE IF NOT EXISTS risk_rules (
			id INT AUTO_INCREMENT PRIMARY KEY,
			behavior_type VARCHAR(50) NOT NULL UNIQUE,
			score DECIMAL(10,2) DEFAULT 0.00,
			description VARCHAR(255)
		)
	`)
	if err != nil {
		return fmt.Errorf("创建风险规则表失败: %w", err)
	}

	return nil
}

// getSafeTableName 获取安全的表名
func getSafeTableName(did string) string {
	// 替换非法字符
	safeName := strings.ReplaceAll(did, ":", "_")
	safeName = strings.ReplaceAll(safeName, "-", "_")
	safeName = strings.ReplaceAll(safeName, ".", "_")
	return "user_behaviors_" + safeName
}

// createUserBehaviorTable 创建用户行为表
func (m *DBManager) createUserBehaviorTable(did string) error {
	tableName := getSafeTableName(did)
	
	// 检查表是否已存在
	var tableExists bool
	query := fmt.Sprintf("SHOW TABLES LIKE '%s'", tableName)
	err := m.db.QueryRow(query).Scan(&tableName)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("检查表是否存在失败: %w", err)
	}
	tableExists = (err != sql.ErrNoRows)
	
	if !tableExists {
		// 创建用户行为表
		createTableSQL := fmt.Sprintf(`
			CREATE TABLE %s (
				id INT AUTO_INCREMENT PRIMARY KEY,
				behavior_type VARCHAR(50) NOT NULL,
				score DECIMAL(10,2) DEFAULT 0.00,
				timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			)
		`, tableName)
		
		_, err = m.db.Exec(createTableSQL)
		if err != nil {
			return fmt.Errorf("创建用户行为表失败: %w", err)
		}
		
		log.Printf("为用户 %s 创建行为表 %s 成功", did, tableName)
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

	// 定义风险规则
	rules := []struct {
		behaviorType string
		score        float64 // 修改为float64类型
		description  string
	}{
		// 路径深度相关规则
		{"edge_honeypot", 2.00, "访问非关键蜜点"},
		{"business_honeypot", 6.00, "访问业务处理类蜜点"},
		{"core_honeypot", 9.00, "访问模拟核心控制系统的蜜点"},
		
		// 路径复杂度相关规则
		{"linear_access", 1.00, "按预设路径顺序访问蜜点"},
		{"multi_branch_access", 4.00, "在多个蜜点分支间跳跃访问"},
		{"cross_domain_access", 7.00, "跨越不同安全域访问蜜点"},
		
		// 权限提升相关规则
		{"no_privilege_escalation", 0.00, "保持初始权限访问蜜点"},
		{"normal_privilege_escalation", 4.00, "获取普通用户权限蜜点访问权"},
		{"admin_privilege_escalation", 7.00, "访问模拟高权限系统的蜜点"},
		
		// 节点关键性相关规则
		{"info_display_honeypot", 1.00, "访问只读信息类蜜点"},
		{"data_collection_honeypot", 4.00, "访问数据采集模拟节点"},
		{"business_control_honeypot", 7.00, "访问业务控制模拟节点"},
		{"core_control_honeypot", 10.00, "访问模拟核心控制系统的蜜点"},
		
		// 操作危险性相关规则
		{"harmless_probe", 0.00, "仅调用信息查询接口"},
		{"sensitive_info_query", 2.00, "调用敏感信息接口"},
		{"parameter_modification", 5.00, "调用参数修改接口"},
		{"control_command", 8.00, "调用控制指令接口"},
		
		// 目标性相关规则
		{"random_scan", 1.00, "无目标调用多个蜜点"},
		{"targeted_wrong_path", 3.00, "目标明确但调用路径不合理"},
		{"precise_high_value", 6.00, "直接调用高价值蜜点接口"},
		
		// 调用频率相关规则
		{"low_frequency", 1.00, "调用间隔 > 10分钟"},
		{"medium_frequency", 3.00, "调用间隔 1-10分钟"},
		{"high_frequency", 6.00, "调用间隔 < 1分钟"},
		
		// 一票否决规则
		{"destructive_command", 100.00, "调用模拟破坏性指令的蜜点接口"},
		{"highest_privilege", 100.00, "调用模拟系统最高权限的蜜点接口"},
		{"coordinated_attack", 100.00, "同时从多个入口调用蜜点，表现出协同攻击特征"},
	}

	// 插入风险规则
	for _, rule := range rules {
		_, err := m.db.Exec(
			"INSERT INTO risk_rules (behavior_type, score, description) VALUES (?, ?, ?)",
			rule.behaviorType, rule.score, rule.description,
		)
		if err != nil {
			return fmt.Errorf("插入风险规则失败: %w", err)
		}
	}

	log.Println("成功初始化风险规则")
	return nil
}

// CreateUser 创建用户
func (m *DBManager) CreateUser(did, name string) (int, error) {
	// 检查用户是否已存在
	var exists bool
	err := m.db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE did = ?)", did).Scan(&exists)
	if err != nil {
		return 0, fmt.Errorf("检查用户是否存在失败: %w", err)
	}

	if exists {
		return 0, fmt.Errorf("用户DID %s 已存在", did)
	}

	// 插入新用户
	result, err := m.db.Exec(
		"INSERT INTO users (did, name, current_score) VALUES (?, ?, ?)",
		did, name, 0.00, // 修改为浮点数
	)
	if err != nil {
		return 0, fmt.Errorf("创建用户失败: %w", err)
	}

	// 获取新用户ID
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("获取用户ID失败: %w", err)
	}
	
	// 创建用户行为表
	if err := m.createUserBehaviorTable(did); err != nil {
		log.Printf("警告: 创建用户行为表失败: %v", err)
	}

	return int(id), nil
}

// GetUserByDID 根据DID获取用户
func (m *DBManager) GetUserByDID(did string) (*User, error) {
	var user User
	err := m.db.QueryRow(
		"SELECT id, did, name, current_score, last_update, created_at FROM users WHERE did = ?",
		did,
	).Scan(&user.ID, &user.DID, &user.Name, &user.CurrentScore, &user.LastUpdate, &user.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("获取用户失败: %w", err)
	}
	
	// 确保用户行为表存在
	if err := m.createUserBehaviorTable(did); err != nil {
		log.Printf("警告: 确保用户行为表存在失败: %v", err)
	}

	return &user, nil
}

// UpdateUserRiskScore 更新用户风险评分
func (m *DBManager) UpdateUserRiskScore(userID int, newScore float64) error { // 修改为float64类型
	_, err := m.db.Exec(
		"UPDATE users SET current_score = ?, last_update = CURRENT_TIMESTAMP WHERE id = ?",
		newScore, userID,
	)
	if err != nil {
		return fmt.Errorf("更新用户风险评分失败: %w", err)
	}

	return nil
}

// GetRiskRuleByType 根据行为类型获取风险规则
func (m *DBManager) GetRiskRuleByType(behaviorType string) (*RiskRule, error) {
	var rule RiskRule
	err := m.db.QueryRow(
		"SELECT id, behavior_type, score, description FROM risk_rules WHERE behavior_type = ?",
		behaviorType,
	).Scan(&rule.ID, &rule.BehaviorType, &rule.Score, &rule.Description)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("获取风险规则失败: %w", err)
	}

	return &rule, nil
}

// RecordRiskBehavior 记录用户风险行为
func (m *DBManager) RecordRiskBehavior(userID int, behaviorType string, score float64) error { // 修改为float64类型
	// 首先获取用户DID
	var did string
	err := m.db.QueryRow("SELECT did FROM users WHERE id = ?", userID).Scan(&did)
	if err != nil {
		return fmt.Errorf("获取用户DID失败: %w", err)
	}
	
	// 确保用户行为表存在
	if err := m.createUserBehaviorTable(did); err != nil {
		return fmt.Errorf("确保用户行为表存在失败: %w", err)
	}
	
	// 获取表名
	tableName := getSafeTableName(did)
	
	// 插入风险行为记录
	insertSQL := fmt.Sprintf(
		"INSERT INTO %s (behavior_type, score) VALUES (?, ?)",
		tableName,
	)
	_, err = m.db.Exec(insertSQL, behaviorType, score)
	if err != nil {
		return fmt.Errorf("记录风险行为失败: %w", err)
	}

	return nil
}

// GetUserRiskBehaviors 获取用户历史风险行为
func (m *DBManager) GetUserRiskBehaviors(userID int, limit int) ([]RiskBehavior, error) {
	// 首先获取用户DID
	var did string
	err := m.db.QueryRow("SELECT did FROM users WHERE id = ?", userID).Scan(&did)
	if err != nil {
		return nil, fmt.Errorf("获取用户DID失败: %w", err)
	}
	
	// 确保用户行为表存在
	if err := m.createUserBehaviorTable(did); err != nil {
		return nil, fmt.Errorf("确保用户行为表存在失败: %w", err)
	}
	
	// 获取表名
	tableName := getSafeTableName(did)
	
	// 查询风险行为记录
	querySQL := fmt.Sprintf(
		"SELECT id, behavior_type, score, timestamp FROM %s ORDER BY timestamp DESC LIMIT ?",
		tableName,
	)
	rows, err := m.db.Query(querySQL, limit)
	if err != nil {
		return nil, fmt.Errorf("获取用户风险行为失败: %w", err)
	}
	defer rows.Close()

	var behaviors []RiskBehavior
	for rows.Next() {
		var behavior RiskBehavior
		if err := rows.Scan(&behavior.ID, &behavior.BehaviorType, &behavior.Score, &behavior.Timestamp); err != nil {
			return nil, fmt.Errorf("扫描风险行为记录失败: %w", err)
		}
		behaviors = append(behaviors, behavior)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("迭代风险行为记录失败: %w", err)
	}

	return behaviors, nil
}

// GetAllRiskRules 获取所有风险规则
func (m *DBManager) GetAllRiskRules() ([]RiskRule, error) {
	rows, err := m.db.Query(
		"SELECT id, behavior_type, score, description FROM risk_rules ORDER BY id",
	)
	if err != nil {
		return nil, fmt.Errorf("获取风险规则失败: %w", err)
	}
	defer rows.Close()

	var rules []RiskRule
	for rows.Next() {
		var rule RiskRule
		if err := rows.Scan(&rule.ID, &rule.BehaviorType, &rule.Score, &rule.Description); err != nil {
			return nil, fmt.Errorf("扫描风险规则记录失败: %w", err)
		}
		rules = append(rules, rule)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("迭代风险规则记录失败: %w", err)
	}

	return rules, nil
}