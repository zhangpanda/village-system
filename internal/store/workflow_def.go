package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"village-system/internal/model"
)

type WorkflowDefStore struct{ DB *sql.DB }

// GetByDocType 获取文档类型对应的活跃工作流
func (s *WorkflowDefStore) GetByDocType(docType string) *model.WorkflowDef {
	var id int64
	var name, label, statesJSON, transJSON string
	var active bool
	err := s.DB.QueryRow(
		`SELECT id, name, label, states, transitions, active FROM workflow_defs WHERE doc_type=? AND active=1 LIMIT 1`, docType,
	).Scan(&id, &name, &label, &statesJSON, &transJSON, &active)
	if err != nil {
		return nil
	}
	def := &model.WorkflowDef{ID: id, Name: name, Label: label, DocType: docType, Active: active}
	json.Unmarshal([]byte(statesJSON), &def.States)
	json.Unmarshal([]byte(transJSON), &def.Transitions)
	return def
}

// FindTransition 查找可用的转换
func (s *WorkflowDefStore) FindTransition(docType, fromState, action, userRole string) *model.TransitionDef {
	def := s.GetByDocType(docType)
	if def == nil {
		return nil
	}
	for _, t := range def.Transitions {
		if t.From == fromState && t.Action == action {
			if t.MinRole == "" || hasRoleLevel(userRole, t.MinRole) {
				return &t
			}
		}
	}
	return nil
}

// ListAll 列出所有工作流定义
func (s *WorkflowDefStore) ListAll() []model.WorkflowDef {
	rows, err := s.DB.Query(`SELECT id, name, label, doc_type, states, transitions, active FROM workflow_defs ORDER BY id`)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var list []model.WorkflowDef
	for rows.Next() {
		var d model.WorkflowDef
		var statesJSON, transJSON string
		rows.Scan(&d.ID, &d.Name, &d.Label, &d.DocType, &statesJSON, &transJSON, &d.Active)
		json.Unmarshal([]byte(statesJSON), &d.States)
		json.Unmarshal([]byte(transJSON), &d.Transitions)
		list = append(list, d)
	}
	return list
}

// Upsert 创建或更新工作流定义
func (s *WorkflowDefStore) Upsert(def *model.WorkflowDef) error {
	statesJSON, _ := json.Marshal(def.States)
	transJSON, _ := json.Marshal(def.Transitions)
	if def.ID > 0 {
		_, err := s.DB.Exec(
			`UPDATE workflow_defs SET name=?, label=?, doc_type=?, states=?, transitions=?, active=? WHERE id=?`,
			def.Name, def.Label, def.DocType, statesJSON, transJSON, def.Active, def.ID,
		)
		return err
	}
	res, err := s.DB.Exec(
		`INSERT INTO workflow_defs (name, label, doc_type, states, transitions, active) VALUES (?,?,?,?,?,?)`,
		def.Name, def.Label, def.DocType, statesJSON, transJSON, def.Active,
	)
	if err != nil {
		return err
	}
	def.ID, _ = res.LastInsertId()
	return nil
}

func (s *WorkflowDefStore) Delete(id int64) error {
	_, err := s.DB.Exec(`DELETE FROM workflow_defs WHERE id=?`, id)
	return err
}

// ========== 报表 ==========

type ReportStore struct{ DB *sql.DB }

func (s *ReportStore) List() []model.ReportDef {
	rows, err := s.DB.Query(`SELECT id, name, label, sql_tpl, params FROM report_defs ORDER BY id`)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var list []model.ReportDef
	for rows.Next() {
		var r model.ReportDef
		rows.Scan(&r.ID, &r.Name, &r.Label, &r.SQL, &r.Params)
		list = append(list, r)
	}
	return list
}

func (s *ReportStore) Get(name string) (*model.ReportDef, error) {
	r := &model.ReportDef{}
	err := s.DB.QueryRow(`SELECT id, name, label, sql_tpl, params FROM report_defs WHERE name=?`, name).
		Scan(&r.ID, &r.Name, &r.Label, &r.SQL, &r.Params)
	return r, err
}

func (s *ReportStore) Upsert(r *model.ReportDef) error {
	if r.ID > 0 {
		_, err := s.DB.Exec(`UPDATE report_defs SET name=?, label=?, sql_tpl=?, params=? WHERE id=?`,
			r.Name, r.Label, r.SQL, r.Params, r.ID)
		return err
	}
	res, err := s.DB.Exec(`INSERT INTO report_defs (name, label, sql_tpl, params) VALUES (?,?,?,?)`,
		r.Name, r.Label, r.SQL, r.Params)
	if err != nil {
		return err
	}
	r.ID, _ = res.LastInsertId()
	return nil
}

func (s *ReportStore) Delete(id int64) error {
	_, err := s.DB.Exec(`DELETE FROM report_defs WHERE id=?`, id)
	return err
}

// Execute 执行报表查询（只允许 SELECT）
func (s *ReportStore) Execute(sqlTpl string, params map[string]string) (*model.ReportResult, error) {
	// 安全检查：只允许 SELECT
	upper := strings.ToUpper(strings.TrimSpace(sqlTpl))
	if !strings.HasPrefix(upper, "SELECT") {
		return nil, fmt.Errorf("只允许 SELECT 查询")
	}
	for _, kw := range []string{" INSERT ", " UPDATE ", " DELETE ", " DROP ", " ALTER ", " CREATE ", " EXEC "} {
		if strings.Contains(" "+upper+" ", kw) {
			return nil, fmt.Errorf("不允许包含%s", strings.TrimSpace(kw))
		}
	}

	// 参数化查询防注入：{{key}} → ?
	query := sqlTpl
	var args []any
	for k, v := range params {
		ph := "{{" + k + "}}"
		for strings.Contains(query, ph) {
			query = strings.Replace(query, ph, "?", 1)
			args = append(args, v)
		}
	}

	rows, err := s.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cols, _ := rows.Columns()
	result := &model.ReportResult{Columns: cols}
	for rows.Next() {
		vals := make([]interface{}, len(cols))
		ptrs := make([]interface{}, len(cols))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		rows.Scan(ptrs...)
		// 转换 []byte 为 string
		row := make([]interface{}, len(cols))
		for i, v := range vals {
			if b, ok := v.([]byte); ok {
				row[i] = string(b)
			} else {
				row[i] = v
			}
		}
		result.Rows = append(result.Rows, row)
	}
	return result, nil
}

// hasRoleLevel 简单角色等级判断（复用 model 的逻辑）
func hasRoleLevel(userRoles, minRole string) bool {
	roleLevel := map[string]int{
		"admin": 99, "secretary": 90, "resident_official": 88, "director": 85, "deputy": 70,
		"supervisor": 65, "committee": 60, "accountant": 50,
		"group_leader": 40, "grid_worker": 35, "villager": 10,
	}
	minLevel := roleLevel[minRole]
	for _, r := range strings.Split(userRoles, ",") {
		r = strings.TrimSpace(r)
		if roleLevel[r] >= minLevel {
			return true
		}
	}
	return false
}
