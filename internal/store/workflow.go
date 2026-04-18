package store

import (
	"database/sql"
	"village-system/internal/model"
)

type WorkflowStore struct{ DB *sql.DB }

func (s *WorkflowStore) Log(docType string, docID int64, from, to, action string, operatorID int64, operatorName, note string) {
	s.DB.Exec(
		`INSERT INTO workflow_logs (doc_type, doc_id, from_state, to_state, action, operator_id, operator_name, note) VALUES (?,?,?,?,?,?,?,?)`,
		docType, docID, from, to, action, operatorID, operatorName, note,
	)
}

func (s *WorkflowStore) GetLogs(docType string, docID int64) []model.WorkflowLog {
	rows, err := s.DB.Query(
		`SELECT id, doc_type, doc_id, from_state, to_state, action, operator_id, operator_name, note, created_at
		FROM workflow_logs WHERE doc_type=? AND doc_id=? ORDER BY created_at`, docType, docID,
	)
	if err != nil {
		return nil
	}
	defer rows.Close()
	list := make([]model.WorkflowLog, 0)
	for rows.Next() {
		var l model.WorkflowLog
		rows.Scan(&l.ID, &l.DocType, &l.DocID, &l.FromState, &l.ToState, &l.Action, &l.OperatorID, &l.OperatorName, &l.Note, &l.CreatedAt)
		list = append(list, l)
	}
	return list
}

func (s *WorkflowStore) ListAll(page, size int, docType string) ([]model.WorkflowLog, int) {
	where := "1=1"
	args := []any{}
	if docType != "" {
		where += " AND w.doc_type=?"
		args = append(args, docType)
	}
	var total int
	s.DB.QueryRow("SELECT COUNT(*) FROM workflow_logs w WHERE "+where, args...).Scan(&total)

	args = append(args, size, (page-1)*size)
	rows, err := s.DB.Query(`
		SELECT w.id, w.doc_type, w.doc_id, w.from_state, w.to_state, w.action, w.operator_id, w.operator_name, w.note, w.created_at,
			COALESCE(
				CASE w.doc_type
					WHEN 'notice' THEN (SELECT title FROM notices WHERE id=w.doc_id)
					WHEN 'subsidy' THEN (SELECT title FROM subsidies WHERE id=w.doc_id)
					WHEN 'ticket' THEN (SELECT title FROM tickets WHERE id=w.doc_id)
					WHEN 'finance' THEN (SELECT category||' '||CASE type WHEN 'income' THEN '收入' ELSE '支出' END||' '||CAST(amount/100.0 AS TEXT)||'元' FROM finance_records WHERE id=w.doc_id)
				END, ''
			) as doc_title
		FROM workflow_logs w WHERE `+where+` ORDER BY w.created_at DESC LIMIT ? OFFSET ?`, args...)
	if err != nil {
		return nil, total
	}
	defer rows.Close()
	list := make([]model.WorkflowLog, 0)
	for rows.Next() {
		var l model.WorkflowLog
		rows.Scan(&l.ID, &l.DocType, &l.DocID, &l.FromState, &l.ToState, &l.Action, &l.OperatorID, &l.OperatorName, &l.Note, &l.CreatedAt, &l.DocTitle)
		list = append(list, l)
	}
	return list, total
}
