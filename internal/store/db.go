package store

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

func InitDB(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", path+"?_journal_mode=WAL&_busy_timeout=5000&_foreign_keys=on")
	if err != nil {
		return nil, err
	}
	// 写连接限1个，读可并发
	db.SetMaxOpenConns(4)
	db.SetMaxIdleConns(2)
	if err := migrate(db); err != nil {
		return nil, err
	}
	return db, nil
}

func migrate(db *sql.DB) error {
	// Schema 版本管理
	db.Exec(`CREATE TABLE IF NOT EXISTS schema_version (version INTEGER NOT NULL DEFAULT 0)`)
	var ver int
	row := db.QueryRow(`SELECT version FROM schema_version LIMIT 1`)
	if row.Scan(&ver) != nil {
		db.Exec(`INSERT INTO schema_version (version) VALUES (1)`)
		ver = 0
	}

	schema := `

	-- ========================================
	-- 角色定义表（系统预置，用于权限等级判断）
	-- ========================================
	CREATE TABLE IF NOT EXISTS roles (
		id    INTEGER PRIMARY KEY AUTOINCREMENT,
		name  TEXT UNIQUE NOT NULL,              -- 角色标识: admin/secretary/villager 等
		label TEXT NOT NULL,                     -- 角色中文名: 系统管理员/党支部书记 等
		level INTEGER NOT NULL DEFAULT 0         -- 权限等级: 数值越大权限越高
	);

	-- ========================================
	-- 用户表（村民/干部/管理员统一存储）
	-- ========================================
	CREATE TABLE IF NOT EXISTS users (
		id                INTEGER PRIMARY KEY AUTOINCREMENT,
		account           TEXT UNIQUE NOT NULL DEFAULT '',  -- 登录账号（手机号/admin/wx_openid）
		phone             TEXT NOT NULL DEFAULT '',         -- 真实手机号（可后续绑定）
		name              TEXT NOT NULL,                    -- 姓名
		gender            TEXT NOT NULL DEFAULT '',         -- 性别: male/female/空
		birth_date        TEXT NOT NULL DEFAULT '',         -- 出生日期: YYYY-MM-DD
		ethnicity         TEXT NOT NULL DEFAULT '汉族',     -- 民族
		education         TEXT NOT NULL DEFAULT '',         -- 文化程度: 小学/初中/高中/大专/本科/硕士及以上/文盲
		marital_status    TEXT NOT NULL DEFAULT '',         -- 婚姻状况: unmarried/married/divorced/widowed
		role              TEXT NOT NULL DEFAULT 'villager', -- 系统角色（逗号分隔，如 admin,secretary）
		position          TEXT NOT NULL DEFAULT 'villager', -- 具体职务: party_secretary/village_director 等
		id_card           TEXT NOT NULL DEFAULT '',         -- 身份证号
		address           TEXT NOT NULL DEFAULT '',         -- 详细地址
		group_id          INTEGER NOT NULL DEFAULT 0,      -- 所属村民小组 → groups.id
		household_id      INTEGER NOT NULL DEFAULT 0,      -- 所属户籍 → households.id
		is_party_member   INTEGER NOT NULL DEFAULT 0,      -- 是否党员: 0/1
		is_low_income     INTEGER NOT NULL DEFAULT 0,      -- 是否低保户: 0/1
		is_five_guarantee INTEGER NOT NULL DEFAULT 0,      -- 是否五保户: 0/1
		is_disabled       INTEGER NOT NULL DEFAULT 0,      -- 是否残疾人: 0/1
		is_military       INTEGER NOT NULL DEFAULT 0,      -- 是否军属/退役军人: 0/1
		wechat_id         TEXT NOT NULL DEFAULT '',        -- 微信号
		emergency_contact TEXT NOT NULL DEFAULT '',        -- 紧急联系人姓名
		emergency_phone   TEXT NOT NULL DEFAULT '',        -- 紧急联系人电话
		openid            TEXT NOT NULL DEFAULT '',        -- 微信小程序 openid
		avatar_url        TEXT NOT NULL DEFAULT '',        -- 微信头像 URL
		active            INTEGER NOT NULL DEFAULT 1,      -- 是否启用: 0禁用/1启用
		password_hash     TEXT NOT NULL DEFAULT '',        -- 密码 bcrypt 哈希
		remark            TEXT NOT NULL DEFAULT '',        -- 备注
		created_at        DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at        DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	-- ========================================
	-- 户籍表（以户为单位管理家庭信息）
	-- ========================================
	CREATE TABLE IF NOT EXISTS households (
		id            INTEGER PRIMARY KEY AUTOINCREMENT,
		household_no  TEXT NOT NULL DEFAULT '',    -- 户号（如 001）
		head_id       INTEGER NOT NULL DEFAULT 0,  -- 户主 → users.id
		address       TEXT NOT NULL DEFAULT '',     -- 户籍地址
		group_id      INTEGER NOT NULL DEFAULT 0,  -- 所属小组 → groups.id
		farmland_area REAL NOT NULL DEFAULT 0,     -- 全户承包耕地面积（亩）
		forest_area   REAL NOT NULL DEFAULT 0,     -- 全户林地面积（亩）
		homesite_area REAL NOT NULL DEFAULT 0,     -- 宅基地面积（平方米）
		remark        TEXT NOT NULL DEFAULT '',     -- 备注
		created_at    DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	-- ========================================
	-- 户籍成员关联表（一户多人）
	-- ========================================
	CREATE TABLE IF NOT EXISTS household_members (
		id           INTEGER PRIMARY KEY AUTOINCREMENT,
		household_id INTEGER NOT NULL,             -- 所属户籍 → households.id
		user_id      INTEGER NOT NULL,             -- 成员 → users.id
		relation     TEXT NOT NULL DEFAULT '其他',  -- 与户主关系: 户主/配偶/子女/父母/其他
		FOREIGN KEY (household_id) REFERENCES households(id) ON DELETE CASCADE,
		FOREIGN KEY (user_id) REFERENCES users(id)
	);

	-- ========================================
	-- 村民小组表
	-- ========================================
	CREATE TABLE IF NOT EXISTS groups (
		id         INTEGER PRIMARY KEY AUTOINCREMENT,
		name       TEXT NOT NULL,                  -- 小组名称（如 第一组）
		leader_id  INTEGER NOT NULL DEFAULT 0,     -- 组长 → users.id
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	-- ========================================
	-- 公告表
	-- 工作流: draft → pending_review → published / rejected
	-- ========================================
	CREATE TABLE IF NOT EXISTS notices (
		id             INTEGER PRIMARY KEY AUTOINCREMENT,
		title          TEXT NOT NULL,               -- 标题
		content        TEXT NOT NULL,               -- 内容（支持 HTML 富文本）
		category       TEXT NOT NULL DEFAULT 'policy', -- 分类: policy/activity/urgent/meeting
		author_id      INTEGER NOT NULL,            -- 作者 → users.id
		pinned         INTEGER NOT NULL DEFAULT 0,  -- 是否置顶: 0/1
		attachments    TEXT NOT NULL DEFAULT '[]',  -- 附件 JSON 数组
		views          INTEGER NOT NULL DEFAULT 0,  -- 阅读次数
		workflow_state TEXT NOT NULL DEFAULT 'draft', -- 状态: draft/pending_review/published/rejected
		reviewer_id    INTEGER NOT NULL DEFAULT 0,  -- 审核人 → users.id
		review_note    TEXT NOT NULL DEFAULT '',     -- 审核意见
		published_at   DATETIME,                    -- 发布时间
		created_at     DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at     DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (author_id) REFERENCES users(id)
	);

	-- ========================================
	-- 财务记录表
	-- 工作流: draft → pending_review → approved / rejected
	-- ========================================
	CREATE TABLE IF NOT EXISTS finance_records (
		id             INTEGER PRIMARY KEY AUTOINCREMENT,
		type           TEXT NOT NULL CHECK(type IN ('income','expense')), -- 类型: income收入/expense支出
		amount         INTEGER NOT NULL,            -- 金额（分）
		category       TEXT NOT NULL DEFAULT '',    -- 分类（如 上级拨款/基础设施）
		remark         TEXT NOT NULL DEFAULT '',    -- 备注
		date           TEXT NOT NULL,               -- 日期: YYYY-MM-DD
		voucher        TEXT NOT NULL DEFAULT '',    -- 凭证图片 URL
		author_id      INTEGER NOT NULL,            -- 录入人 → users.id
		workflow_state TEXT NOT NULL DEFAULT 'draft', -- 状态: draft/pending_review/approved/rejected
		reviewer_id    INTEGER NOT NULL DEFAULT 0,  -- 审核人 → users.id
		review_note    TEXT NOT NULL DEFAULT '',     -- 审核意见
		created_at     DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (author_id) REFERENCES users(id)
	);

	-- ========================================
	-- 补贴申请表（两级审批）
	-- 工作流: submitted → committee_review → secretary_review → approved / rejected
	-- ========================================
	CREATE TABLE IF NOT EXISTS subsidies (
		id             INTEGER PRIMARY KEY AUTOINCREMENT,
		title          TEXT NOT NULL,               -- 补贴名称
		type           TEXT NOT NULL DEFAULT 'other', -- 类型: farming/medical/education/housing/other
		amount         INTEGER NOT NULL,            -- 申请金额（分）
		reason         TEXT NOT NULL DEFAULT '',    -- 申请理由
		attachments    TEXT NOT NULL DEFAULT '[]',  -- 附件 JSON 数组
		applicant_id   INTEGER NOT NULL,            -- 申请人 → users.id
		workflow_state TEXT NOT NULL DEFAULT 'submitted', -- 状态
		current_step   INTEGER NOT NULL DEFAULT 0,  -- 当前审批步骤: 0待初审/1待终审/2完成
		committee_id   INTEGER NOT NULL DEFAULT 0,  -- 村委初审人 → users.id
		committee_note TEXT NOT NULL DEFAULT '',     -- 初审意见
		committee_at   DATETIME,                    -- 初审时间
		secretary_id   INTEGER NOT NULL DEFAULT 0,  -- 村支书终审人 → users.id
		secretary_note TEXT NOT NULL DEFAULT '',     -- 终审意见
		secretary_at   DATETIME,                    -- 终审时间
		created_at     DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (applicant_id) REFERENCES users(id)
	);

	-- ========================================
	-- 工单表（便民服务/报修/投诉/建议）
	-- 工作流: open → assigned → processing → resolved → closed
	-- ========================================
	CREATE TABLE IF NOT EXISTS tickets (
		id             INTEGER PRIMARY KEY AUTOINCREMENT,
		title          TEXT NOT NULL,               -- 标题
		content        TEXT NOT NULL,               -- 详细描述
		category       TEXT NOT NULL DEFAULT 'service', -- 分类: repair/complaint/service/suggestion
		images         TEXT NOT NULL DEFAULT '[]',  -- 图片 JSON 数组
		priority       TEXT NOT NULL DEFAULT 'normal', -- 优先级: low/normal/urgent
		workflow_state TEXT NOT NULL DEFAULT 'open', -- 状态
		submitter_id   INTEGER NOT NULL,            -- 提交人 → users.id
		assignee_id    INTEGER NOT NULL DEFAULT 0,  -- 处理人 → users.id
		resolved_at    DATETIME,                    -- 解决时间
		closed_at      DATETIME,                    -- 关闭时间
		created_at     DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at     DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (submitter_id) REFERENCES users(id)
	);

	-- ========================================
	-- 工单回复表
	-- ========================================
	CREATE TABLE IF NOT EXISTS ticket_comments (
		id         INTEGER PRIMARY KEY AUTOINCREMENT,
		ticket_id  INTEGER NOT NULL,               -- 所属工单 → tickets.id
		user_id    INTEGER NOT NULL,               -- 回复人 → users.id
		content    TEXT NOT NULL,                   -- 回复内容
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (ticket_id) REFERENCES tickets(id) ON DELETE CASCADE,
		FOREIGN KEY (user_id) REFERENCES users(id)
	);

	-- ========================================
	-- 工作流日志表（记录所有审批/状态变更操作）
	-- ========================================
	CREATE TABLE IF NOT EXISTS workflow_logs (
		id            INTEGER PRIMARY KEY AUTOINCREMENT,
		doc_type      TEXT NOT NULL,               -- 文档类型: notice/finance/subsidy/ticket
		doc_id        INTEGER NOT NULL,            -- 文档 ID
		from_state    TEXT NOT NULL,               -- 原状态
		to_state      TEXT NOT NULL,               -- 新状态
		action        TEXT NOT NULL,               -- 操作描述
		operator_id   INTEGER NOT NULL,            -- 操作人 → users.id
		operator_name TEXT NOT NULL DEFAULT '',     -- 操作人姓名（冗余，方便查询）
		note          TEXT NOT NULL DEFAULT '',     -- 操作备注
		created_at    DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	-- ========================================
	-- 消息通知表（站内信）
	-- ========================================
	CREATE TABLE IF NOT EXISTS notifications (
		id         INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id    INTEGER NOT NULL,               -- 接收人 → users.id
		title      TEXT NOT NULL,                  -- 通知标题
		content    TEXT NOT NULL DEFAULT '',        -- 通知内容
		type       TEXT NOT NULL DEFAULT 'system',  -- 类型: system/subsidy/ticket/notice
		ref_type   TEXT NOT NULL DEFAULT '',        -- 关联文档类型
		ref_id     INTEGER NOT NULL DEFAULT 0,     -- 关联文档 ID
		is_read    INTEGER NOT NULL DEFAULT 0,     -- 是否已读: 0/1
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	-- ========================================
	-- 预置角色数据
	-- ========================================
	INSERT OR IGNORE INTO roles (name, label, level) VALUES
		('admin',        '系统管理员',   99),
		('secretary',    '党支部书记',   90),
		('director',     '村委会主任',   85),
		('deputy',       '副书记/副主任', 70),
		('committee',    '两委委员',     60),
		('supervisor',   '监委会委员',   55),
		('accountant',   '村会计',       50),
		('group_leader', '村民小组长',   40),
		('villager',     '村民',         10);

	-- ========================================
	-- 工作流定义表（可配置的状态机）
	-- ========================================
	CREATE TABLE IF NOT EXISTS workflow_defs (
		id          INTEGER PRIMARY KEY AUTOINCREMENT,
		name        TEXT UNIQUE NOT NULL,           -- 唯一标识: notice_review
		label       TEXT NOT NULL,                  -- 中文名: 公告审核流程
		doc_type    TEXT NOT NULL,                  -- 关联文档类型: notice/finance/subsidy/ticket
		states      TEXT NOT NULL DEFAULT '[]',     -- 状态列表 JSON
		transitions TEXT NOT NULL DEFAULT '[]',     -- 转换规则 JSON
		active      INTEGER NOT NULL DEFAULT 1,     -- 是否启用
		created_at  DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	-- ========================================
	-- 报表定义表
	-- ========================================
	CREATE TABLE IF NOT EXISTS report_defs (
		id       INTEGER PRIMARY KEY AUTOINCREMENT,
		name     TEXT UNIQUE NOT NULL,              -- 唯一标识
		label    TEXT NOT NULL,                     -- 中文名
		sql_tpl  TEXT NOT NULL,                     -- SQL 模板
		params   TEXT NOT NULL DEFAULT '[]'         -- 参数定义 JSON
	);

	-- ========================================
	-- 索引
	-- ========================================
	CREATE INDEX IF NOT EXISTS idx_users_account         ON users(account);
	CREATE INDEX IF NOT EXISTS idx_users_group           ON users(group_id);
	CREATE INDEX IF NOT EXISTS idx_users_household       ON users(household_id);
	CREATE INDEX IF NOT EXISTS idx_households_group      ON households(group_id);
	CREATE INDEX IF NOT EXISTS idx_household_members     ON household_members(household_id);
	CREATE INDEX IF NOT EXISTS idx_notices_state         ON notices(workflow_state);
	CREATE INDEX IF NOT EXISTS idx_notices_created       ON notices(created_at DESC);
	CREATE INDEX IF NOT EXISTS idx_finance_state         ON finance_records(workflow_state);
	CREATE INDEX IF NOT EXISTS idx_finance_date          ON finance_records(date DESC);
	CREATE INDEX IF NOT EXISTS idx_subsidies_state       ON subsidies(workflow_state);
	CREATE INDEX IF NOT EXISTS idx_tickets_state         ON tickets(workflow_state);
	CREATE INDEX IF NOT EXISTS idx_ticket_comments_tid   ON ticket_comments(ticket_id);
	CREATE INDEX IF NOT EXISTS idx_workflow_logs         ON workflow_logs(doc_type, doc_id);
	CREATE INDEX IF NOT EXISTS idx_notifications_user    ON notifications(user_id, is_read);
	CREATE INDEX IF NOT EXISTS idx_workflow_defs_doctype ON workflow_defs(doc_type, active);
	`
	if _, err := db.Exec(schema); err != nil {
		return err
	}
	seedWorkflowDefs(db)
	seedReportDefs(db)

	// 更新 schema 版本
	const currentVersion = 1
	if ver < currentVersion {
		db.Exec(`UPDATE schema_version SET version = ?`, currentVersion)
	}
	_ = ver
	return nil
}

func seedWorkflowDefs(db *sql.DB) {
	defs := []struct{ name, label, docType, states, transitions string }{
		{
			"notice_review", "公告审核流程", "notice",
			`[{"name":"draft","label":"草稿","color":"gray"},{"name":"pending_review","label":"待审核","color":"orange"},{"name":"published","label":"已发布","color":"green"},{"name":"rejected","label":"已驳回","color":"red"}]`,
			`[{"action":"submit","label":"提交审核","from":"draft","to":"pending_review","min_role":"villager","notify":false},{"action":"approve","label":"通过","from":"pending_review","to":"published","min_role":"secretary","notify":true},{"action":"reject","label":"驳回","from":"pending_review","to":"rejected","min_role":"secretary","notify":true},{"action":"direct_publish","label":"直接发布","from":"draft","to":"published","min_role":"secretary","notify":false}]`,
		},
		{
			"finance_review", "财务审核流程", "finance",
			`[{"name":"draft","label":"草稿","color":"gray"},{"name":"pending_review","label":"待审核","color":"orange"},{"name":"approved","label":"已审核","color":"green"},{"name":"rejected","label":"已驳回","color":"red"}]`,
			`[{"action":"submit","label":"提交审核","from":"draft","to":"pending_review","min_role":"accountant","notify":false},{"action":"approve","label":"通过","from":"pending_review","to":"approved","min_role":"secretary","notify":true},{"action":"reject","label":"驳回","from":"pending_review","to":"rejected","min_role":"secretary","notify":true},{"action":"direct_approve","label":"直接审核","from":"draft","to":"approved","min_role":"secretary","notify":false}]`,
		},
		{
			"subsidy_approval", "补贴两级审批", "subsidy",
			`[{"name":"submitted","label":"已提交","color":"orange"},{"name":"committee_review","label":"村委初审中","color":"orange"},{"name":"secretary_review","label":"村支书终审中","color":"orange"},{"name":"approved","label":"已通过","color":"green"},{"name":"rejected","label":"已驳回","color":"red"}]`,
			`[{"action":"committee_approve","label":"村委初审通过","from":"submitted","to":"secretary_review","min_role":"committee","notify":true},{"action":"committee_reject","label":"村委初审驳回","from":"submitted","to":"rejected","min_role":"committee","notify":true},{"action":"secretary_approve","label":"村支书终审通过","from":"secretary_review","to":"approved","min_role":"secretary","notify":true},{"action":"secretary_reject","label":"村支书终审驳回","from":"secretary_review","to":"rejected","min_role":"secretary","notify":true}]`,
		},
		{
			"ticket_flow", "工单处理流程", "ticket",
			`[{"name":"open","label":"待处理","color":"orange"},{"name":"assigned","label":"已分配","color":"blue"},{"name":"processing","label":"处理中","color":"blue"},{"name":"resolved","label":"已解决","color":"green"},{"name":"closed","label":"已关闭","color":"gray"}]`,
			`[{"action":"assign","label":"分配","from":"open","to":"assigned","min_role":"group_leader","notify":true},{"action":"start","label":"开始处理","from":"assigned","to":"processing","min_role":"villager","notify":false},{"action":"resolve","label":"解决","from":"processing","to":"resolved","min_role":"villager","notify":true},{"action":"close","label":"关闭","from":"resolved","to":"closed","min_role":"villager","notify":false}]`,
		},
	}
	for _, d := range defs {
		db.Exec(`INSERT OR IGNORE INTO workflow_defs (name, label, doc_type, states, transitions) VALUES (?,?,?,?,?)`,
			d.name, d.label, d.docType, d.states, d.transitions)
	}
}

func seedReportDefs(db *sql.DB) {
	defs := []struct{ name, label, sql, params string }{
		{
			"finance_monthly", "月度收支报表",
			`SELECT strftime('%Y-%m', date) AS 月份, SUM(CASE WHEN type='income' THEN amount ELSE 0 END)/100.0 AS 收入_元, SUM(CASE WHEN type='expense' THEN amount ELSE 0 END)/100.0 AS 支出_元, (SUM(CASE WHEN type='income' THEN amount ELSE 0 END) - SUM(CASE WHEN type='expense' THEN amount ELSE 0 END))/100.0 AS 结余_元 FROM finance_records WHERE workflow_state='approved' AND date LIKE '{{year}}%' GROUP BY strftime('%Y-%m', date) ORDER BY 月份`,
			`[{"name":"year","label":"年份","type":"text","default":"2026"}]`,
		},
		{
			"finance_category", "分类收支统计",
			`SELECT CASE type WHEN 'income' THEN '收入' WHEN 'expense' THEN '支出' END AS 类型, category AS 分类, COUNT(*) AS 笔数, SUM(amount)/100.0 AS 金额_元 FROM finance_records WHERE workflow_state='approved' AND date LIKE '{{year}}%' GROUP BY type, category ORDER BY type, SUM(amount) DESC`,
			`[{"name":"year","label":"年份","type":"text","default":"2026"}]`,
		},
		{
			"subsidy_stats", "补贴统计报表",
			`SELECT CASE s.type WHEN 'farming' THEN '农业补贴' WHEN 'medical' THEN '医疗救助' WHEN 'education' THEN '教育补助' WHEN 'housing' THEN '住房补贴' ELSE '其他' END AS 类型, COUNT(*) AS 申请数, SUM(CASE WHEN s.workflow_state='approved' THEN 1 ELSE 0 END) AS 通过数, SUM(CASE WHEN s.workflow_state='rejected' THEN 1 ELSE 0 END) AS 驳回数, SUM(CASE WHEN s.workflow_state='approved' THEN s.amount ELSE 0 END)/100.0 AS 通过金额_元 FROM subsidies s WHERE s.created_at LIKE '{{year}}%' GROUP BY s.type`,
			`[{"name":"year","label":"年份","type":"text","default":"2026"}]`,
		},
		{
			"ticket_stats", "工单统计报表",
			`SELECT CASE category WHEN 'repair' THEN '报修' WHEN 'complaint' THEN '投诉' WHEN 'service' THEN '便民服务' WHEN 'suggestion' THEN '建议' ELSE category END AS 分类, COUNT(*) AS 总数, SUM(CASE WHEN workflow_state IN ('open','assigned','processing') THEN 1 ELSE 0 END) AS 未完成, SUM(CASE WHEN workflow_state IN ('resolved','closed') THEN 1 ELSE 0 END) AS 已完成, ROUND(SUM(CASE WHEN workflow_state IN ('resolved','closed') THEN 1.0 ELSE 0 END)/COUNT(*)*100, 1) AS 完成率 FROM tickets GROUP BY category`,
			`[]`,
		},
		{
			"group_population", "小组人口统计",
			`SELECT COALESCE(g.name,'未分组') AS 小组, COUNT(u.id) AS 人数, SUM(u.is_party_member) AS 党员数, SUM(u.is_low_income) AS 低保户, SUM(u.is_five_guarantee) AS 五保户, SUM(u.is_disabled) AS 残疾人, COALESCE((SELECT ROUND(SUM(h.farmland_area),1) FROM households h WHERE h.group_id=u.group_id),0) AS 耕地_亩 FROM users u LEFT JOIN groups g ON u.group_id=g.id GROUP BY u.group_id ORDER BY g.name`,
			`[]`,
		},
	}
	for _, d := range defs {
		db.Exec(`INSERT OR IGNORE INTO report_defs (name, label, sql_tpl, params) VALUES (?,?,?,?)`,
			d.name, d.label, d.sql, d.params)
	}
}
