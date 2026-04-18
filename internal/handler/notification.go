package handler

import (
	"net/http"
	"strconv"
)

func (h *Handler) ListNotifications(w http.ResponseWriter, r *http.Request) {
	page, size := pageParams(r)
	unread := r.URL.Query().Get("unread") == "1"
	list, total, _ := h.Notify.List(getUID(r), unread, page, size)
	count := h.Notify.UnreadCount(getUID(r))
	JSON(w, 200, map[string]any{"data": list, "total": total, "page": page, "unread_count": count})
}

func (h *Handler) MarkNotificationRead(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	h.Notify.MarkRead(id, getUID(r))
	JSON(w, 200, map[string]string{"ok": "已读"})
}

func (h *Handler) MarkAllNotificationsRead(w http.ResponseWriter, r *http.Request) {
	h.Notify.MarkAllRead(getUID(r))
	JSON(w, 200, map[string]string{"ok": "全部已读"})
}

func (h *Handler) UnreadNotificationCount(w http.ResponseWriter, r *http.Request) {
	count := h.Notify.UnreadCount(getUID(r))
	JSON(w, 200, map[string]any{"count": count})
}
