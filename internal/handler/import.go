package handler

import (
	"fmt"
	"net/http"
	"strings"

	"village-system/internal/middleware"
	"village-system/internal/model"

	"github.com/xuri/excelize/v2"
)

// ImportUsers 从 Excel 批量导入村民
// POST /api/admin/import/users (multipart/form-data, field: file)
func (h *Handler) ImportUsers(w http.ResponseWriter, r *http.Request) {
	file, _, err := r.FormFile("file")
	if err != nil {
		errJSON(w, 400, "请上传 Excel 文件"); return
	}
	defer file.Close()

	f, err := excelize.OpenReader(file)
	if err != nil {
		errJSON(w, 400, "无法解析 Excel 文件"); return
	}
	defer f.Close()

	sheet := f.GetSheetName(0)
	rows, err := f.GetRows(sheet)
	if err != nil || len(rows) < 2 {
		errJSON(w, 400, "文件为空或格式不正确（第一行应为表头）"); return
	}

	// 解析表头，建立列名→索引映射
	colMap := map[string]int{}
	for i, h := range rows[0] {
		colMap[strings.TrimSpace(h)] = i
	}

	// 必须有姓名列
	if _, ok := colMap["姓名"]; !ok {
		errJSON(w, 400, "缺少「姓名」列"); return
	}

	result := model.ImportResult{Total: len(rows) - 1}
	genderMap := map[string]string{"男": "male", "女": "female"}

	for i, row := range rows[1:] {
		lineNo := i + 2
		u := model.User{Role: "villager", Position: "villager", Active: true}

		u.Name = getCol(row, colMap, "姓名")
		if u.Name == "" {
			result.Failed++
			result.Errors = append(result.Errors, fmt.Sprintf("第%d行：姓名为空", lineNo))
			continue
		}

		u.Phone = getCol(row, colMap, "手机号")
		u.Gender = genderMap[getCol(row, colMap, "性别")]
		u.BirthDate = getCol(row, colMap, "出生日期")
		u.IDCard = getCol(row, colMap, "身份证号")
		u.Address = getCol(row, colMap, "地址")
		u.Ethnicity = getCol(row, colMap, "民族")
		if u.Ethnicity == "" { u.Ethnicity = "汉族" }
		u.Education = getCol(row, colMap, "文化程度")
		maritalMap := map[string]string{"未婚": "unmarried", "已婚": "married", "离异": "divorced", "丧偶": "widowed"}
		u.MaritalStatus = maritalMap[getCol(row, colMap, "婚姻状况")]
		u.WechatID = getCol(row, colMap, "微信号")
		u.EmergencyContact = getCol(row, colMap, "紧急联系人")
		u.EmergencyPhone = getCol(row, colMap, "紧急电话")
		u.Remark = getCol(row, colMap, "备注")

		u.IsPartyMember = getCol(row, colMap, "党员") == "是"
		u.IsLowIncome = getCol(row, colMap, "低保户") == "是"
		u.IsFiveGuarantee = getCol(row, colMap, "五保户") == "是"
		u.IsDisabled = getCol(row, colMap, "残疾人") == "是"
		u.IsMilitary = getCol(row, colMap, "军属") == "是"

		// 账号默认用手机号，没手机号用姓名
		u.Account = u.Phone
		if u.Account == "" { u.Account = u.Name }
		u.PasswordHash = middleware.HashPassword("123456") // 默认密码

		// 检查手机号是否已存在
		if u.Phone != "" {
			if existing, _ := h.User.GetByPhoneOnly(u.Phone); existing != nil {
				result.Failed++
				result.Errors = append(result.Errors, fmt.Sprintf("第%d行：手机号 %s 已存在", lineNo, u.Phone))
				continue
			}
		}

		if err := h.User.Create(&u); err != nil {
			result.Failed++
			result.Errors = append(result.Errors, fmt.Sprintf("第%d行：%s", lineNo, err.Error()))
			continue
		}
		result.Success++
	}

	JSON(w, 200, result)
}

// ImportTemplate 下载导入模板
func (h *Handler) ImportTemplate(w http.ResponseWriter, r *http.Request) {
	f := excelize.NewFile()
	s := "村民导入模板"
	f.SetSheetName("Sheet1", s)
	headers := []string{"姓名", "手机号", "性别", "出生日期", "民族", "文化程度", "婚姻状况", "身份证号", "地址",
		"党员", "低保户", "五保户", "残疾人", "军属",
		"微信号", "紧急联系人", "紧急电话", "备注"}
	writeRow(f, s, 1, headers)
	// 示例行
	writeRow(f, s, 2, []string{"张三", "13800000001", "男", "1980-01-01", "汉族", "初中", "已婚", "", "幸福路1号",
		"是", "", "", "", "", "", "李四", "13900000001", ""})

	writeXlsx(w, f, "村民导入模板.xlsx")
}

func getCol(row []string, colMap map[string]int, name string) string {
	idx, ok := colMap[name]
	if !ok || idx >= len(row) { return "" }
	return strings.TrimSpace(row[idx])
}
