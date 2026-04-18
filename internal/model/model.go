package model

import (
	"strings"
	"time"
)

// ========== 角色与职务 ==========
// 参照中国农村"村两委"体制

// 系统角色（控制系统权限）
var RoleLevel = map[string]int{
	"admin":             99, // 系统管理员
	"secretary":         90, // 村党支部书记（一把手）
	"resident_official": 88, // 驻村第一书记/驻村干部（上级派驻，只读监督）
	"director":          85, // 村委会主任（可与书记一肩挑）
	"deputy":            70, // 副书记/副主任
	"supervisor":        65, // 村务监督委员会
	"committee":         60, // 两委委员（治保、妇女、宣传、组织、纪检等）
	"accountant":        50, // 村会计/出纳
	"group_leader":      40, // 村民小组长
	"grid_worker":       35, // 网格员（信息采集、矛盾排查、代办服务）
	"villager":          10, // 普通村民
}

var RoleLabel = map[string]string{
	"admin":             "系统管理员",
	"secretary":         "党支部书记",
	"resident_official": "驻村干部",
	"director":          "村委会主任",
	"deputy":            "副书记/副主任",
	"supervisor":        "监委会委员",
	"committee":         "两委委员",
	"accountant":        "村会计",
	"group_leader":      "村民小组长",
	"grid_worker":       "网格员",
	"villager":          "村民",
}

// 具体职务（展示用，一个人可以有角色+职务）
var PositionLabels = map[string]string{
	"party_secretary":      "党支部书记",
	"village_director":     "村委会主任",
	"resident_secretary":   "驻村第一书记",
	"resident_official":    "驻村干部",
	"deputy_secretary":     "副书记",
	"deputy_director":      "副主任",
	"org_committee":        "组织委员",
	"prop_committee":       "宣传委员",
	"discipline_committee": "纪检委员",
	"security_director":    "治保主任",
	"women_director":       "妇女主任",
	"militia_captain":      "民兵连长",
	"accountant":           "村会计",
	"cashier":              "出纳",
	"supervisor_director":  "监委会主任",
	"supervisor_member":    "监委会委员",
	"group_leader":         "村民小组长",
	"grid_worker":          "网格员",
	"villager":             "村民",
}

func HasRole(userRoles string, minRole string) bool {
	minLevel := RoleLevel[minRole]
	for _, r := range splitRoles(userRoles) {
		if RoleLevel[r] >= minLevel {
			return true
		}
	}
	return false
}

// HasExactRole 检查是否拥有某个具体角色
func HasExactRole(userRoles string, role string) bool {
	for _, r := range splitRoles(userRoles) {
		if r == role { return true }
	}
	return false
}

// IsReadOnly 驻村干部和网格员只有查看权限，不参与审批
func IsReadOnly(userRoles string) bool {
	roles := splitRoles(userRoles)
	hasReadOnlyRole := false
	for _, r := range roles {
		if r == "resident_official" || r == "grid_worker" {
			hasReadOnlyRole = true
		}
		// 如果同时有其他管理角色，则不限制
		if r != "resident_official" && r != "grid_worker" && r != "villager" && RoleLevel[r] >= RoleLevel["group_leader"] {
			return false
		}
	}
	return hasReadOnlyRole
}

func splitRoles(roles string) []string {
	var result []string
	for _, r := range strings.Split(roles, ",") {
		r = strings.TrimSpace(r)
		if r != "" {
			result = append(result, r)
		}
	}
	return result
}

func TopRole(userRoles string) string {
	best := ""
	bestLevel := -1
	for _, r := range splitRoles(userRoles) {
		if RoleLevel[r] > bestLevel {
			best = r
			bestLevel = RoleLevel[r]
		}
	}
	return best
}

func RoleLabels(userRoles string) string {
	parts := splitRoles(userRoles)
	labels := make([]string, 0, len(parts))
	for _, r := range parts {
		if l, ok := RoleLabel[r]; ok {
			labels = append(labels, l)
		}
	}
	return strings.Join(labels, "/")
}

// ========== 用户/村民 ==========

type User struct {
	ID           int64     `json:"id"`
	Account      string    `json:"account"`
	Phone        string    `json:"phone"`
	Name         string    `json:"name"`
	Gender       string    `json:"gender"`        // male/female
	BirthDate    string    `json:"birth_date"`    // 出生日期
	Ethnicity    string    `json:"ethnicity"`     // 民族
	Education    string    `json:"education"`     // 文化程度
	MaritalStatus string   `json:"marital_status"` // 婚姻状况
	Role         string    `json:"role"`          // 系统角色
	RoleLabel    string    `json:"role_label,omitempty"`
	Position     string    `json:"position"`      // 具体职务
	PositionLabel string   `json:"position_label,omitempty"`
	IDCard       string    `json:"id_card"`       // 身份证号
	Address      string    `json:"address"`       // 详细地址
	GroupID      int64     `json:"group_id"`      // 所属小组
	GroupName    string    `json:"group_name,omitempty"`
	HouseholdID  int64     `json:"household_id"`  // 户号
	// 特殊身份
	IsPartyMember bool    `json:"is_party_member"` // 是否党员
	IsLowIncome   bool    `json:"is_low_income"`   // 低保户
	IsFiveGuarantee bool  `json:"is_five_guarantee"`// 五保户
	IsDisabled    bool    `json:"is_disabled"`     // 残疾人
	IsMilitary    bool    `json:"is_military"`     // 军属/退役军人
	// 联系方式
	WechatID     string    `json:"wechat_id"`
	EmergencyContact string `json:"emergency_contact"` // 紧急联系人
	EmergencyPhone   string `json:"emergency_phone"`
	// 微信
	OpenID       string    `json:"openid,omitempty"`
	AvatarURL    string    `json:"avatar_url,omitempty"`
	Active       bool      `json:"active"`
	PasswordHash string    `json:"-"`
	Remark       string    `json:"remark"`        // 备注
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// ========== 家庭/户 ==========

type Household struct {
	ID          int64     `json:"id"`
	HouseholdNo string    `json:"household_no"`  // 户号
	HeadID      int64     `json:"head_id"`       // 户主ID
	HeadName    string    `json:"head_name,omitempty"`
	Address     string    `json:"address"`
	GroupID     int64     `json:"group_id"`
	GroupName   string    `json:"group_name,omitempty"`
	MemberCount int       `json:"member_count,omitempty"`
	FarmlandArea float64  `json:"farmland_area"` // 全户承包耕地（亩）
	ForestArea   float64  `json:"forest_area"`   // 全户林地（亩）
	HomeSiteArea float64  `json:"homesite_area"` // 宅基地面积（平方米）
	Remark      string    `json:"remark"`
	CreatedAt   time.Time `json:"created_at"`
}

type HouseholdMember struct {
	ID           int64  `json:"id"`
	HouseholdID  int64  `json:"household_id"`
	UserID       int64  `json:"user_id"`
	UserName     string `json:"user_name"`
	Relation     string `json:"relation"` // 户主/配偶/子女/父母/其他
}

// ========== 村民小组 ==========

type Group struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	LeaderID    int64     `json:"leader_id"`
	LeaderName  string    `json:"leader_name,omitempty"`
	MemberCount int       `json:"member_count,omitempty"`
	HouseholdCount int    `json:"household_count,omitempty"`
	FarmlandArea float64  `json:"farmland_area,omitempty"` // 小组总耕地
	CreatedAt   time.Time `json:"created_at"`
}

// ========== 公告 ==========
// 工作流: draft → pending_review → published / rejected

type Notice struct {
	ID            int64      `json:"id"`
	Title         string     `json:"title"`
	Content       string     `json:"content"`
	Category      string     `json:"category"`
	AuthorID      int64      `json:"author_id"`
	Author        string     `json:"author"`
	Pinned        bool       `json:"pinned"`
	Attachments   string     `json:"attachments"`
	Views         int        `json:"views"`
	WorkflowState string     `json:"workflow_state"`
	ReviewerID    int64      `json:"reviewer_id"`
	ReviewerName  string     `json:"reviewer_name,omitempty"`
	ReviewNote    string     `json:"review_note"`
	PublishedAt   *time.Time `json:"published_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// ========== 财务 ==========
// 工作流: draft → pending_review → approved / rejected

type FinanceRecord struct {
	ID            int64     `json:"id"`
	Type          string    `json:"type"`
	Amount        int64     `json:"amount"`
	Category      string    `json:"category"`
	Remark        string    `json:"remark"`
	Date          string    `json:"date"`
	Voucher       string    `json:"voucher"`
	AuthorID      int64     `json:"author_id"`
	Author        string    `json:"author"`
	WorkflowState string    `json:"workflow_state"`
	ReviewerID    int64     `json:"reviewer_id"`
	ReviewerName  string    `json:"reviewer_name,omitempty"`
	ReviewNote    string    `json:"review_note"`
	CreatedAt     time.Time `json:"created_at"`
}

type FinanceSummary struct {
	TotalIncome  int64             `json:"total_income"`
	TotalExpense int64             `json:"total_expense"`
	Balance      int64             `json:"balance"`
	ByCategory   []CategorySummary `json:"by_category,omitempty"`
}

type CategorySummary struct {
	Category string `json:"category"`
	Type     string `json:"type"`
	Amount   int64  `json:"amount"`
}

// ========== 补贴 ==========
// 工作流: submitted → committee_review → secretary_review → approved / rejected

type Subsidy struct {
	ID            int64      `json:"id"`
	Title         string     `json:"title"`
	Type          string     `json:"type"`
	Amount        int64      `json:"amount"`
	Reason        string     `json:"reason"`
	Attachments   string     `json:"attachments"`
	ApplicantID   int64      `json:"applicant_id"`
	Applicant     string     `json:"applicant"`
	WorkflowState string     `json:"workflow_state"`
	CurrentStep   int        `json:"current_step"`
	CommitteeID   int64      `json:"committee_id"`
	CommitteeName string     `json:"committee_name,omitempty"`
	CommitteeNote string     `json:"committee_note"`
	CommitteeAt   *time.Time `json:"committee_at,omitempty"`
	SecretaryID   int64      `json:"secretary_id"`
	SecretaryName string     `json:"secretary_name,omitempty"`
	SecretaryNote string     `json:"secretary_note"`
	SecretaryAt   *time.Time `json:"secretary_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}

// ========== 工单 ==========
// 工作流: open → assigned → processing → resolved → closed

type Ticket struct {
	ID            int64      `json:"id"`
	Title         string     `json:"title"`
	Content       string     `json:"content"`
	Category      string     `json:"category"`
	Images        string     `json:"images"`
	Priority      string     `json:"priority"`
	WorkflowState string     `json:"workflow_state"`
	SubmitterID   int64      `json:"submitter_id"`
	Submitter     string     `json:"submitter"`
	AssigneeID    int64      `json:"assignee_id"`
	Assignee      string     `json:"assignee"`
	ResolvedAt    *time.Time `json:"resolved_at,omitempty"`
	ClosedAt      *time.Time `json:"closed_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type TicketComment struct {
	ID        int64     `json:"id"`
	TicketID  int64     `json:"ticket_id"`
	UserID    int64     `json:"user_id"`
	UserName  string    `json:"user_name"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// ========== 工作流日志 ==========

type WorkflowLog struct {
	ID           int64     `json:"id"`
	DocType      string    `json:"doc_type"`
	DocID        int64     `json:"doc_id"`
	FromState    string    `json:"from_state"`
	ToState      string    `json:"to_state"`
	Action       string    `json:"action"`
	OperatorID   int64     `json:"operator_id"`
	OperatorName string    `json:"operator_name"`
	Note         string    `json:"note"`
	CreatedAt    time.Time `json:"created_at"`
	DocTitle     string    `json:"doc_title,omitempty"`
}

// ========== 看板 ==========

type Dashboard struct {
	NoticeCount      int             `json:"notice_count"`
	NoticePending    int             `json:"notice_pending"`
	TicketOpen       int             `json:"ticket_open"`
	TicketTotal      int             `json:"ticket_total"`
	SubsidyPending   int             `json:"subsidy_pending"`
	SubsidyTotal     int             `json:"subsidy_total"`
	FinancePending   int             `json:"finance_pending"`
	FinanceSummary   *FinanceSummary `json:"finance_summary"`
	UserCount        int             `json:"user_count"`
	GroupCount       int             `json:"group_count"`
	HouseholdCount   int             `json:"household_count"`
	PartyMemberCount int             `json:"party_member_count"`
}
