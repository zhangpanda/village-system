package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"village-system/internal/model"
)

// ==================== 补贴 ====================

func (h *Handler) ListSubsidies(w http.ResponseWriter, r *http.Request) {
	page, size := pageParams(r)
	state := r.URL.Query().Get("state")
	var applicantID int64
	// 普通村民只能看自己的
	if !hasRole(r, "committee") {
		applicantID = getUID(r)
	}
	list, total, _ := h.Subsidy.List(page, size, state, applicantID)
	JSON(w, 200, map[string]any{"data": list, "total": total, "page": page})
}

func (h *Handler) GetSubsidy(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	sub, err := h.Subsidy.Get(id)
	if err != nil { errJSON(w, 404, "补贴申请不存在"); return }
	logs := h.Workflow.GetLogs("subsidy", id)
	JSON(w, 200, map[string]any{"subsidy": sub, "logs": logs})
}

func (h *Handler) CreateSubsidy(w http.ResponseWriter, r *http.Request) {
	var sub model.Subsidy
	if err := json.NewDecoder(r.Body).Decode(&sub); err != nil {
		errJSON(w, 400, "请求格式错误"); return
	}
	if sub.Title == "" { errJSON(w, 400, "补贴名称不能为空"); return }
	if sub.Amount <= 0 { errJSON(w, 400, "申请金额必须大于0"); return }
	if sub.Attachments == "" { sub.Attachments = "[]" }
	sub.ApplicantID = getUID(r)
	h.Subsidy.Create(&sub)
	u, _ := h.User.GetByID(getUID(r))
	name := ""
	if u != nil { name = u.Name }
	h.Workflow.Log("subsidy", sub.ID, "", "submitted", "提交申请", getUID(r), name, "")
	JSON(w, 201, sub)
}

// 村委初审
func (h *Handler) CommitteeReviewSubsidy(w http.ResponseWriter, r *http.Request) {
	if !hasRole(r, "committee") {
		errJSON(w, 403, "需要村委委员以上权限"); return
	}
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	var req struct {
		Action string `json:"action"` // approve / reject
		Note   string `json:"note"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	approve := req.Action == "approve"
	h.Subsidy.CommitteeReview(id, getUID(r), approve, req.Note)

	toState := "secretary_review"
	if !approve { toState = "rejected" }
	u, _ := h.User.GetByID(getUID(r))
	name := ""
	if u != nil { name = u.Name }
	h.Workflow.Log("subsidy", id, "submitted", toState, "村委初审:"+req.Action, getUID(r), name, req.Note)
	// 通知申请人
	sub, _ := h.Subsidy.Get(id)
	if sub != nil {
		msg := "您的补贴申请「" + sub.Title + "」已通过村委初审，等待村支书终审"
		if !approve { msg = "您的补贴申请「" + sub.Title + "」被村委驳回：" + req.Note }
		h.Notify.Create(sub.ApplicantID, "补贴审批通知", msg, "subsidy", "subsidy", id)
	}
	JSON(w, 200, map[string]string{"ok": toState})
}

// 村支书终审
func (h *Handler) SecretaryReviewSubsidy(w http.ResponseWriter, r *http.Request) {
	if !hasRole(r, "secretary") {
		errJSON(w, 403, "需要村支书权限"); return
	}
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	var req struct {
		Action string `json:"action"`
		Note   string `json:"note"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	approve := req.Action == "approve"
	h.Subsidy.SecretaryReview(id, getUID(r), approve, req.Note)

	toState := "approved"
	if !approve { toState = "rejected" }
	u, _ := h.User.GetByID(getUID(r))
	name := ""
	if u != nil { name = u.Name }
	h.Workflow.Log("subsidy", id, "secretary_review", toState, "村支书终审:"+req.Action, getUID(r), name, req.Note)
	// 通知申请人
	sub, _ := h.Subsidy.Get(id)
	if sub != nil {
		msg := "您的补贴申请「" + sub.Title + "」已通过村支书终审"
		if !approve { msg = "您的补贴申请「" + sub.Title + "」被村支书驳回：" + req.Note }
		h.Notify.Create(sub.ApplicantID, "补贴审批通知", msg, "subsidy", "subsidy", id)
	}
	JSON(w, 200, map[string]string{"ok": toState})
}

// ==================== 工单 ====================

func (h *Handler) ListTickets(w http.ResponseWriter, r *http.Request) {
	page, size := pageParams(r)
	state := r.URL.Query().Get("state")
	var submitterID, assigneeID int64
	if r.URL.Query().Get("mine") == "1" {
		submitterID = getUID(r)
	}
	if r.URL.Query().Get("assigned") == "1" {
		assigneeID = getUID(r)
	}
	list, total, _ := h.Ticket.List(page, size, state, submitterID, assigneeID)
	JSON(w, 200, map[string]any{"data": list, "total": total, "page": page})
}

func (h *Handler) GetTicket(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	t, err := h.Ticket.Get(id)
	if err != nil { errJSON(w, 404, "工单不存在"); return }
	comments, _ := h.Ticket.ListComments(id)
	logs := h.Workflow.GetLogs("ticket", id)
	JSON(w, 200, map[string]any{"ticket": t, "comments": comments, "logs": logs})
}

func (h *Handler) CreateTicket(w http.ResponseWriter, r *http.Request) {
	var t model.Ticket
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		errJSON(w, 400, "请求格式错误"); return
	}
	if t.Title == "" || t.Content == "" {
		errJSON(w, 400, "标题和内容不能为空"); return
	}
	t.SubmitterID = getUID(r)
	if t.Images == "" { t.Images = "[]" }
	if t.Priority == "" { t.Priority = "normal" }
	h.Ticket.Create(&t)
	u, _ := h.User.GetByID(getUID(r))
	name := ""
	if u != nil { name = u.Name }
	h.Workflow.Log("ticket", t.ID, "", "open", "提交工单", getUID(r), name, "")
	JSON(w, 201, t)
}

// 分配工单（组长/委员以上）
func (h *Handler) AssignTicket(w http.ResponseWriter, r *http.Request) {
	if !hasRole(r, "group_leader") {
		errJSON(w, 403, "需要小组长以上权限"); return
	}
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	var req struct{ AssigneeID int64 `json:"assignee_id"` }
	json.NewDecoder(r.Body).Decode(&req)
	assignee := req.AssigneeID
	if assignee == 0 { assignee = getUID(r) } // 默认自己认领
	h.Ticket.Assign(id, assignee)
	u, _ := h.User.GetByID(getUID(r))
	name := ""
	if u != nil { name = u.Name }
	h.Workflow.Log("ticket", id, "open", "assigned", "分配工单", getUID(r), name, "")
	JSON(w, 200, map[string]string{"ok": "assigned"})
}

// 更新工单状态
func (h *Handler) UpdateTicketStatus(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	var req struct{ Status string `json:"status"` }
	json.NewDecoder(r.Body).Decode(&req)

	validStates := map[string]bool{"open": true, "assigned": true, "processing": true, "resolved": true, "closed": true}
	if !validStates[req.Status] {
		errJSON(w, 400, "无效的工单状态"); return
	}

	old, _ := h.Ticket.Get(id)
	// 村民只能关闭自己已解决的工单
	if req.Status == "closed" && old != nil && old.SubmitterID == getUID(r) {
		h.Ticket.UpdateState(id, "closed", 0)
	} else if hasRole(r, "group_leader") {
		h.Ticket.UpdateState(id, req.Status, getUID(r))
	} else {
		errJSON(w, 403, "权限不足"); return
	}

	u, _ := h.User.GetByID(getUID(r))
	name := ""
	if u != nil { name = u.Name }
	fromState := ""
	if old != nil { fromState = old.WorkflowState }
	h.Workflow.Log("ticket", id, fromState, req.Status, "状态变更", getUID(r), name, "")
	JSON(w, 200, map[string]string{"ok": req.Status})
}

func (h *Handler) AddTicketComment(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	var c model.TicketComment
	json.NewDecoder(r.Body).Decode(&c)
	c.TicketID = id
	c.UserID = getUID(r)
	h.Ticket.AddComment(&c)
	// 通知工单提交人（如果不是自己回复自己）
	t, _ := h.Ticket.Get(id)
	if t != nil && t.SubmitterID != getUID(r) {
		u, _ := h.User.GetByID(getUID(r))
		name := "管理员"
		if u != nil { name = u.Name }
		h.Notify.Create(t.SubmitterID, "工单回复", name+"回复了您的工单「"+t.Title+"」", "ticket", "ticket", id)
	}
	JSON(w, 201, c)
}

// ==================== 看板 ====================

func (h *Handler) Dashboard(w http.ResponseWriter, r *http.Request) {
	year := r.URL.Query().Get("year")
	if year == "" { year = fmt.Sprintf("%d", time.Now().Year()) }
	fin, _ := h.Finance.Summary(year)
	dash := model.Dashboard{
		NoticeCount:      h.Notice.Count(),
		NoticePending:    h.Notice.PendingCount(),
		TicketOpen:       h.Ticket.OpenCount(),
		TicketTotal:      h.Ticket.TotalCount(),
		SubsidyPending:   h.Subsidy.PendingCount(),
		SubsidyTotal:     h.Subsidy.TotalCount(),
		FinancePending:   h.Finance.PendingCount(),
		FinanceSummary:   fin,
		UserCount:        h.User.Count(),
		GroupCount:       h.Group.Count(),
		HouseholdCount:   h.Household.Count(),
		PartyMemberCount: h.User.PartyMemberCount(),
	}
	JSON(w, 200, dash)
}
