package store

import (
	"database/sql"
	"village-system/internal/model"
)

type UserStore struct{ DB *sql.DB }

const userSelectCols = `u.id, u.account, u.phone, u.name, u.gender, u.birth_date, u.ethnicity, u.education, u.marital_status, u.role, u.position,
	u.id_card, u.address, u.group_id, COALESCE(g.name,''), u.household_id,
	u.is_party_member, u.is_low_income, u.is_five_guarantee, u.is_disabled, u.is_military,
	u.wechat_id, u.emergency_contact, u.emergency_phone,
	u.openid, u.avatar_url, u.active, u.remark, u.created_at, u.updated_at`

const userSelectColsPwd = `u.id, u.account, u.phone, u.name, u.gender, u.birth_date, u.ethnicity, u.education, u.marital_status, u.role, u.position,
	u.id_card, u.address, u.group_id, u.household_id,
	u.is_party_member, u.is_low_income, u.is_five_guarantee, u.is_disabled, u.is_military,
	u.wechat_id, u.emergency_contact, u.emergency_phone,
	u.openid, u.avatar_url, u.active, u.password_hash, u.remark, u.created_at, u.updated_at`

func scanUser(row interface{ Scan(...any) error }, u *model.User) error {
	err := row.Scan(&u.ID, &u.Account, &u.Phone, &u.Name, &u.Gender, &u.BirthDate, &u.Ethnicity, &u.Education, &u.MaritalStatus, &u.Role, &u.Position,
		&u.IDCard, &u.Address, &u.GroupID, &u.GroupName, &u.HouseholdID,
		&u.IsPartyMember, &u.IsLowIncome, &u.IsFiveGuarantee, &u.IsDisabled, &u.IsMilitary,
		&u.WechatID, &u.EmergencyContact, &u.EmergencyPhone,
		&u.OpenID, &u.AvatarURL, &u.Active, &u.Remark, &u.CreatedAt, &u.UpdatedAt)
	if err == nil {
		u.RoleLabel = model.RoleLabels(u.Role)
		u.PositionLabel = model.PositionLabels[u.Position]
	}
	return err
}

func scanUserPwd(row interface{ Scan(...any) error }, u *model.User) error {
	err := row.Scan(&u.ID, &u.Account, &u.Phone, &u.Name, &u.Gender, &u.BirthDate, &u.Ethnicity, &u.Education, &u.MaritalStatus, &u.Role, &u.Position,
		&u.IDCard, &u.Address, &u.GroupID, &u.HouseholdID,
		&u.IsPartyMember, &u.IsLowIncome, &u.IsFiveGuarantee, &u.IsDisabled, &u.IsMilitary,
		&u.WechatID, &u.EmergencyContact, &u.EmergencyPhone,
		&u.OpenID, &u.AvatarURL, &u.Active, &u.PasswordHash, &u.Remark, &u.CreatedAt, &u.UpdatedAt)
	if err == nil {
		u.RoleLabel = model.RoleLabels(u.Role)
		u.PositionLabel = model.PositionLabels[u.Position]
	}
	return err
}

func (s *UserStore) Create(u *model.User) error {
	if u.Account == "" { u.Account = u.Phone }
	res, err := s.DB.Exec(
		`INSERT INTO users (account, phone, name, gender, birth_date, ethnicity, education, marital_status, role, position, id_card, address,
		group_id, household_id, is_party_member, is_low_income, is_five_guarantee, is_disabled, is_military,
		wechat_id, emergency_contact, emergency_phone, remark, password_hash)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		u.Account, u.Phone, u.Name, u.Gender, u.BirthDate, u.Ethnicity, u.Education, u.MaritalStatus, u.Role, u.Position, u.IDCard, u.Address,
		u.GroupID, u.HouseholdID, u.IsPartyMember, u.IsLowIncome, u.IsFiveGuarantee, u.IsDisabled, u.IsMilitary,
		u.WechatID, u.EmergencyContact, u.EmergencyPhone, u.Remark, u.PasswordHash,
	)
	if err != nil {
		return err
	}
	u.ID, _ = res.LastInsertId()
	return nil
}

func (s *UserStore) GetByPhone(phone string) (*model.User, error) {
	u := &model.User{}
	err := scanUserPwd(s.DB.QueryRow(`SELECT `+userSelectColsPwd+` FROM users u WHERE u.account=? OR u.phone=?`, phone, phone), u)
	if err != nil { return nil, err }
	return u, nil
}

func (s *UserStore) GetByPhoneOnly(phone string) (*model.User, error) {
	u := &model.User{}
	err := scanUserPwd(s.DB.QueryRow(`SELECT `+userSelectColsPwd+` FROM users u WHERE u.phone=?`, phone), u)
	if err != nil { return nil, err }
	return u, nil
}

func (s *UserStore) GetByID(id int64) (*model.User, error) {
	u := &model.User{}
	err := scanUser(s.DB.QueryRow(`SELECT `+userSelectCols+` FROM users u LEFT JOIN groups g ON u.group_id=g.id WHERE u.id=?`, id), u)
	return u, err
}

func (s *UserStore) GetByIDWithPwd(id int64) (*model.User, error) {
	u := &model.User{}
	err := scanUserPwd(s.DB.QueryRow(`SELECT `+userSelectColsPwd+` FROM users u WHERE u.id=?`, id), u)
	if err != nil { return nil, err }
	return u, nil
}

func (s *UserStore) GetByOpenID(openid string) (*model.User, error) {
	u := &model.User{}
	err := scanUser(s.DB.QueryRow(`SELECT `+userSelectCols+` FROM users u LEFT JOIN groups g ON u.group_id=g.id WHERE u.openid=?`, openid), u)
	if err != nil { return nil, err }
	return u, nil
}

func (s *UserStore) List(page, size int, filters map[string]string, groupID int64) ([]model.User, int, error) {
	where, args := " WHERE 1=1", []any{}
	if v := filters["role"]; v != "" {
		where += " AND (',' || u.role || ',' LIKE '%,' || ? || ',%')"
		args = append(args, v)
	}
	if v := filters["q"]; v != "" {
		where += " AND (u.name LIKE ? OR u.phone LIKE ?)"
		kw := "%" + v + "%"
		args = append(args, kw, kw)
	}
	if groupID > 0 {
		where += " AND u.group_id=?"
		args = append(args, groupID)
	}
	if v := filters["gender"]; v != "" {
		where += " AND u.gender=?"
		args = append(args, v)
	}
	if v := filters["education"]; v != "" {
		where += " AND u.education=?"
		args = append(args, v)
	}
	if v := filters["marital_status"]; v != "" {
		where += " AND u.marital_status=?"
		args = append(args, v)
	}
	if v := filters["tag"]; v != "" {
		switch v {
		case "party": where += " AND u.is_party_member=1"
		case "low_income": where += " AND u.is_low_income=1"
		case "five_guarantee": where += " AND u.is_five_guarantee=1"
		case "disabled": where += " AND u.is_disabled=1"
		case "military": where += " AND u.is_military=1"
		}
	}
	var total int
	s.DB.QueryRow(`SELECT COUNT(*) FROM users u`+where, args...).Scan(&total)

	args = append(args, size, (page-1)*size)
	rows, err := s.DB.Query(`SELECT `+userSelectCols+` FROM users u LEFT JOIN groups g ON u.group_id=g.id`+where+` ORDER BY u.id DESC LIMIT ? OFFSET ?`, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	list := make([]model.User, 0)
	for rows.Next() {
		var u model.User
		scanUser(rows, &u)
		list = append(list, u)
	}
	return list, total, nil
}

func (s *UserStore) Update(u *model.User) error {
	_, err := s.DB.Exec(
		`UPDATE users SET name=?, gender=?, birth_date=?, ethnicity=?, education=?, marital_status=?, role=?, position=?, id_card=?, address=?,
		group_id=?, household_id=?, is_party_member=?, is_low_income=?, is_five_guarantee=?,
		is_disabled=?, is_military=?, wechat_id=?,
		emergency_contact=?, emergency_phone=?, active=?, remark=?, updated_at=CURRENT_TIMESTAMP WHERE id=?`,
		u.Name, u.Gender, u.BirthDate, u.Ethnicity, u.Education, u.MaritalStatus, u.Role, u.Position, u.IDCard, u.Address,
		u.GroupID, u.HouseholdID, u.IsPartyMember, u.IsLowIncome, u.IsFiveGuarantee,
		u.IsDisabled, u.IsMilitary, u.WechatID,
		u.EmergencyContact, u.EmergencyPhone, u.Active, u.Remark, u.ID,
	)
	return err
}

func (s *UserStore) UpdateProfile(userID int64, name, idCard, address string) error {
	_, err := s.DB.Exec(`UPDATE users SET name=?, id_card=?, address=?, updated_at=CURRENT_TIMESTAMP WHERE id=?`, name, idCard, address, userID)
	return err
}

func (s *UserStore) UpdatePhone(userID int64, phone string) error {
	_, err := s.DB.Exec(`UPDATE users SET phone=?, updated_at=CURRENT_TIMESTAMP WHERE id=?`, phone, userID)
	return err
}

func (s *UserStore) UpdateAccount(userID int64, account string) error {
	_, err := s.DB.Exec(`UPDATE users SET account=?, updated_at=CURRENT_TIMESTAMP WHERE id=?`, account, userID)
	return err
}

func (s *UserStore) UpdatePassword(userID int64, hash string) error {
	_, err := s.DB.Exec(`UPDATE users SET password_hash=?, updated_at=CURRENT_TIMESTAMP WHERE id=?`, hash, userID)
	return err
}

func (s *UserStore) BindOpenID(userID int64, openid string) error {
	_, err := s.DB.Exec(`UPDATE users SET openid=?, updated_at=CURRENT_TIMESTAMP WHERE id=?`, openid, userID)
	return err
}

func (s *UserStore) CreateWxUser(u *model.User) error {
	if u.Account == "" { u.Account = u.Phone }
	res, err := s.DB.Exec(
		`INSERT INTO users (account, phone, name, role, position, openid, avatar_url) VALUES (?,?,?,?,?,?,?)`,
		u.Account, u.Phone, u.Name, u.Role, "villager", u.OpenID, u.AvatarURL,
	)
	if err != nil {
		return err
	}
	u.ID, _ = res.LastInsertId()
	return nil
}

func (s *UserStore) Count() int {
	var c int
	s.DB.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&c)
	return c
}

func (s *UserStore) PartyMemberCount() int {
	var c int
	s.DB.QueryRow(`SELECT COUNT(*) FROM users WHERE is_party_member=1`).Scan(&c)
	return c
}

func (s *UserStore) HouseholdCount() int {
	var c int
	s.DB.QueryRow(`SELECT COUNT(*) FROM households`).Scan(&c)
	return c
}
