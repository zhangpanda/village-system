package store

import (
	"database/sql"
	"time"
)

type Notification struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Type      string    `json:"type"`      // subsidy, ticket, notice, system
	RefType   string    `json:"ref_type"`  // subsidy, ticket, notice
	RefID     int64     `json:"ref_id"`
	IsRead    bool      `json:"is_read"`
	CreatedAt time.Time `json:"created_at"`
}

type NotificationStore struct{ DB *sql.DB }

func (s *NotificationStore) Create(userID int64, title, content, typ, refType string, refID int64) {
	s.DB.Exec(
		`INSERT INTO notifications (user_id, title, content, type, ref_type, ref_id) VALUES (?,?,?,?,?,?)`,
		userID, title, content, typ, refType, refID,
	)
}

// 批量通知（给多个用户发同一条）
func (s *NotificationStore) CreateBatch(userIDs []int64, title, content, typ, refType string, refID int64) {
	for _, uid := range userIDs {
		s.Create(uid, title, content, typ, refType, refID)
	}
}

func (s *NotificationStore) List(userID int64, onlyUnread bool, page, size int) ([]Notification, int, error) {
	where, args := " WHERE user_id=?", []any{userID}
	if onlyUnread {
		where += " AND is_read=0"
	}
	var total int
	s.DB.QueryRow(`SELECT COUNT(*) FROM notifications`+where, args...).Scan(&total)

	args = append(args, size, (page-1)*size)
	rows, err := s.DB.Query(
		`SELECT id, user_id, title, content, type, ref_type, ref_id, is_read, created_at
		FROM notifications`+where+` ORDER BY created_at DESC LIMIT ? OFFSET ?`, args...,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	list := make([]Notification, 0)
	for rows.Next() {
		var n Notification
		rows.Scan(&n.ID, &n.UserID, &n.Title, &n.Content, &n.Type, &n.RefType, &n.RefID, &n.IsRead, &n.CreatedAt)
		list = append(list, n)
	}
	return list, total, nil
}

func (s *NotificationStore) MarkRead(id, userID int64) {
	s.DB.Exec(`UPDATE notifications SET is_read=1 WHERE id=? AND user_id=?`, id, userID)
}

func (s *NotificationStore) MarkAllRead(userID int64) {
	s.DB.Exec(`UPDATE notifications SET is_read=1 WHERE user_id=? AND is_read=0`, userID)
}

func (s *NotificationStore) UnreadCount(userID int64) int {
	var c int
	s.DB.QueryRow(`SELECT COUNT(*) FROM notifications WHERE user_id=? AND is_read=0`, userID).Scan(&c)
	return c
}
