package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"village-system/internal/middleware"
	"village-system/internal/model"
	"village-system/internal/store"
)

type Handler struct {
	User        *store.UserStore
	Notice      *store.NoticeStore
	Finance     *store.FinanceStore
	Subsidy     *store.SubsidyStore
	Ticket      *store.TicketStore
	Group       *store.GroupStore
	Household   *store.HouseholdStore
	Workflow    *store.WorkflowStore
	WorkflowDef *store.WorkflowDefStore
	Notify      *store.NotificationStore
	Report      *store.ReportStore
	VillageName string
}

func JSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}

func errJSON(w http.ResponseWriter, code int, msg string) {
	JSON(w, code, map[string]string{"error": msg})
}

func getUID(r *http.Request) int64 {
	if v := r.Context().Value(middleware.UserIDKey); v != nil {
		return v.(int64)
	}
	return 0
}

func getRole(r *http.Request) string {
	if v := r.Context().Value(middleware.UserRoleKey); v != nil {
		return v.(string)
	}
	return ""
}

func hasRole(r *http.Request, minRole string) bool {
	return model.HasRole(getRole(r), minRole)
}

func isReadOnly(r *http.Request) bool {
	return model.IsReadOnly(getRole(r))
}

func pageParams(r *http.Request) (int, int) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	size, _ := strconv.Atoi(r.URL.Query().Get("size"))
	if page < 1 { page = 1 }
	if size < 1 || size > 100 { size = 20 }
	return page, size
}

func validPhone(p string) bool {
	if len(p) != 11 { return false }
	for _, c := range p {
		if c < '0' || c > '9' { return false }
	}
	return true
}

func (h *Handler) SiteConfig(w http.ResponseWriter, r *http.Request) {
	JSON(w, 200, map[string]any{
		"village_name":     h.VillageName,
		"roles":            model.RoleLabel,
		"wx_phone_enabled": wxAppID != "",
	})
}
