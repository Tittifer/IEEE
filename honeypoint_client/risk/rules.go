package risk

// RiskRule 风险规则结构体
type RiskRule struct {
	BehaviorType string  // 行为类型标识符
	Category     string  // 行为类别
	Score        float64 // 基础风险分数
	Weight       float64 // 权重
	Description  string  // 描述
}

// RiskRules 所有风险规则的集合
var RiskRules = []RiskRule{
	// 侦察阶段
	{
		BehaviorType: "visit_trap_ip",
		Category:     "Recon.NetworkScan",
		Score:        10.0,
		Weight:       0.2,
		Description:  "访问陷阱IP",
	},
	{
		BehaviorType: "connect_bait_wifi",
		Category:     "Recon.WirelessScan",
		Score:        15.0,
		Weight:       0.2,
		Description:  "连接诱饵WiFi",
	},
	{
		BehaviorType: "port_scan_honeypot",
		Category:     "Recon.PortScan",
		Score:        20.0,
		Weight:       0.2,
		Description:  "对蜜点进行端口扫描",
	},

	// 初始接入阶段
	{
		BehaviorType: "weak_password_login",
		Category:     "InitialAccess.WeakCred",
		Score:        40.0,
		Weight:       0.5,
		Description:  "尝试弱口令登录",
	},
	{
		BehaviorType: "exploit_known_vulnerability",
		Category:     "InitialAccess.Exploit",
		Score:        80.0,
		Weight:       0.8,
		Description:  "利用已知漏洞攻击",
	},

	// 执行阶段
	{
		BehaviorType: "execute_info_gathering",
		Category:     "Execution.Discovery",
		Score:        30.0,
		Weight:       0.4,
		Description:  "执行信息收集命令",
	},
	{
		BehaviorType: "upload_script",
		Category:     "Execution.FileUpload",
		Score:        100.0,
		Weight:       1.0,
		Description:  "上传脚本文件",
	},
	{
		BehaviorType: "upload_known_backdoor",
		Category:     "Execution.Malware",
		Score:        1000.0,
		Weight:       0.0,
		Description:  "上传已知后门程序",
	},
	{
		BehaviorType: "modify_config_file",
		Category:     "Execution.Tamper",
		Score:        150.0,
		Weight:       1.2,
		Description:  "修改系统配置文件",
	},

	// 持久化阶段
	{
		BehaviorType: "create_scheduled_task",
		Category:     "Persistence.CronJob",
		Score:        120.0,
		Weight:       1.2,
		Description:  "创建定时任务",
	},
	{
		BehaviorType: "modify_system_service",
		Category:     "Persistence.ServiceMod",
		Score:        150.0,
		Weight:       1.2,
		Description:  "修改系统服务",
	},

	// 防御规避阶段
	{
		BehaviorType: "clear_stop_log_service",
		Category:     "DefenseEvasion.ClearLog",
		Score:        100.0,
		Weight:       0.8,
		Description:  "清空或停止日志服务",
	},
	{
		BehaviorType: "use_rootkit",
		Category:     "DefenseEvasion.Rootkit",
		Score:        1000.0,
		Weight:       0.0,
		Description:  "使用Rootkit技术",
	},

	// 凭证访问阶段
	{
		BehaviorType: "read_fake_credential",
		Category:     "CredentialAccess.File",
		Score:        200.0,
		Weight:       1.5,
		Description:  "读取伪造的凭证文件",
	},
	{
		BehaviorType: "attempt_memory_credential",
		Category:     "CredentialAccess.Memory",
		Score:        250.0,
		Weight:       1.5,
		Description:  "尝试内存抓取凭证",
	},

	// 横向移动阶段
	{
		BehaviorType: "login_with_stolen_credential",
		Category:     "LateralMovement.StolenCred",
		Score:        1000.0,
		Weight:       0.0,
		Description:  "使用窃取的凭证登录",
	},

	// 数据收集阶段
	{
		BehaviorType: "compress_sensitive_files",
		Category:     "Collection.Archive",
		Score:        180.0,
		Weight:       1.0,
		Description:  "打包压缩敏感文件",
	},

	// 渗出阶段
	{
		BehaviorType: "transfer_data_outside",
		Category:     "Exfiltration.DataTransfer",
		Score:        300.0,
		Weight:       1.8,
		Description:  "向外网传输数据",
	},
	{
		BehaviorType: "trigger_bait_file_callback",
		Category:     "Exfiltration.CanaryToken",
		Score:        1000.0,
		Weight:       0.0,
		Description:  "触发诱饵文件回调",
	},
}

// GetRiskRuleByType 根据行为类型获取风险规则
func GetRiskRuleByType(behaviorType string) *RiskRule {
	for _, rule := range RiskRules {
		if rule.BehaviorType == behaviorType {
			return &rule
		}
	}
	return nil
}

// GetAllRiskRules 获取所有风险规则
func GetAllRiskRules() []RiskRule {
	return RiskRules
}
