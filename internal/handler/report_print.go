package handler

import (
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"strconv"
	"strings"
	"time"

	"village-system/internal/model"
)

// ==================== 报表 ====================

func (h *Handler) ListReports(w http.ResponseWriter, r *http.Request) {
	JSON(w, 200, h.Report.List())
}

func (h *Handler) RunReport(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	rpt, err := h.Report.Get(name)
	if err != nil {
		errJSON(w, 404, "报表不存在"); return
	}

	// 从 query 参数获取报表参数
	params := map[string]string{}
	var paramDefs []struct {
		Name    string `json:"name"`
		Default string `json:"default"`
	}
	json.Unmarshal([]byte(rpt.Params), &paramDefs)
	for _, p := range paramDefs {
		v := r.URL.Query().Get(p.Name)
		if v == "" { v = p.Default }
		// 安全过滤：只允许字母数字和基本符号
		v = sanitizeParam(v)
		params[p.Name] = v
	}

	result, err := h.Report.Execute(rpt.SQL, params)
	if err != nil {
		errJSON(w, 500, "查询失败: "+err.Error()); return
	}
	JSON(w, 200, map[string]any{"report": rpt, "result": result})
}

func (h *Handler) SaveReport(w http.ResponseWriter, r *http.Request) {
	var rpt model.ReportDef
	if err := json.NewDecoder(r.Body).Decode(&rpt); err != nil {
		errJSON(w, 400, "请求格式错误"); return
	}
	if rpt.Name == "" || rpt.SQL == "" {
		errJSON(w, 400, "name 和 sql 不能为空"); return
	}
	if err := h.Report.Upsert(&rpt); err != nil {
		errJSON(w, 500, err.Error()); return
	}
	JSON(w, 200, rpt)
}

func (h *Handler) DeleteReport(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	h.Report.Delete(id)
	JSON(w, 200, map[string]string{"ok": "deleted"})
}

// ==================== 打印/PDF ====================

// PrintSubsidy 补贴审批单打印页
func (h *Handler) PrintSubsidy(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	sub, err := h.Subsidy.Get(id)
	if err != nil {
		errJSON(w, 404, "补贴申请不存在"); return
	}
	logs := h.Workflow.GetLogs("subsidy", id)

	typeMap := map[string]string{"farming": "农业补贴", "medical": "医疗救助", "education": "教育补助", "housing": "住房补贴", "other": "其他"}
	stateMap := map[string]string{"submitted": "已提交", "committee_review": "村委初审中", "secretary_review": "村支书终审中", "approved": "已通过", "rejected": "已驳回"}

	esc := html.EscapeString
	html_ := printHeader(esc(h.VillageName) + " — 补贴审批单")
	html_ += `<h1>` + esc(h.VillageName) + ` 补贴审批单</h1>`
	html_ += `<table><tr><td width="50%">补贴名称：` + esc(sub.Title) + `</td><td>类型：` + typeMap[sub.Type] + `</td></tr>`
	html_ += `<tr><td>申请人：` + esc(sub.Applicant) + `</td><td>金额：` + fmt.Sprintf("%.2f 元", float64(sub.Amount)/100) + `</td></tr>`
	html_ += `<tr><td colspan="2">申请理由：` + esc(sub.Reason) + `</td></tr>`
	html_ += `<tr><td>状态：` + stateMap[sub.WorkflowState] + `</td><td>申请时间：` + sub.CreatedAt.Format("2006-01-02") + `</td></tr>`
	if sub.CommitteeName != "" {
		html_ += `<tr><td>村委初审：` + esc(sub.CommitteeName) + `</td><td>意见：` + esc(sub.CommitteeNote) + `</td></tr>`
	}
	if sub.SecretaryName != "" {
		html_ += `<tr><td>村支书终审：` + esc(sub.SecretaryName) + `</td><td>意见：` + esc(sub.SecretaryNote) + `</td></tr>`
	}
	html_ += `</table>`

	if len(logs) > 0 {
		html_ += `<h3>审批记录</h3><table><tr><th>时间</th><th>操作人</th><th>操作</th><th>备注</th></tr>`
		for _, l := range logs {
			html_ += `<tr><td>` + l.CreatedAt.Format("2006-01-02 15:04") + `</td><td>` + esc(l.OperatorName) + `</td><td>` + esc(l.Action) + `</td><td>` + esc(l.Note) + `</td></tr>`
		}
		html_ += `</table>`
	}
	html_ += printFooter()
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html_))
}

// PrintFinance 财务报表打印页
func (h *Handler) PrintFinance(w http.ResponseWriter, r *http.Request) {
	year := r.URL.Query().Get("year")
	if year == "" { year = time.Now().Format("2006") }
	list, _, _ := h.Finance.List(1, 10000, year, "", true)
	sum, _ := h.Finance.Summary(year)

	esc2 := html.EscapeString
	html2 := printHeader(esc2(h.VillageName) + " — " + year + "年财务报表")
	html2 += `<h1>` + esc2(h.VillageName) + ` ` + year + `年财务报表</h1>`
	html2 += `<div class="summary">收入：` + fmt.Sprintf("%.2f", float64(sum.TotalIncome)/100) + ` 元 | 支出：` + fmt.Sprintf("%.2f", float64(sum.TotalExpense)/100) + ` 元 | 结余：` + fmt.Sprintf("%.2f", float64(sum.Balance)/100) + ` 元</div>`
	html2 += `<table><tr><th>日期</th><th>类型</th><th>金额(元)</th><th>分类</th><th>备注</th><th>录入人</th></tr>`
	for _, rec := range list {
		if rec.WorkflowState != "approved" { continue }
		typ := "收入"
		if rec.Type == "expense" { typ = "支出" }
		html2 += `<tr><td>` + rec.Date + `</td><td>` + typ + `</td><td>` + fmt.Sprintf("%.2f", float64(rec.Amount)/100) + `</td><td>` + esc2(rec.Category) + `</td><td>` + esc2(rec.Remark) + `</td><td>` + esc2(rec.Author) + `</td></tr>`
	}
	html2 += `</table>`
	html2 += printFooter()
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html2))
}

// PrintReport 通用报表打印
func (h *Handler) PrintReport(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	rpt, err := h.Report.Get(name)
	if err != nil {
		errJSON(w, 404, "报表不存在"); return
	}

	params := map[string]string{}
	var paramDefs []struct {
		Name    string `json:"name"`
		Default string `json:"default"`
	}
	json.Unmarshal([]byte(rpt.Params), &paramDefs)
	for _, p := range paramDefs {
		v := r.URL.Query().Get(p.Name)
		if v == "" { v = p.Default }
		params[p.Name] = sanitizeParam(v)
	}

	result, err := h.Report.Execute(rpt.SQL, params)
	if err != nil {
		errJSON(w, 500, "查询失败"); return
	}

	esc3 := html.EscapeString
	html3 := printHeader(esc3(h.VillageName) + " — " + esc3(rpt.Label))
	html3 += `<h1>` + esc3(h.VillageName) + ` ` + esc3(rpt.Label) + `</h1>`
	html3 += `<p>生成时间：` + time.Now().Format("2006-01-02 15:04") + `</p>`
	html3 += `<table><tr>`
	for _, col := range result.Columns {
		html3 += `<th>` + esc3(col) + `</th>`
	}
	html3 += `</tr>`
	for _, row := range result.Rows {
		html3 += `<tr>`
		for _, v := range row {
			html3 += `<td>` + esc3(fmt.Sprintf("%v", v)) + `</td>`
		}
		html3 += `</tr>`
	}
	html3 += `</table>`
	html3 += printFooter()
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html3))
}

func printHeader(title string) string {
	return `<!DOCTYPE html><html><head><meta charset="utf-8"><title>` + title + `</title>
<style>
body{font-family:SimSun,serif;max-width:800px;margin:0 auto;padding:20px;font-size:13px}
h1{text-align:center;font-size:20px;margin-bottom:5px}
.summary{text-align:center;margin:10px 0;font-size:16px;font-weight:bold}
table{width:100%;border-collapse:collapse;margin:15px 0}
th,td{border:1px solid #333;padding:4px 6px;text-align:left;font-size:12px}
th{background:#f0f0f0}
.footer{margin-top:30px;text-align:right;font-size:12px;color:#666}
@media print{body{margin:0;padding:10px}.no-print{display:none}}
</style></head><body>
<div class="no-print" style="text-align:right;margin-bottom:10px"><button onclick="window.print()">🖨️ 打印</button></div>`
}

func printFooter() string {
	return `<div class="footer">打印时间：` + time.Now().Format("2006-01-02 15:04:05") + `</div></body></html>`
}

func sanitizeParam(v string) string {
	v = strings.ReplaceAll(v, "'", "")
	v = strings.ReplaceAll(v, ";", "")
	v = strings.ReplaceAll(v, "--", "")
	v = strings.ReplaceAll(v, "/*", "")
	return v
}

// PrintRoster 花名册打印
func (h *Handler) PrintRoster(w http.ResponseWriter, r *http.Request) {
	esc := html.EscapeString
	groupID, _ := strconv.ParseInt(r.URL.Query().Get("group_id"), 10, 64)
	filters := map[string]string{}
	for _, k := range []string{"role", "q", "gender", "education", "marital_status", "tag"} {
		if v := r.URL.Query().Get(k); v != "" { filters[k] = v }
	}
	list, total, _ := h.User.List(1, 10000, filters, groupID)

	// 标题
	title := h.VillageName
	parts := []string{}
	if v := filters["role"]; v != "" { parts = append(parts, model.RoleLabel[v]) }
	if v := filters["tag"]; v != "" {
		tagMap := map[string]string{"party":"党员","low_income":"低保户","five_guarantee":"五保户","disabled":"残疾人","military":"军属/退役"}
		parts = append(parts, tagMap[v])
	}
	if v := filters["gender"]; v != "" {
		if v == "male" { parts = append(parts, "男") } else { parts = append(parts, "女") }
	}
	if len(parts) > 0 { title += " " + strings.Join(parts, "·") }
	title += " 花名册"

	genderMap := map[string]string{"male":"男","female":"女"}

	h_ := printHeader(esc(title))
	h_ += `<h1>` + esc(title) + `</h1>`
	h_ += `<p style="text-align:center;color:#666">共 ` + fmt.Sprintf("%d", total) + ` 人 · ` + time.Now().Format("2006-01-02") + `</p>`
	h_ += `<table><tr><th>序号</th><th>姓名</th><th>年龄</th><th>性别</th><th>身份证号</th><th>电话</th><th>小组</th><th>学历</th><th>身份</th></tr>`
	for i, u := range list {
		age := ""
		if len(u.BirthDate) >= 4 {
			var y int
			fmt.Sscanf(u.BirthDate[:4], "%d", &y)
			if y > 0 { age = fmt.Sprintf("%d", time.Now().Year()-y) }
		}
		tags := []string{}
		if u.IsPartyMember { tags = append(tags, "党员") }
		if u.IsLowIncome { tags = append(tags, "低保") }
		if u.IsFiveGuarantee { tags = append(tags, "五保") }
		if u.IsDisabled { tags = append(tags, "残疾") }
		if u.IsMilitary { tags = append(tags, "军属") }
		h_ += `<tr><td>` + fmt.Sprintf("%d", i+1) + `</td><td>` + esc(u.Name) + `</td><td>` + age + `</td><td>` + genderMap[u.Gender] + `</td><td>` + esc(u.IDCard) + `</td><td>` + u.Phone + `</td><td>` + esc(u.GroupName) + `</td><td>` + esc(u.Education) + `</td><td>` + strings.Join(tags, "/") + `</td></tr>`
	}
	h_ += `</table>`
	h_ += printFooter()
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(h_))
}
