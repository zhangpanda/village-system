package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"village-system/internal/handler"
	"village-system/internal/middleware"
	"village-system/internal/store"
)

func main() {
	addr := flag.String("addr", ":8080", "listen address")
	dbPath := flag.String("db", "village.db", "sqlite database path")
	flag.Parse()

	villageName := os.Getenv("VILLAGE_NAME")
	if villageName == "" {
		villageName = "东兴堡村"
	}

	db, err := store.InitDB(*dbPath)
	if err != nil {
		log.Fatal("数据库初始化失败:", err)
	}
	defer db.Close()

	h := &handler.Handler{
		User:        &store.UserStore{DB: db},
		Notice:      &store.NoticeStore{DB: db},
		Finance:     &store.FinanceStore{DB: db},
		Subsidy:     &store.SubsidyStore{DB: db},
		Ticket:      &store.TicketStore{DB: db},
		Group:       &store.GroupStore{DB: db},
		Household:   &store.HouseholdStore{DB: db},
		Workflow:    &store.WorkflowStore{DB: db},
		WorkflowDef: &store.WorkflowDefStore{DB: db},
		Notify:      &store.NotificationStore{DB: db},
		Report:      &store.ReportStore{DB: db},
		VillageName: villageName,
	}

	mux := http.NewServeMux()

	// === 健康检查 ===
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	// === 公开接口 ===
	mux.HandleFunc("GET /api/config", h.SiteConfig)
	mux.HandleFunc("POST /api/login", h.Login)
	mux.HandleFunc("POST /api/register", h.Register)
	mux.HandleFunc("POST /api/wx/login", h.WxLogin)
	mux.HandleFunc("GET /api/notices", h.ListNotices)
	mux.HandleFunc("GET /api/notices/{id}", h.GetNotice)
	mux.HandleFunc("GET /api/finance", h.ListFinance)
	mux.HandleFunc("GET /api/finance/summary", h.FinanceSummary)
	mux.HandleFunc("GET /api/groups", h.ListGroups)

	// === 登录用户接口 ===
	authed := http.NewServeMux()
	authed.HandleFunc("GET /api/me", h.GetProfile)
	authed.HandleFunc("PUT /api/me", h.UpdateProfile)
	authed.HandleFunc("POST /api/me/bindphone", h.BindPhone)
	authed.HandleFunc("POST /api/me/password", h.ChangePassword)
	authed.HandleFunc("POST /api/reset-password", h.ResetPassword)
	authed.HandleFunc("POST /api/wx/phone", h.WxGetPhone)
	authed.HandleFunc("POST /api/upload", h.Upload)
	// 通知
	authed.HandleFunc("GET /api/notifications", h.ListNotifications)
	authed.HandleFunc("GET /api/notifications/unread-count", h.UnreadNotificationCount)
	authed.HandleFunc("PUT /api/notifications/{id}/read", h.MarkNotificationRead)
	authed.HandleFunc("POST /api/notifications/read-all", h.MarkAllNotificationsRead)
	// 补贴
	authed.HandleFunc("GET /api/subsidies", h.ListSubsidies)
	authed.HandleFunc("GET /api/subsidies/{id}", h.GetSubsidy)
	authed.HandleFunc("POST /api/subsidies", h.CreateSubsidy)
	// 工单
	authed.HandleFunc("GET /api/tickets", h.ListTickets)
	authed.HandleFunc("GET /api/tickets/{id}", h.GetTicket)
	authed.HandleFunc("POST /api/tickets", h.CreateTicket)
	authed.HandleFunc("POST /api/tickets/{id}/comments", h.AddTicketComment)
	// 工单状态（村民可关闭自己的）
	authed.HandleFunc("PUT /api/tickets/{id}/status", h.UpdateTicketStatus)

	mux.Handle("/api/me", middleware.Auth(authed))
	mux.Handle("/api/me/", middleware.Auth(authed))
	mux.Handle("/api/upload", middleware.Auth(authed))
	mux.Handle("/api/wx/phone", middleware.Auth(authed))
	mux.Handle("/api/notifications", middleware.Auth(authed))
	mux.Handle("/api/notifications/", middleware.Auth(authed))
	mux.Handle("/api/subsidies", middleware.Auth(authed))
	mux.Handle("/api/subsidies/", middleware.Auth(authed))
	mux.Handle("/api/tickets", middleware.Auth(authed))
	mux.Handle("/api/tickets/", middleware.Auth(authed))

	// === 管理接口（村委委员以上） ===
	admin := http.NewServeMux()
	admin.HandleFunc("GET /api/admin/dashboard", h.Dashboard)
	// 用户管理
	admin.HandleFunc("GET /api/admin/users", h.ListUsers)
	admin.HandleFunc("PUT /api/admin/users/{id}", h.UpdateUser)
	admin.HandleFunc("POST /api/admin/users/{id}/reset-password", h.AdminResetPassword)
	// 小组管理
	admin.HandleFunc("POST /api/admin/groups", h.CreateGroup)
	admin.HandleFunc("PUT /api/admin/groups/{id}", h.UpdateGroup)
	admin.HandleFunc("DELETE /api/admin/groups/{id}", h.DeleteGroup)
	// 户籍管理
	admin.HandleFunc("GET /api/admin/households", h.ListHouseholds)
	admin.HandleFunc("GET /api/admin/households/{id}", h.GetHousehold)
	admin.HandleFunc("POST /api/admin/households", h.CreateHousehold)
	admin.HandleFunc("PUT /api/admin/households/{id}", h.UpdateHousehold)
	admin.HandleFunc("DELETE /api/admin/households/{id}", h.DeleteHousehold)
	admin.HandleFunc("POST /api/admin/households/{id}/members", h.AddHouseholdMember)
	admin.HandleFunc("DELETE /api/admin/households/{id}/members/{member_id}", h.RemoveHouseholdMember)
	admin.HandleFunc("PUT /api/admin/households/{id}/members/{member_id}", h.UpdateHouseholdMember)
	// 公告管理
	admin.HandleFunc("GET /api/admin/notices", h.AdminListNotices)
	admin.HandleFunc("POST /api/admin/notices", h.CreateNotice)
	admin.HandleFunc("PUT /api/admin/notices/{id}", h.UpdateNotice)
	admin.HandleFunc("DELETE /api/admin/notices/{id}", h.DeleteNotice)
	admin.HandleFunc("PUT /api/admin/notices/{id}/review", h.ReviewNotice)
	// 财务管理
	admin.HandleFunc("GET /api/admin/finance", h.AdminListFinance)
	admin.HandleFunc("POST /api/admin/finance", h.CreateFinance)
	admin.HandleFunc("PUT /api/admin/finance/{id}/review", h.ReviewFinance)
	admin.HandleFunc("DELETE /api/admin/finance/{id}", h.DeleteFinance)
	// 补贴审批
	admin.HandleFunc("PUT /api/admin/subsidies/{id}/committee-review", h.CommitteeReviewSubsidy)
	admin.HandleFunc("PUT /api/admin/subsidies/{id}/secretary-review", h.SecretaryReviewSubsidy)
	// 工单管理
	admin.HandleFunc("PUT /api/admin/tickets/{id}/assign", h.AssignTicket)
	admin.HandleFunc("PUT /api/admin/tickets/{id}/status", h.UpdateTicketStatus)
	// 数据导出
	admin.HandleFunc("GET /api/admin/export/users", h.ExportUsers)
	admin.HandleFunc("GET /api/admin/export/finance", h.ExportFinance)
	admin.HandleFunc("GET /api/admin/export/subsidies", h.ExportSubsidies)
	// 数据导入
	admin.HandleFunc("POST /api/admin/import/users", h.ImportUsers)
	admin.HandleFunc("GET /api/admin/import/template", h.ImportTemplate)
	// 工作流定义管理
	admin.HandleFunc("GET /api/admin/workflow-defs", h.ListWorkflowDefs)
	admin.HandleFunc("GET /api/admin/workflow-def", h.GetWorkflowDef)
	admin.HandleFunc("POST /api/admin/workflow-defs", h.SaveWorkflowDef)
	admin.HandleFunc("DELETE /api/admin/workflow-defs/{id}", h.DeleteWorkflowDef)
	admin.HandleFunc("POST /api/admin/workflow/apply", h.ApplyTransition)
	admin.HandleFunc("GET /api/admin/workflow-logs", h.ListWorkflowLogs)
	// 报表
	admin.HandleFunc("GET /api/admin/reports", h.ListReports)
	admin.HandleFunc("GET /api/admin/reports/{name}", h.RunReport)
	admin.HandleFunc("POST /api/admin/reports", h.SaveReport)
	admin.HandleFunc("DELETE /api/admin/reports/{id}", h.DeleteReport)
	// 打印
	admin.HandleFunc("GET /api/admin/print/subsidy/{id}", h.PrintSubsidy)
	admin.HandleFunc("GET /api/admin/print/finance", h.PrintFinance)
	admin.HandleFunc("GET /api/admin/print/report/{name}", h.PrintReport)
	admin.HandleFunc("GET /api/admin/print/roster", h.PrintRoster)

	mux.Handle("/api/admin/", middleware.Auth(middleware.RequireRole("grid_worker",
		middleware.ReadOnlyGuard(
		middleware.DataScope(func(uid int64) int64 {
			u, err := h.User.GetByID(uid)
			if err != nil { return 0 }
			return u.GroupID
		}, admin)))))

	// === 静态文件 ===
	mux.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("uploads"))))
	mux.Handle("/admin/", http.StripPrefix("/admin/", http.FileServer(http.Dir("web/admin"))))
	mux.Handle("/app/", http.StripPrefix("/app/", http.FileServer(http.Dir("web/public"))))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" { http.Redirect(w, r, "/app/", http.StatusFound); return }
		http.NotFound(w, r)
	})

	limited := middleware.RateLimit(120, cors(accessLog(mux)))

	srv := &http.Server{Addr: *addr, Handler: limited}

	log.Printf("🏘️  %s · 村务系统启动 %s", villageName, *addr)
	log.Printf("   村民端: http://localhost%s/app/", *addr)
	log.Printf("   管理端: http://localhost%s/admin/", *addr)

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Println("正在关闭服务...")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		srv.Shutdown(ctx)
	}()

	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatal(err)
	}
	log.Println("服务已停止")
}

// cors 默认仅允许与请求 Host 同源的 Origin；生产请设置 CORS_ALLOWED_ORIGINS（逗号分隔，可含 * 仅用于开发）。
func cors(next http.Handler) http.Handler {
	raw := os.Getenv("CORS_ALLOWED_ORIGINS")
	if raw == "" {
		raw = os.Getenv("CORS_ORIGIN")
	}
	var explicit []string
	for _, o := range strings.Split(raw, ",") {
		o = strings.TrimSpace(o)
		if o != "" {
			explicit = append(explicit, o)
		}
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqOrigin := r.Header.Get("Origin")
		allow := ""

		if len(explicit) > 0 {
			for _, o := range explicit {
				if o == "*" {
					allow = "*"
					break
				}
				if o == reqOrigin {
					allow = reqOrigin
					break
				}
			}
		} else {
			scheme := "http"
			if r.TLS != nil {
				scheme = "https"
			}
			if xfp := r.Header.Get("X-Forwarded-Proto"); xfp == "https" || xfp == "http" {
				scheme = xfp
			}
			same := scheme + "://" + r.Host
			if reqOrigin == same || reqOrigin == "" {
				if reqOrigin != "" {
					allow = reqOrigin
				}
			}
		}

		if allow != "" {
			w.Header().Set("Access-Control-Allow-Origin", allow)
			w.Header().Set("Vary", "Origin")
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Authorization")
		if r.Method == "OPTIONS" {
			if reqOrigin != "" && allow == "" {
				w.WriteHeader(http.StatusForbidden)
				return
			}
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func accessLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			start := time.Now()
			rw := &statusWriter{ResponseWriter: w, status: 200}
			next.ServeHTTP(rw, r)
			log.Printf("%s %s %d %s", r.Method, r.URL.Path, rw.status, time.Since(start).Round(time.Millisecond))
		} else {
			next.ServeHTTP(w, r)
		}
	})
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}
