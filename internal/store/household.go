package store

import (
	"database/sql"
	"village-system/internal/model"
)

type HouseholdStore struct{ DB *sql.DB }

func (s *HouseholdStore) Create(h *model.Household) error {
	res, err := s.DB.Exec(
		`INSERT INTO households (household_no, head_id, address, group_id, farmland_area, forest_area, homesite_area, remark) VALUES (?,?,?,?,?,?,?,?)`,
		h.HouseholdNo, h.HeadID, h.Address, h.GroupID, h.FarmlandArea, h.ForestArea, h.HomeSiteArea, h.Remark,
	)
	if err != nil {
		return err
	}
	h.ID, _ = res.LastInsertId()
	return nil
}

func (s *HouseholdStore) Update(h *model.Household) error {
	_, err := s.DB.Exec(
		`UPDATE households SET household_no=?, head_id=?, address=?, group_id=?, farmland_area=?, forest_area=?, homesite_area=?, remark=? WHERE id=?`,
		h.HouseholdNo, h.HeadID, h.Address, h.GroupID, h.FarmlandArea, h.ForestArea, h.HomeSiteArea, h.Remark, h.ID,
	)
	return err
}

func (s *HouseholdStore) Delete(id int64) error {
	s.DB.Exec(`DELETE FROM household_members WHERE household_id=?`, id)
	_, err := s.DB.Exec(`DELETE FROM households WHERE id=?`, id)
	return err
}

func (s *HouseholdStore) List(page, size int, groupID int64) ([]model.Household, int, error) {
	where, args := " WHERE 1=1", []any{}
	if groupID > 0 {
		where += " AND h.group_id=?"
		args = append(args, groupID)
	}
	var total int
	s.DB.QueryRow(`SELECT COUNT(*) FROM households h`+where, args...).Scan(&total)

	args = append(args, size, (page-1)*size)
	rows, err := s.DB.Query(
		`SELECT h.id, h.household_no, h.head_id, COALESCE(u.name,''), h.address, h.group_id, COALESCE(g.name,''),
			(SELECT COUNT(*) FROM household_members WHERE household_id=h.id),
			h.farmland_area, h.forest_area, h.homesite_area, h.remark, h.created_at
		FROM households h
		LEFT JOIN users u ON h.head_id=u.id
		LEFT JOIN groups g ON h.group_id=g.id`+where+` ORDER BY h.id LIMIT ? OFFSET ?`, args...,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	list := make([]model.Household, 0)
	for rows.Next() {
		var hh model.Household
		rows.Scan(&hh.ID, &hh.HouseholdNo, &hh.HeadID, &hh.HeadName, &hh.Address, &hh.GroupID, &hh.GroupName,
			&hh.MemberCount, &hh.FarmlandArea, &hh.ForestArea, &hh.HomeSiteArea, &hh.Remark, &hh.CreatedAt)
		list = append(list, hh)
	}
	return list, total, nil
}

func (s *HouseholdStore) Get(id int64) (*model.Household, error) {
	h := &model.Household{}
	err := s.DB.QueryRow(
		`SELECT h.id, h.household_no, h.head_id, COALESCE(u.name,''), h.address, h.group_id, COALESCE(g.name,''),
			(SELECT COUNT(*) FROM household_members WHERE household_id=h.id),
			h.farmland_area, h.forest_area, h.homesite_area, h.remark, h.created_at
		FROM households h LEFT JOIN users u ON h.head_id=u.id LEFT JOIN groups g ON h.group_id=g.id WHERE h.id=?`, id,
	).Scan(&h.ID, &h.HouseholdNo, &h.HeadID, &h.HeadName, &h.Address, &h.GroupID, &h.GroupName,
		&h.MemberCount, &h.FarmlandArea, &h.ForestArea, &h.HomeSiteArea, &h.Remark, &h.CreatedAt)
	return h, err
}

func (s *HouseholdStore) ListMembers(householdID int64) ([]model.HouseholdMember, error) {
	rows, err := s.DB.Query(
		`SELECT m.id, m.household_id, m.user_id, u.name, m.relation
		FROM household_members m JOIN users u ON m.user_id=u.id WHERE m.household_id=?`, householdID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]model.HouseholdMember, 0)
	for rows.Next() {
		var m model.HouseholdMember
		rows.Scan(&m.ID, &m.HouseholdID, &m.UserID, &m.UserName, &m.Relation)
		list = append(list, m)
	}
	return list, nil
}

func (s *HouseholdStore) AddMember(householdID, userID int64, relation string) error {
	_, err := s.DB.Exec(`INSERT INTO household_members (household_id, user_id, relation) VALUES (?,?,?)`, householdID, userID, relation)
	if err == nil {
		s.DB.Exec(`UPDATE users SET household_id=? WHERE id=?`, householdID, userID)
	}
	return err
}

func (s *HouseholdStore) RemoveMember(id int64) error {
	var userID int64
	s.DB.QueryRow(`SELECT user_id FROM household_members WHERE id=?`, id).Scan(&userID)
	s.DB.Exec(`DELETE FROM household_members WHERE id=?`, id)
	if userID > 0 {
		s.DB.Exec(`UPDATE users SET household_id=0 WHERE id=?`, userID)
	}
	return nil
}

func (s *HouseholdStore) UpdateMemberRelation(id int64, relation string) {
	s.DB.Exec(`UPDATE household_members SET relation=? WHERE id=?`, relation, id)
}

func (s *HouseholdStore) Count() int {
	var c int
	s.DB.QueryRow(`SELECT COUNT(*) FROM households`).Scan(&c)
	return c
}
