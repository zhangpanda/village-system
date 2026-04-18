package store

import (
	"database/sql"
	"village-system/internal/model"
)

type NoticeStore struct{ DB *sql.DB }

func (s *NoticeStore) Create(n *model.Notice) error {
	res, err := s.DB.Exec(
		`INSERT INTO notices (title, content, category, author_id, pinned, attachments, workflow_state) VALUES (?,?,?,?,?,?,?)`,
		n.Title, n.Content, n.Category, n.AuthorID, n.Pinned, n.Attachments, n.WorkflowState,
	)
	if err != nil {
		return err
	}
	n.ID, _ = res.LastInsertId()
	return nil
}

func (s *NoticeStore) Update(n *model.Notice) error {
	_, err := s.DB.Exec(
		`UPDATE notices SET title=?, content=?, category=?, pinned=?, attachments=?, updated_at=CURRENT_TIMESTAMP WHERE id=?`,
		n.Title, n.Content, n.Category, n.Pinned, n.Attachments, n.ID,
	)
	return err
}

func (s *NoticeStore) UpdateState(id int64, state string, reviewerID int64, note string) error {
	extra := ""
	if state == "published" {
		extra = ", published_at=CURRENT_TIMESTAMP"
	}
	_, err := s.DB.Exec(
		`UPDATE notices SET workflow_state=?, reviewer_id=?, review_note=?`+extra+`, updated_at=CURRENT_TIMESTAMP WHERE id=?`,
		state, reviewerID, note, id,
	)
	return err
}

func (s *NoticeStore) Get(id int64) (*model.Notice, error) {
	s.DB.Exec(`UPDATE notices SET views=views+1 WHERE id=?`, id)
	n := &model.Notice{}
	err := s.DB.QueryRow(
		`SELECT n.id, n.title, n.content, n.category, n.author_id, u.name, n.pinned, n.attachments, n.views,
			n.workflow_state, n.reviewer_id, COALESCE(r.name,''), n.review_note, n.published_at,
			n.created_at, n.updated_at
		FROM notices n JOIN users u ON n.author_id = u.id
		LEFT JOIN users r ON n.reviewer_id = r.id WHERE n.id=?`, id,
	).Scan(&n.ID, &n.Title, &n.Content, &n.Category, &n.AuthorID, &n.Author, &n.Pinned, &n.Attachments, &n.Views,
		&n.WorkflowState, &n.ReviewerID, &n.ReviewerName, &n.ReviewNote, &n.PublishedAt,
		&n.CreatedAt, &n.UpdatedAt)
	return n, err
}

// List 公开接口只返回已发布的
func (s *NoticeStore) List(page, size int, category, keyword string, includeAll bool) ([]model.Notice, int, error) {
	where, args := " WHERE 1=1", []any{}
	if !includeAll {
		where += " AND n.workflow_state='published'"
	}
	if category != "" {
		where += " AND n.category=?"
		args = append(args, category)
	}
	if keyword != "" {
		where += " AND (n.title LIKE ? OR n.content LIKE ?)"
		kw := "%" + keyword + "%"
		args = append(args, kw, kw)
	}
	return s.queryNotices(where, args, page, size)
}

// ListAdmin 管理端：支持按 state 在 SQL 层过滤
func (s *NoticeStore) ListAdmin(page, size int, category, state string) ([]model.Notice, int, error) {
	where, args := " WHERE 1=1", []any{}
	if state != "" {
		where += " AND n.workflow_state=?"
		args = append(args, state)
	}
	if category != "" {
		where += " AND n.category=?"
		args = append(args, category)
	}
	return s.queryNotices(where, args, page, size)
}

func (s *NoticeStore) queryNotices(where string, args []any, page, size int) ([]model.Notice, int, error) {
	var total int
	s.DB.QueryRow(`SELECT COUNT(*) FROM notices n`+where, args...).Scan(&total)

	args = append(args, size, (page-1)*size)
	rows, err := s.DB.Query(
		`SELECT n.id, n.title, n.content, n.category, n.author_id, u.name, n.pinned, n.attachments, n.views,
			n.workflow_state, n.reviewer_id, COALESCE(r.name,''), n.review_note, n.published_at,
			n.created_at, n.updated_at
		FROM notices n JOIN users u ON n.author_id = u.id
		LEFT JOIN users r ON n.reviewer_id = r.id`+where+
			` ORDER BY n.pinned DESC, n.created_at DESC LIMIT ? OFFSET ?`, args...,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	list := make([]model.Notice, 0)
	for rows.Next() {
		var n model.Notice
		rows.Scan(&n.ID, &n.Title, &n.Content, &n.Category, &n.AuthorID, &n.Author, &n.Pinned, &n.Attachments, &n.Views,
			&n.WorkflowState, &n.ReviewerID, &n.ReviewerName, &n.ReviewNote, &n.PublishedAt,
			&n.CreatedAt, &n.UpdatedAt)
		list = append(list, n)
	}
	return list, total, nil
}

func (s *NoticeStore) Delete(id int64) error {
	_, err := s.DB.Exec(`DELETE FROM notices WHERE id=?`, id)
	return err
}

func (s *NoticeStore) Count() int {
	var c int
	s.DB.QueryRow(`SELECT COUNT(*) FROM notices WHERE workflow_state='published'`).Scan(&c)
	return c
}

func (s *NoticeStore) PendingCount() int {
	var c int
	s.DB.QueryRow(`SELECT COUNT(*) FROM notices WHERE workflow_state='pending_review'`).Scan(&c)
	return c
}
