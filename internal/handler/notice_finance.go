package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"village-system/internal/model"
)

// ==================== 公告 ====================

func (h *Handler) ListNotices(w http.ResponseWriter, r *http.Request) {
	page, size := pageParams(r)
	cat := r.URL.Query().Get("category")
	keyword := r.URL.Query().Get("q")
	list, total, _ := h.Notice.List(page, size, cat, keyword, false)
	JSON(w, 200, map[string]any{"data": list, "total": total, "page": page})
}

func (h *Handler) GetNotice(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	n, err := h.Notice.Get(id)
	if err != nil { errJSON(w, 404, "公告不存在"); return }
	logs := h.Workflow.GetLogs("notice", id)
	JSON(w, 200, map[string]any{"notice": n, "logs": logs})
}

// 管理端：列出所有公告（含草稿、待审核）
func (h *Handler) AdminListNotices(w http.ResponseWriter, r *http.Request) {
	page, size := pageParams(r)
	cat := r.URL.Query().Get("category")
	state := r.URL.Query().Get("state")
	list, total, _ := h.Notice.ListAdmin(page, size, cat, state)
	JSON(w, 200, map[string]any{"data": list, "total": total, "page": page})
}

// 创建公告：committee 以上可直接发布，其他人起草待审
func (h *Handler) CreateNotice(w http.ResponseWriter, r *http.Request) {
	var n model.Notice
	if err := json.NewDecoder(r.Body).Decode(&n); err != nil {
		errJSON(w, 400, "请求格式错误"); return
	}
	if n.Title == "" || n.Content == "" {
		errJSON(w, 400, "标题和内容不能为空"); return
	}
	n.AuthorID = getUID(r)
	if n.Attachments == "" { n.Attachments = "[]" }

	// secretary 可直接发布
	if hasRole(r, "secretary") {
		n.WorkflowState = "published"
	} else if hasRole(r, "committee") {
		n.WorkflowState = "pending_review" // 委员起草需支书审核
	} else {
		n.WorkflowState = "draft"
	}

	if err := h.Notice.Create(&n); err != nil {
		errJSON(w, 500, err.Error()); return
	}
	u, _ := h.User.GetByID(getUID(r))
	name := ""
	if u != nil { name = u.Name }
	h.Workflow.Log("notice", n.ID, "", n.WorkflowState, "创建", getUID(r), name, "")
	JSON(w, 201, n)
}

func (h *Handler) UpdateNotice(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	var req struct {
		Title       *string `json:"title"`
		Content     *string `json:"content"`
		Category    *string `json:"category"`
		Pinned      *bool   `json:"pinned"`
		Attachments *string `json:"attachments"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	// 只更新 pinned
	if req.Pinned != nil && req.Title == nil {
		h.Notice.DB.Exec(`UPDATE notices SET pinned=?, updated_at=CURRENT_TIMESTAMP WHERE id=?`, *req.Pinned, id)
		JSON(w, 200, map[string]string{"ok": "updated"})
		return
	}
	var n model.Notice
	n.ID = id
	if req.Title != nil { n.Title = *req.Title }
	if req.Content != nil { n.Content = *req.Content }
	if req.Category != nil { n.Category = *req.Category }
	if req.Pinned != nil { n.Pinned = *req.Pinned }
	att := "[]"
	if req.Attachments != nil { att = *req.Attachments }
	n.Attachments = att
	h.Notice.Update(&n)
	JSON(w, 200, map[string]string{"ok": "updated"})
}

func (h *Handler) DeleteNotice(w http.ResponseWriter, r *http.Request) {
	if !hasRole(r, "deputy") {
		errJSON(w, 403, "需要副书记/副主任以上权限"); return
	}
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	h.Notice.Delete(id)
	JSON(w, 200, map[string]string{"ok": "deleted"})
}

// 审核公告（副书记/副主任以上）
func (h *Handler) ReviewNotice(w http.ResponseWriter, r *http.Request) {
	if !hasRole(r, "deputy") {
		errJSON(w, 403, "需要副书记/副主任以上权限"); return
	}
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	var req struct {
		Action string `json:"action"` // approve / reject
		Note   string `json:"note"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	old, _ := h.Notice.Get(id)
	if old == nil { errJSON(w, 404, "公告不存在"); return }

	var newState string
	if req.Action == "approve" {
		newState = "published"
	} else {
		newState = "rejected"
	}
	h.Notice.UpdateState(id, newState, getUID(r), req.Note)
	u, _ := h.User.GetByID(getUID(r))
	name := ""
	if u != nil { name = u.Name }
	h.Workflow.Log("notice", id, old.WorkflowState, newState, req.Action, getUID(r), name, req.Note)
	JSON(w, 200, map[string]string{"ok": newState})
}

// ==================== 财务 ====================

func (h *Handler) ListFinance(w http.ResponseWriter, r *http.Request) {
	page, size := pageParams(r)
	year := r.URL.Query().Get("year")
	typ := r.URL.Query().Get("type")
	list, total, _ := h.Finance.List(page, size, year, typ, false)
	JSON(w, 200, map[string]any{"data": list, "total": total, "page": page})
}

func (h *Handler) FinanceSummary(w http.ResponseWriter, r *http.Request) {
	year := r.URL.Query().Get("year")
	sum, _ := h.Finance.Summary(year)
	JSON(w, 200, sum)
}

func (h *Handler) AdminListFinance(w http.ResponseWriter, r *http.Request) {
	page, size := pageParams(r)
	year := r.URL.Query().Get("year")
	typ := r.URL.Query().Get("type")
	list, total, _ := h.Finance.List(page, size, year, typ, true)
	JSON(w, 200, map[string]any{"data": list, "total": total, "page": page})
}

func (h *Handler) CreateFinance(w http.ResponseWriter, r *http.Request) {
	var rec model.FinanceRecord
	if err := json.NewDecoder(r.Body).Decode(&rec); err != nil {
		errJSON(w, 400, "请求格式错误"); return
	}
	if rec.Type != "income" && rec.Type != "expense" {
		errJSON(w, 400, "type 必须是 income 或 expense"); return
	}
	if rec.Amount <= 0 { errJSON(w, 400, "金额必须大于0"); return }
	if rec.Date == "" { errJSON(w, 400, "日期不能为空"); return }
	rec.AuthorID = getUID(r)

	if hasRole(r, "secretary") {
		rec.WorkflowState = "approved"
	} else {
		rec.WorkflowState = "pending_review"
	}
	h.Finance.Create(&rec)
	u, _ := h.User.GetByID(getUID(r))
	name := ""
	if u != nil { name = u.Name }
	h.Workflow.Log("finance", rec.ID, "", rec.WorkflowState, "创建", getUID(r), name, "")
	JSON(w, 201, rec)
}

// 审核财务（监委会以上，且不能审核自己录入的）
func (h *Handler) ReviewFinance(w http.ResponseWriter, r *http.Request) {
	if !hasRole(r, "supervisor") {
		errJSON(w, 403, "需要监委会以上权限"); return
	}
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	rec, _ := h.Finance.Get(id)
	if rec != nil && rec.AuthorID == getUID(r) {
		errJSON(w, 403, "不能审核自己录入的财务记录"); return
	}
	var req struct {
		Action string `json:"action"`
		Note   string `json:"note"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	newState := "approved"
	if req.Action != "approve" { newState = "rejected" }
	h.Finance.UpdateState(id, newState, getUID(r), req.Note)
	u, _ := h.User.GetByID(getUID(r))
	name := ""
	if u != nil { name = u.Name }
	h.Workflow.Log("finance", id, "pending_review", newState, req.Action, getUID(r), name, req.Note)
	JSON(w, 200, map[string]string{"ok": newState})
}

// 删除财务（副书记/副主任以上）
func (h *Handler) DeleteFinance(w http.ResponseWriter, r *http.Request) {
	if !hasRole(r, "deputy") {
		errJSON(w, 403, "需要副书记/副主任以上权限"); return
	}
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	h.Finance.Delete(id)
	JSON(w, 200, map[string]string{"ok": "deleted"})
}
