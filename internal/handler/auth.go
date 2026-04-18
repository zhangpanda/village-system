package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"village-system/internal/middleware"
	"village-system/internal/model"
)

type loginReq struct {
	Phone    string `json:"phone"`
	Password string `json:"password"`
}

type registerReq struct {
	Phone    string `json:"phone"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

// 登录失败锁定：连续5次错误锁定15分钟
var loginAttempts = struct {
	sync.Mutex
	m map[string]*loginRecord
}{m: make(map[string]*loginRecord)}

type loginRecord struct {
	fails    int
	lockedAt time.Time
}

func checkLoginLock(phone string) (blocked bool, remaining time.Duration) {
	loginAttempts.Lock()
	defer loginAttempts.Unlock()
	r := loginAttempts.m[phone]
	if r == nil { return false, 0 }
	if r.fails >= 5 && time.Since(r.lockedAt) < 15*time.Minute {
		return true, 15*time.Minute - time.Since(r.lockedAt)
	}
	if time.Since(r.lockedAt) >= 15*time.Minute {
		delete(loginAttempts.m, phone)
	}
	return false, 0
}

func recordLoginFail(phone string) {
	loginAttempts.Lock()
	defer loginAttempts.Unlock()
	r := loginAttempts.m[phone]
	if r == nil {
		r = &loginRecord{}
		loginAttempts.m[phone] = r
	}
	r.fails++
	r.lockedAt = time.Now()
}

func clearLoginFails(phone string) {
	loginAttempts.Lock()
	delete(loginAttempts.m, phone)
	loginAttempts.Unlock()
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errJSON(w, 400, "请求格式错误"); return
	}
	if req.Phone == "" || req.Password == "" {
		errJSON(w, 400, "手机号和密码不能为空"); return
	}
	if locked, remain := checkLoginLock(req.Phone); locked {
		errJSON(w, 429, "登录失败次数过多，请"+strconv.Itoa(int(remain.Minutes())+1)+"分钟后再试"); return
	}
	u, err := h.User.GetByPhone(req.Phone)
	if err != nil || !middleware.CheckPassword(u.PasswordHash, req.Password) {
		recordLoginFail(req.Phone)
		errJSON(w, 401, "用户名或密码错误"); return
	}
	if !u.Active {
		errJSON(w, 403, "账号已被禁用"); return
	}
	clearLoginFails(req.Phone)
	token, _ := middleware.GenerateToken(u.ID, u.Role)
	JSON(w, 200, map[string]any{"token": token, "user": u})
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errJSON(w, 400, "请求格式错误"); return
	}
	if req.Phone == "" || req.Name == "" || req.Password == "" {
		errJSON(w, 400, "手机号、姓名、密码不能为空"); return
	}
	if !validPhone(req.Phone) {
		errJSON(w, 400, "手机号格式不正确"); return
	}
	if len(req.Password) < 6 {
		errJSON(w, 400, "密码至少6位"); return
	}
	u := &model.User{
		Account: req.Phone, Phone: req.Phone, Name: req.Name, Role: "villager",
		PasswordHash: middleware.HashPassword(req.Password),
	}
	if err := h.User.Create(u); err != nil {
		errJSON(w, 400, "手机号已注册"); return
	}
	token, _ := middleware.GenerateToken(u.ID, u.Role)
	JSON(w, 201, map[string]any{"token": token, "user": u})
}

func (h *Handler) GetProfile(w http.ResponseWriter, r *http.Request) {
	u, err := h.User.GetByID(getUID(r))
	if err != nil {
		errJSON(w, 404, "用户不存在"); return
	}
	full, _ := h.User.GetByIDWithPwd(getUID(r))
	hasPassword := full != nil && full.PasswordHash != ""
	u.PasswordHash = ""
	// 直接用 struct 序列化，补上 password_set
	type profileResp struct {
		*model.User
		PasswordSet bool `json:"password_set"`
	}
	JSON(w, 200, profileResp{User: u, PasswordSet: hasPassword})
}

func (h *Handler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name             *string `json:"name"`
		Gender           *string `json:"gender"`
		BirthDate        *string `json:"birth_date"`
		Ethnicity        *string `json:"ethnicity"`
		Education        *string `json:"education"`
		MaritalStatus    *string `json:"marital_status"`
		IDCard           *string `json:"id_card"`
		Address          *string `json:"address"`
		EmergencyContact *string `json:"emergency_contact"`
		EmergencyPhone   *string `json:"emergency_phone"`
		WechatID         *string `json:"wechat_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errJSON(w, 400, "请求格式错误"); return
	}
	uid := getUID(r)
	u, err := h.User.GetByID(uid)
	if err != nil { errJSON(w, 404, "用户不存在"); return }

	// 只更新前端传了的字段
	if req.Name != nil { u.Name = *req.Name }
	if req.Gender != nil { u.Gender = *req.Gender }
	if req.BirthDate != nil { u.BirthDate = *req.BirthDate }
	if req.Ethnicity != nil { u.Ethnicity = *req.Ethnicity }
	if req.Education != nil { u.Education = *req.Education }
	if req.MaritalStatus != nil { u.MaritalStatus = *req.MaritalStatus }
	if req.IDCard != nil { u.IDCard = *req.IDCard }
	if req.Address != nil { u.Address = *req.Address }
	if req.EmergencyContact != nil { u.EmergencyContact = *req.EmergencyContact }
	if req.EmergencyPhone != nil { u.EmergencyPhone = *req.EmergencyPhone }
	if req.WechatID != nil { u.WechatID = *req.WechatID }

	if u.Name == "" { errJSON(w, 400, "姓名不能为空"); return }

	h.User.DB.Exec(
		`UPDATE users SET name=?, gender=?, birth_date=?, ethnicity=?, education=?, marital_status=?,
		id_card=?, address=?, emergency_contact=?, emergency_phone=?, wechat_id=?, updated_at=CURRENT_TIMESTAMP WHERE id=?`,
		u.Name, u.Gender, u.BirthDate, u.Ethnicity, u.Education, u.MaritalStatus,
		u.IDCard, u.Address, u.EmergencyContact, u.EmergencyPhone, u.WechatID, uid,
	)
	JSON(w, 200, map[string]string{"ok": "更新成功"})
}

func (h *Handler) BindPhone(w http.ResponseWriter, r *http.Request) {
	var req struct{ Phone string `json:"phone"` }
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errJSON(w, 400, "请求格式错误"); return
	}
	if !validPhone(req.Phone) {
		errJSON(w, 400, "手机号格式不正确"); return
	}
	existing, _ := h.User.GetByPhoneOnly(req.Phone)
	if existing != nil && existing.ID != getUID(r) {
		errJSON(w, 400, "该手机号已被其他用户绑定"); return
	}
	h.User.UpdatePhone(getUID(r), req.Phone)
	// 微信用户绑手机号后，更新 account 为手机号（方便手机号登录）
	u, _ := h.User.GetByID(getUID(r))
	if u != nil && strings.HasPrefix(u.Account, "wx_") {
		h.User.UpdateAccount(getUID(r), req.Phone)
	}
	JSON(w, 200, map[string]string{"ok": "绑定成功"})
}

func (h *Handler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	var req struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errJSON(w, 400, "请求格式错误"); return
	}
	if len(req.NewPassword) < 6 {
		errJSON(w, 400, "新密码至少6位"); return
	}
	u, _ := h.User.GetByID(getUID(r))
	if u == nil {
		errJSON(w, 404, "用户不存在"); return
	}
	full, _ := h.User.GetByIDWithPwd(getUID(r))
	if full != nil && full.PasswordHash != "" && !middleware.CheckPassword(full.PasswordHash, req.OldPassword) {
		errJSON(w, 400, "旧密码错误"); return
	}
	h.User.UpdatePassword(getUID(r), middleware.HashPassword(req.NewPassword))
	JSON(w, 200, map[string]string{"ok": "密码修改成功"})
}

func (h *Handler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	uid := getUID(r)
	if uid == 0 {
		errJSON(w, 401, "请先登录"); return
	}
	var req struct {
		NewPassword string `json:"new_password"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	if len(req.NewPassword) < 6 { errJSON(w, 400, "新密码至少6位"); return }
	h.User.UpdatePassword(uid, middleware.HashPassword(req.NewPassword))
	JSON(w, 200, map[string]string{"ok": "密码重置成功"})
}

// 密码重置专用限流器
var resetLimiter = &ipLimiter{visitors: make(map[string]*rlEntry)}

type rlEntry struct {
	count    int
	resetAt  time.Time
}

type ipLimiter struct {
	mu       sync.Mutex
	visitors map[string]*rlEntry
}

func (l *ipLimiter) allow(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	e, ok := l.visitors[ip]
	now := time.Now()
	if !ok || now.After(e.resetAt) {
		l.visitors[ip] = &rlEntry{count: 1, resetAt: now.Add(time.Minute)}
		return true
	}
	e.count++
	return e.count <= 3
}

// ==================== 管理员用户管理 ====================

func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	page, size := pageParams(r)
	groupID, _ := strconv.ParseInt(r.URL.Query().Get("group_id"), 10, 64)
	if scopeGID := middleware.GetScopeGroupID(r); scopeGID > 0 && groupID == 0 {
		groupID = scopeGID
	}
	filters := map[string]string{}
	for _, k := range []string{"role", "q", "gender", "education", "marital_status", "tag"} {
		if v := r.URL.Query().Get(k); v != "" { filters[k] = v }
	}
	list, total, _ := h.User.List(page, size, filters, groupID)
	JSON(w, 200, map[string]any{"data": list, "total": total, "page": page})
}

func (h *Handler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	old, _ := h.User.GetByID(id)
	var u model.User
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		errJSON(w, 400, "请求格式错误"); return
	}
	u.ID = id
	h.User.Update(&u)
	// household_id 变化时同步 household_members
	oldHH := int64(0)
	if old != nil { oldHH = old.HouseholdID }
	if u.HouseholdID != oldHH {
		// 从旧户移除
		if oldHH > 0 {
			h.User.DB.Exec(`DELETE FROM household_members WHERE user_id=? AND household_id=?`, id, oldHH)
		}
		// 加入新户
		if u.HouseholdID > 0 {
			h.Household.AddMember(u.HouseholdID, id, "其他")
		}
	}
	JSON(w, 200, map[string]string{"ok": "updated"})
}

func (h *Handler) AdminResetPassword(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	h.User.UpdatePassword(id, middleware.HashPassword("123456"))
	JSON(w, 200, map[string]string{"ok": "已重置为123456"})
}

// ==================== 小组管理 ====================

func (h *Handler) ListGroups(w http.ResponseWriter, r *http.Request) {
	list, _ := h.Group.List()
	JSON(w, 200, map[string]any{"data": list})
}

func (h *Handler) CreateGroup(w http.ResponseWriter, r *http.Request) {
	var g model.Group
	json.NewDecoder(r.Body).Decode(&g)
	if g.Name == "" { errJSON(w, 400, "小组名称不能为空"); return }
	h.Group.Create(&g)
	JSON(w, 201, g)
}

func (h *Handler) UpdateGroup(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	var g model.Group
	json.NewDecoder(r.Body).Decode(&g)
	g.ID = id
	h.Group.Update(&g)
	JSON(w, 200, map[string]string{"ok": "updated"})
}

func (h *Handler) DeleteGroup(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	h.Group.Delete(id)
	JSON(w, 200, map[string]string{"ok": "deleted"})
}
