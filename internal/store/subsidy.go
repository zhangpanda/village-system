package store

import (
	"database/sql"
	"village-system/internal/model"
)

type SubsidyStore struct{ DB *sql.DB }

func (s *SubsidyStore) Create(sub *model.Subsidy) error {
	res, err := s.DB.Exec(
		`INSERT INTO subsidies (title, type, amount, reason, attachments, applicant_id, workflow_state) VALUES (?,?,?,?,?,?,?)`,
		sub.Title, sub.Type, sub.Amount, sub.Reason, sub.Attachments, sub.ApplicantID, "submitted",
	)
	if err != nil {
		return err
	}
	sub.ID, _ = res.LastInsertId()
	sub.WorkflowState = "submitted"
	return nil
}

// 村委初审
func (s *SubsidyStore) CommitteeReview(id, reviewerID int64, approve bool, note string) error {
	state := "secretary_review"
	if !approve {
		state = "rejected"
	}
	_, err := s.DB.Exec(
		`UPDATE subsidies SET workflow_state=?, committee_id=?, committee_note=?, committee_at=CURRENT_TIMESTAMP, current_step=1 WHERE id=? AND workflow_state='submitted'`,
		state, reviewerID, note, id,
	)
	return err
}

// 村支书终审
func (s *SubsidyStore) SecretaryReview(id, reviewerID int64, approve bool, note string) error {
	state := "approved"
	if !approve {
		state = "rejected"
	}
	_, err := s.DB.Exec(
		`UPDATE subsidies SET workflow_state=?, secretary_id=?, secretary_note=?, secretary_at=CURRENT_TIMESTAMP, current_step=2 WHERE id=? AND workflow_state='secretary_review'`,
		state, reviewerID, note, id,
	)
	return err
}

func (s *SubsidyStore) List(page, size int, state string, applicantID int64) ([]model.Subsidy, int, error) {
	where, args := " WHERE 1=1", []any{}
	if state != "" {
		where += " AND s.workflow_state=?"
		args = append(args, state)
	}
	if applicantID > 0 {
		where += " AND s.applicant_id=?"
		args = append(args, applicantID)
	}
	var total int
	s.DB.QueryRow(`SELECT COUNT(*) FROM subsidies s`+where, args...).Scan(&total)

	args = append(args, size, (page-1)*size)
	rows, err := s.DB.Query(
		`SELECT s.id, s.title, s.type, s.amount, s.reason, s.attachments, s.applicant_id, a.name,
			s.workflow_state, s.current_step,
			s.committee_id, COALESCE(c.name,''), s.committee_note, s.committee_at,
			s.secretary_id, COALESCE(r.name,''), s.secretary_note, s.secretary_at,
			s.created_at
		FROM subsidies s
		JOIN users a ON s.applicant_id = a.id
		LEFT JOIN users c ON s.committee_id = c.id
		LEFT JOIN users r ON s.secretary_id = r.id`+where+
			` ORDER BY s.created_at DESC LIMIT ? OFFSET ?`, args...,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	list := make([]model.Subsidy, 0)
	for rows.Next() {
		var sub model.Subsidy
		rows.Scan(&sub.ID, &sub.Title, &sub.Type, &sub.Amount, &sub.Reason, &sub.Attachments,
			&sub.ApplicantID, &sub.Applicant, &sub.WorkflowState, &sub.CurrentStep,
			&sub.CommitteeID, &sub.CommitteeName, &sub.CommitteeNote, &sub.CommitteeAt,
			&sub.SecretaryID, &sub.SecretaryName, &sub.SecretaryNote, &sub.SecretaryAt,
			&sub.CreatedAt)
		list = append(list, sub)
	}
	return list, total, nil
}

func (s *SubsidyStore) Get(id int64) (*model.Subsidy, error) {
	sub := &model.Subsidy{}
	err := s.DB.QueryRow(
		`SELECT s.id, s.title, s.type, s.amount, s.reason, s.attachments, s.applicant_id, a.name,
			s.workflow_state, s.current_step,
			s.committee_id, COALESCE(c.name,''), s.committee_note, s.committee_at,
			s.secretary_id, COALESCE(r.name,''), s.secretary_note, s.secretary_at,
			s.created_at
		FROM subsidies s
		JOIN users a ON s.applicant_id = a.id
		LEFT JOIN users c ON s.committee_id = c.id
		LEFT JOIN users r ON s.secretary_id = r.id
		WHERE s.id=?`, id,
	).Scan(&sub.ID, &sub.Title, &sub.Type, &sub.Amount, &sub.Reason, &sub.Attachments,
		&sub.ApplicantID, &sub.Applicant, &sub.WorkflowState, &sub.CurrentStep,
		&sub.CommitteeID, &sub.CommitteeName, &sub.CommitteeNote, &sub.CommitteeAt,
		&sub.SecretaryID, &sub.SecretaryName, &sub.SecretaryNote, &sub.SecretaryAt,
		&sub.CreatedAt)
	return sub, err
}

func (s *SubsidyStore) PendingCount() int {
	var c int
	s.DB.QueryRow(`SELECT COUNT(*) FROM subsidies WHERE workflow_state IN ('submitted','secretary_review')`).Scan(&c)
	return c
}

func (s *SubsidyStore) TotalCount() int {
	var c int
	s.DB.QueryRow(`SELECT COUNT(*) FROM subsidies`).Scan(&c)
	return c
}
