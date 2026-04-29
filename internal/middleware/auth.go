package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"village-system/internal/model"
)

type ctxKey string

const UserIDKey ctxKey = "user_id"
const UserRoleKey ctxKey = "user_role"

var JWTSecret = func() []byte {
	if s := os.Getenv("JWT_SECRET"); s != "" {
		return []byte(s)
	}
	log.Println("⚠️  未设置 JWT_SECRET 环境变量，使用随机密钥（重启后所有 token 失效）")
	b := make([]byte, 32)
	rand.Read(b)
	return []byte(hex.EncodeToString(b))
}()

// tokenClaims 与签发时的 JSON 字段一致；避免 MapClaims 类型断言导致 panic。
type tokenClaims struct {
	UID  int64  `json:"uid"`
	Role string `json:"role"`
	jwt.RegisteredClaims
}

func HashPassword(password string) string {
	h, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(h)
}

func CheckPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

func GenerateToken(userID int64, role string) (string, error) {
	claims := tokenClaims{
		UID:  userID,
		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, &claims).SignedString(JWTSecret)
}

func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		if token == "" {
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}
		var tc tokenClaims
		parsed, err := jwt.ParseWithClaims(token, &tc, func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return JWTSecret, nil
		})
		if err != nil || !parsed.Valid {
			http.Error(w, `{"error":"invalid token"}`, http.StatusUnauthorized)
			return
		}
		if tc.UID <= 0 || tc.Role == "" {
			http.Error(w, `{"error":"invalid token"}`, http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), UserIDKey, tc.UID)
		ctx = context.WithValue(ctx, UserRoleKey, tc.Role)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireRole 要求最低角色等级
func RequireRole(minRole string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role, _ := r.Context().Value(UserRoleKey).(string)
		if !model.HasRole(role, minRole) {
			http.Error(w, `{"error":"权限不足"}`, http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// AdminOnly 保留兼容
func AdminOnly(next http.Handler) http.Handler {
	return RequireRole("secretary", next)
}

// ReadOnlyGuard 驻村干部和网格员只允许 GET 请求（只读）
func ReadOnlyGuard(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			role, _ := r.Context().Value(UserRoleKey).(string)
			if model.IsReadOnly(role) {
				http.Error(w, `{"error":"当前角色为只读权限，不能执行此操作"}`, http.StatusForbidden)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}
