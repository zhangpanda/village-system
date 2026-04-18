package store

import (
	"database/sql"
	"village-system/internal/model"
)

type TicketStore struct{ DB *sql.DB }

func (s *TicketStore) Create(t *model.Ticket) error {
	res, err := s.DB.Exec(
		`INSERT INTO tickets (title, content, category, images, priority, submitter_id, workflow_state) VALUES (?,?,?,?,?,?,?)`,
		t.Title, t.Content, t.Category, t.Images, t.Priority, t.SubmitterID, "open",
	)
	if err != nil {
		return err
	}
	t.ID, _ = res.LastInsertId()
	t.WorkflowState = "open"
	return nil
}

func (s *TicketStore) UpdateState(id int64, state string, assigneeID int64) error {
	extra := ""
	if state == "resolved" {
		extra = ", resolved_at=CURRENT_TIMESTAMP"
	} else if state == "closed" {
		extra = ", closed_at=CURRENT_TIMESTAMP"
	}
	_, err := s.DB.Exec(
		`UPDATE tickets SET workflow_state=?, assignee_id=CASE WHEN ?=0 THEN assignee_id ELSE ? END`+extra+`, updated_at=CURRENT_TIMESTAMP WHERE id=?`,
		state, assigneeID, assigneeID, id,
	)
	return err
}

func (s *TicketStore) Assign(id, assigneeID int64) error {
	_, err := s.DB.Exec(
		`UPDATE tickets SET workflow_state='assigned', assignee_id=?, updated_at=CURRENT_TIMESTAMP WHERE id=? AND workflow_state='open'`,
		assigneeID, id,
	)
	return err
}

func (s *TicketStore) AddComment(c *model.TicketComment) error {
	res, err := s.DB.Exec(
		`INSERT INTO ticket_comments (ticket_id, user_id, content) VALUES (?,?,?)`,
		c.TicketID, c.UserID, c.Content,
	)
	if err != nil {
		return err
	}
	c.ID, _ = res.LastInsertId()
	s.DB.Exec(`UPDATE tickets SET updated_at=CURRENT_TIMESTAMP WHERE id=?`, c.TicketID)
	return nil
}

func (s *TicketStore) List(page, size int, state string, submitterID int64, assigneeID int64) ([]model.Ticket, int, error) {
	where, args := " WHERE 1=1", []any{}
	if state != "" {
		where += " AND t.workflow_state=?"
		args = append(args, state)
	}
	if submitterID > 0 {
		where += " AND t.submitter_id=?"
		args = append(args, submitterID)
	}
	if assigneeID > 0 {
		where += " AND t.assignee_id=?"
		args = append(args, assigneeID)
	}
	var total int
	s.DB.QueryRow(`SELECT COUNT(*) FROM tickets t`+where, args...).Scan(&total)

	args = append(args, size, (page-1)*size)
	rows, err := s.DB.Query(
		`SELECT t.id, t.title, t.content, t.category, t.images, t.priority, t.workflow_state,
			t.submitter_id, s.name, t.assignee_id, COALESCE(a.name,''), t.resolved_at, t.closed_at, t.created_at, t.updated_at
		FROM tickets t
		JOIN users s ON t.submitter_id = s.id
		LEFT JOIN users a ON t.assignee_id = a.id`+where+
			` ORDER BY CASE t.priority WHEN 'urgent' THEN 0 WHEN 'normal' THEN 1 ELSE 2 END, t.updated_at DESC LIMIT ? OFFSET ?`, args...,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	list := make([]model.Ticket, 0)
	for rows.Next() {
		var t model.Ticket
		rows.Scan(&t.ID, &t.Title, &t.Content, &t.Category, &t.Images, &t.Priority, &t.WorkflowState,
			&t.SubmitterID, &t.Submitter, &t.AssigneeID, &t.Assignee, &t.ResolvedAt, &t.ClosedAt, &t.CreatedAt, &t.UpdatedAt)
		list = append(list, t)
	}
	return list, total, nil
}

func (s *TicketStore) Get(id int64) (*model.Ticket, error) {
	t := &model.Ticket{}
	err := s.DB.QueryRow(
		`SELECT t.id, t.title, t.content, t.category, t.images, t.priority, t.workflow_state,
			t.submitter_id, s.name, t.assignee_id, COALESCE(a.name,''), t.resolved_at, t.closed_at, t.created_at, t.updated_at
		FROM tickets t
		JOIN users s ON t.submitter_id = s.id
		LEFT JOIN users a ON t.assignee_id = a.id
		WHERE t.id=?`, id,
	).Scan(&t.ID, &t.Title, &t.Content, &t.Category, &t.Images, &t.Priority, &t.WorkflowState,
		&t.SubmitterID, &t.Submitter, &t.AssigneeID, &t.Assignee, &t.ResolvedAt, &t.ClosedAt, &t.CreatedAt, &t.UpdatedAt)
	return t, err
}

func (s *TicketStore) ListComments(ticketID int64) ([]model.TicketComment, error) {
	rows, err := s.DB.Query(
		`SELECT c.id, c.ticket_id, c.user_id, u.name, c.content, c.created_at
		FROM ticket_comments c JOIN users u ON c.user_id = u.id
		WHERE c.ticket_id=? ORDER BY c.created_at`, ticketID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]model.TicketComment, 0)
	for rows.Next() {
		var c model.TicketComment
		rows.Scan(&c.ID, &c.TicketID, &c.UserID, &c.UserName, &c.Content, &c.CreatedAt)
		list = append(list, c)
	}
	return list, nil
}

func (s *TicketStore) OpenCount() int {
	var c int
	s.DB.QueryRow(`SELECT COUNT(*) FROM tickets WHERE workflow_state IN ('open','assigned','processing')`).Scan(&c)
	return c
}

func (s *TicketStore) TotalCount() int {
	var c int
	s.DB.QueryRow(`SELECT COUNT(*) FROM tickets`).Scan(&c)
	return c
}
