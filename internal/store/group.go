package store

import (
	"database/sql"
	"village-system/internal/model"
)

type GroupStore struct{ DB *sql.DB }

func (s *GroupStore) Create(g *model.Group) error {
	res, err := s.DB.Exec(`INSERT INTO groups (name, leader_id) VALUES (?,?)`, g.Name, g.LeaderID)
	if err != nil {
		return err
	}
	g.ID, _ = res.LastInsertId()
	return nil
}

func (s *GroupStore) Update(g *model.Group) error {
	_, err := s.DB.Exec(`UPDATE groups SET name=?, leader_id=? WHERE id=?`, g.Name, g.LeaderID, g.ID)
	return err
}

func (s *GroupStore) Delete(id int64) error {
	_, err := s.DB.Exec(`DELETE FROM groups WHERE id=?`, id)
	return err
}

func (s *GroupStore) List() ([]model.Group, error) {
	rows, err := s.DB.Query(
		`SELECT g.id, g.name, g.leader_id, COALESCE(u.name,''), g.created_at,
			(SELECT COUNT(*) FROM users WHERE group_id=g.id) as cnt
		FROM groups g LEFT JOIN users u ON g.leader_id=u.id ORDER BY g.id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]model.Group, 0)
	for rows.Next() {
		var g model.Group
		rows.Scan(&g.ID, &g.Name, &g.LeaderID, &g.LeaderName, &g.CreatedAt, &g.MemberCount)
		list = append(list, g)
	}
	return list, nil
}

func (s *GroupStore) Count() int {
	var c int
	s.DB.QueryRow(`SELECT COUNT(*) FROM groups`).Scan(&c)
	return c
}
