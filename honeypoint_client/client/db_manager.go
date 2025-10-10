package client

import (
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// 数据库配置
type DBConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
}

// 用户信息结构体
type UserDB struct {
	ID          int
	DID         string
	Name        string
	CurrentScore int
	LastUpdate  time.Time
	CreatedAt   time.Time
}

// 风险行为记录结构体
type RiskBehaviorRecord struct {
	ID           int
	UserID       int
	BehaviorType string
	Score        int
	Timestamp    time.Time
}

// 风险规则结构体
type RiskRule struct {
	ID           int
	BehaviorType string
	Score        int
	Description  string
	CreatedAt    time.Time
}

// 数据库管理器
type DBManager struct {
	db     *sql.DB
	config *DBConfig
	mutex  sync.RWMutex
}

// 创建新的数据库管理器
func NewDBManager(config *DBConfig) (*DBManager, error) {
	// 构建DSN连接字符串
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.User, config.Password, config.Host, config.Port, config.DBName)

	// 连接数据库
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %w", err)
	}

	// 测试连接
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("测试数据库连接失败: %w", err)
	}

	// 设置连接池参数
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Hour)

	return &DBManager{
		db:     db,
		config: config,
	}, nil
}

// 关闭数据库连接
func (m *DBManager) Close() error {
	return m.db.Close()
}

// 注册新用户
func (m *DBManager) RegisterUser(did, name string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 检查用户是否已存在
	var exists bool
	err := m.db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE did = ?)", did).Scan(&exists)
	if err != nil {
		return fmt.Errorf("检查用户是否存在失败: %w", err)
	}

	if exists {
		return fmt.Errorf("用户 %s 已存在", did)
	}

	// 插入新用户
	_, err = m.db.Exec("INSERT INTO users (did, name, current_score) VALUES (?, ?, 0)", did, name)
	if err != nil {
		return fmt.Errorf("注册用户失败: %w", err)
	}

	log.Printf("用户 %s (DID: %s) 已成功注册到数据库", name, did)
	return nil
}

// 获取用户信息
func (m *DBManager) GetUser(did string) (*UserDB, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	user := &UserDB{}
	err := m.db.QueryRow("SELECT id, did, name, current_score, last_update, created_at FROM users WHERE did = ?", did).
		Scan(&user.ID, &user.DID, &user.Name, &user.CurrentScore, &user.LastUpdate, &user.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("用户 %s 不存在", did)
		}
		return nil, fmt.Errorf("获取用户信息失败: %w", err)
	}

	return user, nil
}

// 记录风险行为
func (m *DBManager) RecordRiskBehavior(did, behaviorType string) (int, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 获取用户ID
	user, err := m.GetUser(did)
	if err != nil {
		return 0, err
	}

	// 获取风险规则
	var ruleID int
	var score int
	err = m.db.QueryRow("SELECT id, score FROM risk_rules WHERE behavior_type = ?", behaviorType).
		Scan(&ruleID, &score)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("风险行为类型 %s 不存在", behaviorType)
		}
		return 0, fmt.Errorf("获取风险规则失败: %w", err)
	}

	// 记录风险行为
	_, err = m.db.Exec("INSERT INTO risk_behaviors (user_id, behavior_type, score) VALUES (?, ?, ?)",
		user.ID, behaviorType, score)
	if err != nil {
		return 0, fmt.Errorf("记录风险行为失败: %w", err)
	}

	log.Printf("用户 %s 的风险行为 %s (分数: %d) 已记录", did, behaviorType, score)
	return score, nil
}

// 获取用户历史风险行为
func (m *DBManager) GetUserRiskBehaviors(did string, limit int) ([]RiskBehaviorRecord, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// 获取用户ID
	user, err := m.GetUser(did)
	if err != nil {
		return nil, err
	}

	// 查询用户历史风险行为
	rows, err := m.db.Query(
		"SELECT id, user_id, behavior_type, score, timestamp FROM risk_behaviors WHERE user_id = ? ORDER BY timestamp DESC LIMIT ?",
		user.ID, limit)
	if err != nil {
		return nil, fmt.Errorf("查询用户历史风险行为失败: %w", err)
	}
	defer rows.Close()

	var behaviors []RiskBehaviorRecord
	for rows.Next() {
		var behavior RiskBehaviorRecord
		if err := rows.Scan(&behavior.ID, &behavior.UserID, &behavior.BehaviorType, &behavior.Score, &behavior.Timestamp); err != nil {
			return nil, fmt.Errorf("扫描风险行为记录失败: %w", err)
		}
		behaviors = append(behaviors, behavior)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("迭代风险行为记录失败: %w", err)
	}

	return behaviors, nil
}

// 更新用户风险评分
func (m *DBManager) UpdateUserRiskScore(did string, score int) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 更新用户风险评分
	result, err := m.db.Exec("UPDATE users SET current_score = ?, last_update = NOW() WHERE did = ?", score, did)
	if err != nil {
		return fmt.Errorf("更新用户风险评分失败: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取受影响行数失败: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("用户 %s 不存在", did)
	}

	log.Printf("用户 %s 的风险评分已更新为 %d", did, score)
	return nil
}

// 获取风险规则
func (m *DBManager) GetRiskRule(behaviorType string) (*RiskRule, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	rule := &RiskRule{}
	err := m.db.QueryRow("SELECT id, behavior_type, score, description, created_at FROM risk_rules WHERE behavior_type = ?", behaviorType).
		Scan(&rule.ID, &rule.BehaviorType, &rule.Score, &rule.Description, &rule.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("风险规则 %s 不存在", behaviorType)
		}
		return nil, fmt.Errorf("获取风险规则失败: %w", err)
	}

	return rule, nil
}

// 获取所有风险规则
func (m *DBManager) GetAllRiskRules() ([]RiskRule, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	rows, err := m.db.Query("SELECT id, behavior_type, score, description, created_at FROM risk_rules")
	if err != nil {
		return nil, fmt.Errorf("查询风险规则失败: %w", err)
	}
	defer rows.Close()

	var rules []RiskRule
	for rows.Next() {
		var rule RiskRule
		if err := rows.Scan(&rule.ID, &rule.BehaviorType, &rule.Score, &rule.Description, &rule.CreatedAt); err != nil {
			return nil, fmt.Errorf("扫描风险规则失败: %w", err)
		}
		rules = append(rules, rule)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("迭代风险规则失败: %w", err)
	}

	return rules, nil
}


