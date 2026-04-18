package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"village-system/internal/model"
)

// ==================== 工作流定义管理 ====================

func (h *Handler) ListWorkflowDefs(w http.ResponseWriter, r *http.Request) {
	JSON(w, 200, h.WorkflowDef.ListAll())
}

func (h *Handler) GetWorkflowDef(w http.ResponseWriter, r *http.Request) {
	docType := r.URL.Query().Get("doc_type")
	if docType == "" {
		errJSON(w, 400, "缺少 doc_type 参数"); return
	}
	def := h.WorkflowDef.GetByDocType(docType)
	if def == nil {
		errJSON(w, 404, "未找到工作流定义"); return
	}
	JSON(w, 200, def)
}

func (h *Handler) SaveWorkflowDef(w http.ResponseWriter, r *http.Request) {
	var def struct {
		ID          int64  `json:"id"`
		Name        string `json:"name"`
		Label       string `json:"label"`
		DocType     string `json:"doc_type"`
		States      json.RawMessage `json:"states"`
		Transitions json.RawMessage `json:"transitions"`
		Active      bool   `json:"active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&def); err != nil {
		errJSON(w, 400, "请求格式错误"); return
	}
	if def.Name == "" || def.DocType == "" {
		errJSON(w, 400, "name 和 doc_type 不能为空"); return
	}

	// 解析 states 和 transitions 验证格式
	var m struct {
		States      interface{} `json:"states"`
		Transitions interface{} `json:"transitions"`
	}
	json.Unmarshal(def.States, &m.States)
	json.Unmarshal(def.Transitions, &m.Transitions)

	wf := &model.WorkflowDef{
		ID: def.ID, Name: def.Name, Label: def.Label, DocType: def.DocType, Active: def.Active,
	}
	json.Unmarshal(def.States, &wf.States)
	json.Unmarshal(def.Transitions, &wf.Transitions)

	if err := h.WorkflowDef.Upsert(wf); err != nil {
		errJSON(w, 500, err.Error()); return
	}
	JSON(w, 200, wf)
}

func (h *Handler) DeleteWorkflowDef(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	h.WorkflowDef.Delete(id)
	JSON(w, 200, map[string]string{"ok": "deleted"})
}

// ApplyTransition 通用工作流转换接口
// POST /api/admin/workflow/apply
// {"doc_type":"notice", "doc_id":1, "action":"approve", "note":"同意"}
func (h *Handler) ApplyTransition(w http.ResponseWriter, r *http.Request) {
	var req struct {
		DocType string `json:"doc_type"`
		DocID   int64  `json:"doc_id"`
		Action  string `json:"action"`
		Note    string `json:"note"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errJSON(w, 400, "请求格式错误"); return
	}

	userRole := getRole(r)
	uid := getUID(r)

	// 获取文档当前状态
	currentState := h.getDocState(req.DocType, req.DocID)
	if currentState == "" {
		errJSON(w, 404, "文档不存在"); return
	}

	// 查找匹配的转换规则
	trans := h.WorkflowDef.FindTransition(req.DocType, currentState, req.Action, userRole)
	if trans == nil {
		errJSON(w, 403, "当前状态不允许此操作，或权限不足"); return
	}

	// 执行状态变更
	if err := h.setDocState(req.DocType, req.DocID, trans.To, uid, req.Note); err != nil {
		errJSON(w, 500, err.Error()); return
	}

	// 记录日志
	u, _ := h.User.GetByID(uid)
	name := ""
	if u != nil { name = u.Name }
	h.Workflow.Log(req.DocType, req.DocID, currentState, trans.To, trans.Label, uid, name, req.Note)

	// 发通知
	if trans.Notify {
		h.notifyDocOwner(req.DocType, req.DocID, trans.Label, req.Note, name)
	}

	JSON(w, 200, map[string]any{"ok": trans.To, "from": currentState, "to": trans.To, "action": trans.Label})
}

// getDocState 获取文档当前工作流状态
func (h *Handler) getDocState(docType string, docID int64) string {
	var state string
	table := docTypeTable(docType)
	if table == "" { return "" }
	h.User.DB.QueryRow(`SELECT workflow_state FROM `+table+` WHERE id=?`, docID).Scan(&state)
	return state
}

// setDocState 更新文档工作流状态
func (h *Handler) setDocState(docType string, docID int64, state string, reviewerID int64, note string) error {
	table := docTypeTable(docType)
	if table == "" { return nil }
	switch docType {
	case "notice":
		return h.Notice.UpdateState(docID, state, reviewerID, note)
	case "finance":
		return h.Finance.UpdateState(docID, state, reviewerID, note)
	case "subsidy":
		// 根据目标状态判断是初审还是终审
		if state == "secretary_review" || (state == "rejected" && h.getDocState("subsidy", docID) == "submitted") {
			h.Subsidy.CommitteeReview(docID, reviewerID, state == "secretary_review", note)
		} else {
			h.Subsidy.SecretaryReview(docID, reviewerID, state == "approved", note)
		}
		return nil
	case "ticket":
		h.Ticket.UpdateState(docID, state, reviewerID)
		return nil
	}
	return nil
}

// notifyDocOwner 通知文档所有者
func (h *Handler) notifyDocOwner(docType string, docID int64, action, note, operatorName string) {
	var ownerID int64
	var title string
	switch docType {
	case "notice":
		n, _ := h.Notice.Get(docID)
		if n != nil { ownerID = n.AuthorID; title = n.Title }
	case "subsidy":
		s, _ := h.Subsidy.Get(docID)
		if s != nil { ownerID = s.ApplicantID; title = s.Title }
	case "ticket":
		t, _ := h.Ticket.Get(docID)
		if t != nil { ownerID = t.SubmitterID; title = t.Title }
	default:
		return
	}
	if ownerID > 0 {
		msg := operatorName + " " + action + "了「" + title + "」"
		if note != "" { msg += "：" + note }
		h.Notify.Create(ownerID, docType+"审批通知", msg, docType, docType, docID)
		// 同时尝试微信推送
		h.NotifyViaWechat(ownerID, docType+"审批通知", msg, docType, docID)
	}
}

func docTypeTable(docType string) string {
	switch docType {
	case "notice": return "notices"
	case "finance": return "finance_records"
	case "subsidy": return "subsidies"
	case "ticket": return "tickets"
	}
	return ""
}

func (h *Handler) ListWorkflowLogs(w http.ResponseWriter, r *http.Request) {
	page, size := pageParams(r)
	docType := r.URL.Query().Get("doc_type")
	list, total := h.Workflow.ListAll(page, size, docType)
	JSON(w, 200, map[string]any{"data": list, "total": total})
}
