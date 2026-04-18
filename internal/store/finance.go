package store

import (
	"database/sql"
	"village-system/internal/model"
)

type FinanceStore struct{ DB *sql.DB }

func (s *FinanceStore) Create(r *model.FinanceRecord) error {
	res, err := s.DB.Exec(
		`INSERT INTO finance_records (type, amount, category, remark, date, voucher, author_id, workflow_state) VALUES (?,?,?,?,?,?,?,?)`,
		r.Type, r.Amount, r.Category, r.Remark, r.Date, r.Voucher, r.AuthorID, r.WorkflowState,
	)
	if err != nil {
		return err
	}
	r.ID, _ = res.LastInsertId()
	return nil
}

func (s *FinanceStore) UpdateState(id int64, state string, reviewerID int64, note string) error {
	_, err := s.DB.Exec(
		`UPDATE finance_records SET workflow_state=?, reviewer_id=?, review_note=? WHERE id=?`,
		state, reviewerID, note, id,
	)
	return err
}

// 公开接口只返回 approved
func (s *FinanceStore) List(page, size int, year, typ string, includeAll bool) ([]model.FinanceRecord, int, error) {
	where, args := " WHERE 1=1", []any{}
	if !includeAll {
		where += " AND workflow_state='approved'"
	}
	if year != "" {
		where += " AND date LIKE ?"
		args = append(args, year+"%")
	}
	if typ != "" {
		where += " AND type=?"
		args = append(args, typ)
	}
	var total int
	s.DB.QueryRow(`SELECT COUNT(*) FROM finance_records`+where, args...).Scan(&total)

	args = append(args, size, (page-1)*size)
	rows, err := s.DB.Query(
		`SELECT f.id, f.type, f.amount, f.category, f.remark, f.date, f.voucher, f.author_id, u.name,
			f.workflow_state, f.reviewer_id, COALESCE(r.name,''), f.review_note, f.created_at
		FROM finance_records f JOIN users u ON f.author_id = u.id
		LEFT JOIN users r ON f.reviewer_id = r.id`+where+
			` ORDER BY f.date DESC, f.id DESC LIMIT ? OFFSET ?`, args...,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	list := make([]model.FinanceRecord, 0)
	for rows.Next() {
		var r model.FinanceRecord
		rows.Scan(&r.ID, &r.Type, &r.Amount, &r.Category, &r.Remark, &r.Date, &r.Voucher, &r.AuthorID, &r.Author,
			&r.WorkflowState, &r.ReviewerID, &r.ReviewerName, &r.ReviewNote, &r.CreatedAt)
		list = append(list, r)
	}
	return list, total, nil
}

// Summary 只统计已审核的
func (s *FinanceStore) Summary(year string) (*model.FinanceSummary, error) {
	where, args := " WHERE workflow_state='approved'", []any{}
	if year != "" {
		where += " AND date LIKE ?"
		args = append(args, year+"%")
	}
	sum := &model.FinanceSummary{}
	s.DB.QueryRow(`SELECT COALESCE(SUM(CASE WHEN type='income' THEN amount ELSE 0 END),0),
		COALESCE(SUM(CASE WHEN type='expense' THEN amount ELSE 0 END),0)
		FROM finance_records`+where, args...).Scan(&sum.TotalIncome, &sum.TotalExpense)
	sum.Balance = sum.TotalIncome - sum.TotalExpense

	rows, err := s.DB.Query(
		`SELECT type, category, SUM(amount) FROM finance_records`+where+` GROUP BY type, category ORDER BY SUM(amount) DESC`, args...,
	)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var c model.CategorySummary
			rows.Scan(&c.Type, &c.Category, &c.Amount)
			sum.ByCategory = append(sum.ByCategory, c)
		}
	}
	return sum, nil
}

func (s *FinanceStore) Delete(id int64) error {
	_, err := s.DB.Exec(`DELETE FROM finance_records WHERE id=?`, id)
	return err
}

func (s *FinanceStore) Get(id int64) (*model.FinanceRecord, error) {
	var r model.FinanceRecord
	err := s.DB.QueryRow(`SELECT id, author_id, workflow_state FROM finance_records WHERE id=?`, id).Scan(&r.ID, &r.AuthorID, &r.WorkflowState)
	if err != nil { return nil, err }
	return &r, nil
}

func (s *FinanceStore) PendingCount() int {
	var c int
	s.DB.QueryRow(`SELECT COUNT(*) FROM finance_records WHERE workflow_state='pending_review'`).Scan(&c)
	return c
}
