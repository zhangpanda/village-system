package model

// ========== 工作流定义（可配置的状态机） ==========

// WorkflowDef 工作流定义
type WorkflowDef struct {
	ID          int64              `json:"id"`
	Name        string             `json:"name"`         // 唯一标识: notice_review, subsidy_approval
	Label       string             `json:"label"`        // 中文名: 公告审核流程
	DocType     string             `json:"doc_type"`     // 关联文档类型: notice, finance, subsidy, ticket
	States      []WorkflowStateDef `json:"states"`       // 状态列表
	Transitions []TransitionDef    `json:"transitions"`  // 转换规则
	Active      bool               `json:"active"`       // 是否启用
}

// WorkflowStateDef 状态定义
type WorkflowStateDef struct {
	Name  string `json:"name"`  // 状态标识: draft, pending_review
	Label string `json:"label"` // 中文名: 草稿, 待审核
	Color string `json:"color"` // 前端颜色: gray, orange, green, red
}

// TransitionDef 转换规则
type TransitionDef struct {
	Action   string `json:"action"`    // 操作标识: submit, approve, reject
	Label    string `json:"label"`     // 中文名: 提交, 通过, 驳回
	From     string `json:"from"`      // 源状态
	To       string `json:"to"`        // 目标状态
	MinRole  string `json:"min_role"`  // 最低角色要求
	Notify   bool   `json:"notify"`    // 是否发通知
}

// ========== 报表定义 ==========

type ReportDef struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`   // 唯一标识
	Label  string `json:"label"`  // 中文名
	SQL    string `json:"sql"`    // SQL 模板（支持 {{year}} 等占位符）
	Params string `json:"params"` // 参数定义 JSON: [{"name":"year","label":"年份","type":"text","default":"2026"}]
}

type ReportResult struct {
	Columns []string        `json:"columns"`
	Rows    [][]interface{} `json:"rows"`
}

// ========== 数据导入 ==========

type ImportResult struct {
	Total   int      `json:"total"`
	Success int      `json:"success"`
	Failed  int      `json:"failed"`
	Errors  []string `json:"errors,omitempty"`
}
