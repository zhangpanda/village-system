package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"village-system/internal/model"
)

func (h *Handler) ListHouseholds(w http.ResponseWriter, r *http.Request) {
	page, size := pageParams(r)
	groupID, _ := strconv.ParseInt(r.URL.Query().Get("group_id"), 10, 64)
	list, total, _ := h.Household.List(page, size, groupID)
	JSON(w, 200, map[string]any{"data": list, "total": total, "page": page})
}

func (h *Handler) GetHousehold(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	hh, err := h.Household.Get(id)
	if err != nil {
		errJSON(w, 404, "户籍不存在"); return
	}
	members, _ := h.Household.ListMembers(id)
	JSON(w, 200, map[string]any{"household": hh, "members": members})
}

func (h *Handler) CreateHousehold(w http.ResponseWriter, r *http.Request) {
	var hh model.Household
	json.NewDecoder(r.Body).Decode(&hh)
	if hh.HouseholdNo == "" {
		errJSON(w, 400, "户号不能为空"); return
	}
	h.Household.Create(&hh)
	JSON(w, 201, hh)
}

func (h *Handler) UpdateHousehold(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	var hh model.Household
	json.NewDecoder(r.Body).Decode(&hh)
	hh.ID = id
	h.Household.Update(&hh)
	JSON(w, 200, map[string]string{"ok": "updated"})
}

func (h *Handler) DeleteHousehold(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	h.Household.Delete(id)
	JSON(w, 200, map[string]string{"ok": "deleted"})
}

func (h *Handler) AddHouseholdMember(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	var req struct {
		UserID   int64  `json:"user_id"`
		Relation string `json:"relation"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	if req.UserID == 0 {
		errJSON(w, 400, "用户ID不能为空"); return
	}
	if req.Relation == "" {
		req.Relation = "其他"
	}
	// 一个用户只能加入一个户籍
	u, _ := h.User.GetByID(req.UserID)
	if u != nil && u.HouseholdID > 0 && u.HouseholdID != id {
		errJSON(w, 400, u.Name+"已属于其他户籍"); return
	}
	// 一个户籍只能有一个户主
	if req.Relation == "户主" {
		members, _ := h.Household.ListMembers(id)
		for _, m := range members {
			if m.Relation == "户主" {
				errJSON(w, 400, "该户已有户主："+m.UserName); return
			}
		}
	}
	h.Household.AddMember(id, req.UserID, req.Relation)
	JSON(w, 201, map[string]string{"ok": "added"})
}

func (h *Handler) RemoveHouseholdMember(w http.ResponseWriter, r *http.Request) {
	memberID, _ := strconv.ParseInt(r.PathValue("member_id"), 10, 64)
	h.Household.RemoveMember(memberID)
	JSON(w, 200, map[string]string{"ok": "removed"})
}

func (h *Handler) UpdateHouseholdMember(w http.ResponseWriter, r *http.Request) {
	memberID, _ := strconv.ParseInt(r.PathValue("member_id"), 10, 64)
	var req struct{ Relation string `json:"relation"` }
	json.NewDecoder(r.Body).Decode(&req)
	if req.Relation == "" { errJSON(w, 400, "关系不能为空"); return }
	h.Household.UpdateMemberRelation(memberID, req.Relation)
	JSON(w, 200, map[string]string{"ok": "updated"})
}
