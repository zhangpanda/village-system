package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"

	"village-system/internal/middleware"
	"village-system/internal/model"
	"village-system/internal/store"
)

func main() {
	villageName := os.Getenv("VILLAGE_NAME")
	if villageName == "" {
		villageName = "东兴堡村"
	}

	db, err := store.InitDB("village.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	us := &store.UserStore{DB: db}
	gs := &store.GroupStore{DB: db}
	hs := &store.HouseholdStore{DB: db}
	ns := &store.NoticeStore{DB: db}
	fs := &store.FinanceStore{DB: db}
	ss := &store.SubsidyStore{DB: db}
	ts := &store.TicketStore{DB: db}

	// === 村民小组 ===
	groups := []model.Group{
		{Name: villageName + "1组"},
		{Name: villageName + "2组"},
		{Name: villageName + "3组"},
		{Name: villageName + "4组"},
		{Name: villageName + "5组"},
	}
	for i := range groups {
		gs.Create(&groups[i])
	}
	fmt.Printf("✅ 创建 %d 个村民小组\n", len(groups))

	// === 用户（完整村两委班子 + 村民）===
	pwd := middleware.HashPassword("123456")
	users := []model.User{
		{Account: "admin", Phone: "admin", Name: "张超", Gender: "male", BirthDate: "1975-03-15", Ethnicity: "汉族", Education: "高中", MaritalStatus: "married", Role: "secretary", Position: "party_secretary", IDCard: "510xxx19750315xxxx", Address: villageName + "1组18号", GroupID: 1, IsPartyMember: true, PasswordHash: pwd},
		{Phone: "13800000001", Name: "李秀英", Gender: "female", BirthDate: "1980-06-22", Ethnicity: "汉族", Education: "初中", MaritalStatus: "married", Role: "committee", Position: "women_director", IDCard: "510xxx19800622xxxx", Address: villageName + "1组22号", GroupID: 1, IsPartyMember: true, PasswordHash: pwd},
		{Phone: "13800000002", Name: "张大山", Gender: "male", BirthDate: "1985-01-10", Ethnicity: "汉族", Education: "初中", MaritalStatus: "married", Role: "villager", Position: "villager", IDCard: "510xxx19850110xxxx", Address: villageName + "2组5号", GroupID: 2, PasswordHash: pwd},
		{Phone: "13800000003", Name: "刘翠花", Gender: "female", BirthDate: "1990-04-18", Ethnicity: "汉族", Education: "高中", MaritalStatus: "married", Role: "group_leader", Position: "group_leader", IDCard: "510xxx19900418xxxx", Address: villageName + "2组12号", GroupID: 2, IsPartyMember: true, PasswordHash: pwd},
		{Phone: "13800000004", Name: "陈志强", Gender: "male", BirthDate: "1978-08-05", Ethnicity: "汉族", Education: "初中", MaritalStatus: "married", Role: "villager", Position: "villager", IDCard: "510xxx19780805xxxx", Address: villageName + "3组8号", GroupID: 3, IsMilitary: true, PasswordHash: pwd},
		{Phone: "13800000005", Name: "赵小燕", Gender: "female", BirthDate: "1995-02-20", Ethnicity: "汉族", Education: "大专", MaritalStatus: "married", Role: "group_leader", Position: "group_leader", IDCard: "510xxx19950220xxxx", Address: villageName + "3组15号", GroupID: 3, PasswordHash: pwd},
		{Phone: "13800000006", Name: "孙德福", Gender: "male", BirthDate: "1970-09-12", Ethnicity: "汉族", Education: "小学", MaritalStatus: "widowed", Role: "villager", Position: "villager", IDCard: "510xxx19700912xxxx", Address: villageName + "4组3号", GroupID: 4, IsLowIncome: true, PasswordHash: pwd},
		{Phone: "13800000007", Name: "周美玲", Gender: "female", BirthDate: "1988-03-03", Ethnicity: "汉族", Education: "大专", MaritalStatus: "married", Role: "committee", Position: "accountant", IDCard: "510xxx19880303xxxx", Address: villageName + "4组20号", GroupID: 4, IsPartyMember: true, PasswordHash: pwd},
		{Phone: "13800000008", Name: "吴国庆", Gender: "male", BirthDate: "1982-07-16", Ethnicity: "汉族", Education: "高中", MaritalStatus: "married", Role: "supervisor", Position: "supervisor_director", IDCard: "510xxx19820716xxxx", Address: villageName + "5组6号", GroupID: 5, IsPartyMember: true, PasswordHash: pwd},
		{Phone: "13800000009", Name: "郑小明", Gender: "male", BirthDate: "2000-01-01", Ethnicity: "汉族", Education: "本科", MaritalStatus: "unmarried", Role: "villager", Position: "villager", IDCard: "510xxx20000101xxxx", Address: villageName + "5组11号", GroupID: 5, PasswordHash: pwd},
		{Phone: "13800000010", Name: "王建国", Gender: "male", BirthDate: "1972-10-01", Ethnicity: "汉族", Education: "高中", MaritalStatus: "married", Role: "deputy", Position: "deputy_secretary", IDCard: "510xxx19721001xxxx", Address: villageName + "1组30号", GroupID: 1, IsPartyMember: true, PasswordHash: pwd},
		{Phone: "13800000011", Name: "马兰花", Gender: "female", BirthDate: "1968-05-08", Ethnicity: "汉族", Education: "文盲", MaritalStatus: "widowed", Role: "villager", Position: "villager", IDCard: "510xxx19680508xxxx", Address: villageName + "3组2号", GroupID: 3, IsFiveGuarantee: true, PasswordHash: pwd},
	}
	for i := range users {
		if users[i].Position == "" { users[i].Position = "villager" }
		us.Create(&users[i])
	}
	fmt.Printf("✅ 创建 %d 个用户 (密码统一: 123456)\n", len(users))

	// === 户籍 ===
	households := []model.Household{
		{HouseholdNo: "001", HeadID: 1, Address: villageName + "1组18号", GroupID: 1, FarmlandArea: 5.2, ForestArea: 2.0, HomeSiteArea: 180},
		{HouseholdNo: "002", HeadID: 3, Address: villageName + "2组5号", GroupID: 2, FarmlandArea: 12.7, ForestArea: 3.5, HomeSiteArea: 200},
		{HouseholdNo: "003", HeadID: 5, Address: villageName + "3组8号", GroupID: 3, FarmlandArea: 6.0, ForestArea: 1.8, HomeSiteArea: 150},
		{HouseholdNo: "004", HeadID: 7, Address: villageName + "4组3号", GroupID: 4, FarmlandArea: 2.0, ForestArea: 0.5, HomeSiteArea: 120},
		{HouseholdNo: "005", HeadID: 9, Address: villageName + "5组6号", GroupID: 5, FarmlandArea: 7.0, ForestArea: 4.0, HomeSiteArea: 220},
	}
	for i := range households {
		hs.Create(&households[i])
	}
	// 家庭成员
	hs.AddMember(1, 1, "户主")
	hs.AddMember(1, 2, "配偶")
	hs.AddMember(2, 3, "户主")
	hs.AddMember(2, 4, "配偶")
	hs.AddMember(3, 5, "户主")
	hs.AddMember(3, 6, "配偶")
	hs.AddMember(4, 7, "户主")
	hs.AddMember(5, 9, "户主")
	hs.AddMember(5, 10, "子女")
	fmt.Printf("✅ 创建 %d 户家庭\n", len(households))

	// === 公告 ===
	notices := []model.Notice{
		{Title: "关于2026年春耕补贴发放的通知", Content: "各位村民，2026年春耕补贴将于4月15日开始发放，请携带身份证及银行卡到村委会办理登记。\n\n补贴标准：\n- 水稻种植：每亩200元\n- 玉米种植：每亩150元\n- 油菜种植：每亩120元\n\n办理时间：工作日 9:00-17:00\n联系电话：村委会 0838-xxxxxxx", Category: "policy", AuthorID: 1, Pinned: true, Attachments: "[]"},
		{Title: "【紧急】明日暴雨预警 请做好防范", Content: "据县气象台预报，4月10日将有大到暴雨，局部地区可能出现短时强降水。\n\n请各户注意：\n1. 低洼地带做好排水准备\n2. 检查房屋是否漏雨\n3. 山边住户注意山体滑坡\n4. 有险情请立即拨打村委电话\n\n值班电话：张超 138xxxx0000", Category: "urgent", AuthorID: 1, Pinned: true, Attachments: "[]"},
		{Title: "村道硬化工程完工通知", Content: "经过两个月施工，从村口到李家湾的3.2公里村道硬化工程已全部完工，即日起正式通车。\n\n工程总投资28.5万元（其中上级补助20万元，村集体自筹8.5万元），由县交通局验收合格。\n\n感谢各位村民在施工期间的理解与支持！", Category: "policy", AuthorID: 1, Pinned: false, Attachments: "[]"},
		{Title: "清明节文明祭扫倡议书", Content: "各位村民：\n\n清明节将至，为保护生态环境，倡导文明祭扫：\n- 提倡鲜花祭扫、网络祭扫\n- 严禁在林区及路边焚烧纸钱\n- 严禁燃放鞭炮\n- 共同守护绿水青山\n\n违反规定者将按照《森林防火条例》处理。", Category: "policy", AuthorID: 2, Pinned: false, Attachments: "[]"},
		{Title: "种植合作社2025年度分红公告", Content: "经合作社理事会审议，2025年度经营情况如下：\n\n总收入：486,000元\n总支出：312,000元\n净利润：174,000元\n\n按章程规定，提取公积金20%后，剩余按股分红。社员分红将于4月20日发放到各社员账户。\n\n详细分红明细已张贴在村委会公示栏。", Category: "policy", AuthorID: 1, Pinned: false, Attachments: "[]"},
		{Title: "村篮球场开放时间调整", Content: "因夏季来临，村篮球场开放时间调整为：\n\n每日 6:00 - 21:00\n灯光照明至21:00自动关闭\n\n请爱护公共设施，使用后请自觉清理场地。\n\n另：每周六晚7点有村民篮球友谊赛，欢迎参加！", Category: "activity", AuthorID: 2, Pinned: false, Attachments: "[]"},
		{Title: "农村医保缴费通知", Content: "2026年度城乡居民基本医疗保险开始缴费：\n\n缴费标准：每人380元/年\n缴费截止：2026年6月30日\n缴费方式：\n1. 微信搜索[社保缴费]小程序\n2. 到村委会代缴\n3. 银行柜台缴费\n\n未按时缴费将影响医保待遇，请及时办理。", Category: "policy", AuthorID: 1, Pinned: false, Attachments: "[]"},
		{Title: "端午节包粽子活动报名", Content: "一年一度的端午节即将到来！村委会组织包粽子活动：\n\n时间：5月28日上午9:00\n地点：村委会大院\n材料：村委会统一准备\n\n欢迎各家各户踊跃参加，名额有限（50户），先到先得！\n报名方式：到村委会登记或电话报名。", Category: "activity", AuthorID: 2, Pinned: false, Attachments: "[]"},
	}
	for i := range notices {
		notices[i].WorkflowState = "published"
		ns.Create(&notices[i])
		db.Exec(`UPDATE notices SET views=?, published_at=CURRENT_TIMESTAMP WHERE id=?`, rand.Intn(200)+10, notices[i].ID)
	}
	fmt.Printf("✅ 创建 %d 条公告\n", len(notices))

	// === 财务记录 ===
	finances := []model.FinanceRecord{
		{Type: "income", Amount: 50000000, Category: "上级拨款", Remark: "2026年度村级运转经费", Date: "2026-01-15", AuthorID: 1},
		{Type: "income", Amount: 12000000, Category: "上级拨款", Remark: "春耕补贴专项资金", Date: "2026-03-01", AuthorID: 1},
		{Type: "income", Amount: 3600000, Category: "集体收入", Remark: "鱼塘承包费（张大山）", Date: "2026-01-20", AuthorID: 1},
		{Type: "income", Amount: 1800000, Category: "集体收入", Remark: "村集体门面房租金（一季度）", Date: "2026-01-25", AuthorID: 1},
		{Type: "income", Amount: 2000000, Category: "上级拨款", Remark: "村道硬化工程补助款（尾款）", Date: "2026-02-20", AuthorID: 1},
		{Type: "income", Amount: 500000, Category: "其他收入", Remark: "废旧物资处置收入", Date: "2026-03-10", AuthorID: 1},
		{Type: "expense", Amount: 8500000, Category: "基础设施", Remark: "村道硬化工程款（第二期）", Date: "2026-01-18", AuthorID: 1},
		{Type: "expense", Amount: 1200000, Category: "基础设施", Remark: "村道路灯维修更换12盏", Date: "2026-02-10", AuthorID: 1},
		{Type: "expense", Amount: 2400000, Category: "人员支出", Remark: "保洁员工资（1-3月，2人×4000元/月）", Date: "2026-03-31", AuthorID: 1},
		{Type: "expense", Amount: 1500000, Category: "人员支出", Remark: "护林员补贴（1-3月）", Date: "2026-03-31", AuthorID: 1},
		{Type: "expense", Amount: 350000, Category: "办公费用", Remark: "村委会办公用品及打印耗材", Date: "2026-03-10", AuthorID: 1},
		{Type: "expense", Amount: 680000, Category: "公共服务", Remark: "自来水管网维修（3组片区）", Date: "2026-03-20", AuthorID: 1},
		{Type: "expense", Amount: 280000, Category: "办公费用", Remark: "村委会电费、网费（一季度）", Date: "2026-03-25", AuthorID: 1},
		{Type: "expense", Amount: 150000, Category: "公共服务", Remark: "清明节文明祭扫宣传物资", Date: "2026-04-01", AuthorID: 1},
		{Type: "expense", Amount: 960000, Category: "公共服务", Remark: "垃圾清运费（一季度）", Date: "2026-04-05", AuthorID: 1},
	}
	for i := range finances {
		finances[i].WorkflowState = "approved"
		fs.Create(&finances[i])
	}
	fmt.Printf("✅ 创建 %d 条财务记录\n", len(finances))

	// === 补贴申请 ===
	subsidies := []model.Subsidy{
		{Title: "2026年春耕补贴", Type: "farming", Amount: 240000, Reason: "种植水稻12亩", ApplicantID: 3},
		{Title: "2026年春耕补贴", Type: "farming", Amount: 150000, Reason: "种植玉米10亩", ApplicantID: 5},
		{Title: "大病医疗救助", Type: "medical", Amount: 500000, Reason: "母亲住院手术，医保报销后自费部分较大", ApplicantID: 4},
		{Title: "危房改造补贴", Type: "housing", Amount: 2000000, Reason: "房屋为D级危房，需要翻新重建", ApplicantID: 7},
		{Title: "子女教育补助", Type: "education", Amount: 300000, Reason: "两个孩子在县城读高中，家庭困难", ApplicantID: 6},
		{Title: "2026年春耕补贴", Type: "farming", Amount: 180000, Reason: "种植油菜15亩", ApplicantID: 8},
	}
	for i := range subsidies {
		subsidies[i].Attachments = "[]"
		ss.Create(&subsidies[i])
	}
	// 审批部分（村委初审 + 村支书终审）
	ss.CommitteeReview(1, 2, true, "符合补贴条件，已核实种植面积")
	ss.SecretaryReview(1, 1, true, "同意")
	ss.CommitteeReview(2, 2, true, "已核实")
	ss.SecretaryReview(2, 1, true, "同意")
	ss.CommitteeReview(3, 2, true, "情况属实，同意救助")
	ss.SecretaryReview(3, 1, true, "同意发放")
	ss.CommitteeReview(5, 2, false, "不符合教育补助条件，建议申请助学贷款")
	fmt.Printf("✅ 创建 %d 条补贴申请\n", len(subsidies))

	// === 工单 ===
	tickets := []model.Ticket{
		{Title: "3组路灯不亮", Content: "3组到4组路段有3盏路灯不亮了，晚上走路很不安全，请尽快维修。", Category: "repair", Priority: "urgent", SubmitterID: 5, Images: "[]"},
		{Title: "自来水浑浊", Content: "最近一周家里自来水发黄浑浊，烧开后有沉淀，不敢饮用。周围几户邻居也有同样问题。", Category: "complaint", Priority: "urgent", SubmitterID: 4, Images: "[]"},
		{Title: "申请开具户籍证明", Content: "孩子上学需要户籍证明，请问什么时候可以到村委会办理？需要带什么材料？", Category: "service", Priority: "normal", SubmitterID: 6, Images: "[]"},
		{Title: "垃圾桶满了没人收", Content: "2组村口的垃圾桶已经满了好几天了，垃圾都溢出来了，夏天容易滋生蚊虫，请安排清运。", Category: "complaint", Priority: "normal", SubmitterID: 3, Images: "[]"},
		{Title: "建议增设健身器材", Content: "村里老年人越来越多，建议在篮球场旁边增设一些适合老年人的健身器材，方便大家锻炼身体。", Category: "suggestion", Priority: "low", SubmitterID: 7, Images: "[]"},
		{Title: "邻居占用公共通道", Content: "5组李某在公共通道上堆放建材，影响通行，沟通多次未果，请村委会协调处理。", Category: "complaint", Priority: "normal", SubmitterID: 9, Images: "[]"},
		{Title: "申请使用村委会会议室", Content: "合作社想在下周六借用村委会会议室开社员大会，大约30人参加，请问是否可以？", Category: "service", Priority: "low", SubmitterID: 8, Images: "[]"},
	}
	for i := range tickets {
		ts.Create(&tickets[i])
	}
	// 处理部分工单
	ts.Assign(1, 1)
	ts.UpdateState(1, "processing", 1)
	ts.AddComment(&model.TicketComment{TicketID: 1, UserID: 1, Content: "已安排电工明天上午去检修，预计明天下午恢复。"})
	ts.Assign(2, 2)
	ts.UpdateState(2, "processing", 2)
	ts.AddComment(&model.TicketComment{TicketID: 2, UserID: 2, Content: "已联系水务站，初步判断是水源地管道老化导致，已申请维修。"})
	ts.Assign(3, 2)
	ts.UpdateState(3, "resolved", 2)
	ts.AddComment(&model.TicketComment{TicketID: 3, UserID: 2, Content: "请携带户口本和身份证，工作日到村委会即可办理，当场出具。"})
	ts.Assign(7, 1)
	ts.UpdateState(7, "resolved", 1)
	ts.AddComment(&model.TicketComment{TicketID: 7, UserID: 1, Content: "可以使用，请提前一天到村委会登记。"})
	fmt.Printf("✅ 创建 %d 条工单\n", len(tickets))

	fmt.Println("\n🎉 种子数据初始化完成！")
	fmt.Println("   管理员账号: admin / 123456")
	fmt.Println("   村民账号: 13800000002 / 123456 (张大山)")
}
