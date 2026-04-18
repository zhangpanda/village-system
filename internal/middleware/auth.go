package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
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

func HashPassword(password string) string {
	h, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(h)
}

func CheckPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

func GenerateToken(userID int64, role string) (string, error) {
	claims := jwt.MapClaims{
		"uid":  userID,
		"role": role,
		"exp":  time.Now().Add(24 * time.Hour).Unix(),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(JWTSecret)
}

func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		if token == "" {
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}
		parsed, err := jwt.Parse(token, func(t *jwt.Token) (any, error) {
			return JWTSecret, nil
		})
		if err != nil || !parsed.Valid {
			http.Error(w, `{"error":"invalid token"}`, http.StatusUnauthorized)
			return
		}
		claims := parsed.Claims.(jwt.MapClaims)
		uid := int64(claims["uid"].(float64))
		role := claims["role"].(string)
		ctx := context.WithValue(r.Context(), UserIDKey, uid)
		ctx = context.WithValue(ctx, UserRoleKey, role)
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
