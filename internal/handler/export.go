package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/xuri/excelize/v2"
)

func (h *Handler) ExportUsers(w http.ResponseWriter, r *http.Request) {
	groupID, _ := strconv.ParseInt(r.URL.Query().Get("group_id"), 10, 64)
	list, _, _ := h.User.List(1, 10000, map[string]string{}, groupID)

	f := excelize.NewFile()
	s := "用户列表"
	f.SetSheetName("Sheet1", s)
	headers := []string{"姓名", "手机号", "性别", "出生日期", "民族", "文化程度", "婚姻状况", "身份证号", "地址", "所属小组", "角色",
		"党员", "低保户", "五保户", "残疾人", "军属", "紧急联系人", "紧急电话", "备注"}
	writeRow(f, s, 1, headers)
	for i, u := range list {
		gender := ""
		if u.Gender == "male" { gender = "男" } else if u.Gender == "female" { gender = "女" }
		marital := map[string]string{"unmarried":"未婚","married":"已婚","divorced":"离异","widowed":"丧偶"}[u.MaritalStatus]
		writeRow(f, s, i+2, []string{
			u.Name, u.Phone, gender, u.BirthDate, u.Ethnicity, u.Education, marital, u.IDCard, u.Address, u.GroupName, u.RoleLabel,
			boolStr(u.IsPartyMember), boolStr(u.IsLowIncome), boolStr(u.IsFiveGuarantee), boolStr(u.IsDisabled), boolStr(u.IsMilitary),
			u.EmergencyContact, u.EmergencyPhone, u.Remark,
		})
	}
	writeXlsx(w, f, "users_"+time.Now().Format("20060102")+".xlsx")
}

func (h *Handler) ExportFinance(w http.ResponseWriter, r *http.Request) {
	year := r.URL.Query().Get("year")
	if year == "" { year = time.Now().Format("2006") }
	list, _, _ := h.Finance.List(1, 10000, year, "", true)

	f := excelize.NewFile()
	s := year + "年财务"
	f.SetSheetName("Sheet1", s)
	writeRow(f, s, 1, []string{"日期", "类型", "金额(元)", "分类", "备注", "录入人", "状态"})
	stateMap := map[string]string{"draft": "草稿", "pending_review": "待审核", "approved": "已审核", "rejected": "已驳回"}
	for i, r := range list {
		typ := "收入"
		if r.Type == "expense" { typ = "支出" }
		writeRow(f, s, i+2, []string{
			r.Date, typ, fmt.Sprintf("%.2f", float64(r.Amount)/100), r.Category, r.Remark, r.Author, stateMap[r.WorkflowState],
		})
	}
	writeXlsx(w, f, "finance_"+year+".xlsx")
}

func (h *Handler) ExportSubsidies(w http.ResponseWriter, r *http.Request) {
	list, _, _ := h.Subsidy.List(1, 10000, "", 0)

	f := excelize.NewFile()
	s := "补贴台账"
	f.SetSheetName("Sheet1", s)
	typeMap := map[string]string{"farming": "农业补贴", "medical": "医疗救助", "education": "教育补助", "housing": "住房补贴", "other": "其他"}
	stateMap := map[string]string{"submitted": "已提交", "committee_review": "村委初审中", "secretary_review": "村支书终审中", "approved": "已通过", "rejected": "已驳回"}
	writeRow(f, s, 1, []string{"补贴名称", "类型", "金额(元)", "申请人", "申请理由", "状态", "村委初审人", "初审意见", "村支书终审人", "终审意见", "申请时间"})
	for i, sub := range list {
		writeRow(f, s, i+2, []string{
			sub.Title, typeMap[sub.Type], fmt.Sprintf("%.2f", float64(sub.Amount)/100), sub.Applicant, sub.Reason,
			stateMap[sub.WorkflowState], sub.CommitteeName, sub.CommitteeNote, sub.SecretaryName, sub.SecretaryNote,
			sub.CreatedAt.Format("2006-01-02"),
		})
	}
	writeXlsx(w, f, "subsidies_"+time.Now().Format("20060102")+".xlsx")
}

func writeRow(f *excelize.File, sheet string, row int, vals []string) {
	for j, v := range vals {
		cell, _ := excelize.CoordinatesToCellName(j+1, row)
		f.SetCellValue(sheet, cell, v)
	}
}

func writeXlsx(w http.ResponseWriter, f *excelize.File, filename string) {
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", "attachment; filename="+filename)
	f.Write(w)
}

func boolStr(b bool) string {
	if b { return "是" }
	return ""
}
