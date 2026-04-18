package middleware

import (
	"context"
	"net/http"

	"village-system/internal/model"
)

// Scope 数据范围上下文键
const ScopeGroupIDKey ctxKey = "scope_group_id"

// DataScope 行级权限中间件
// 小组长只能看自己组的数据，网格员和会计以上看全部
func DataScope(getGroupID func(int64) int64, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role, _ := r.Context().Value(UserRoleKey).(string)
		uid, _ := r.Context().Value(UserIDKey).(int64)

		var scopeGroupID int64
		// 只有小组长受小组数据隔离限制
		if !model.HasRole(role, "accountant") && !model.HasExactRole(role, "grid_worker") {
			scopeGroupID = getGroupID(uid)
		}
		// accountant 以上或网格员 scopeGroupID=0 表示不限制

		ctx := context.WithValue(r.Context(), ScopeGroupIDKey, scopeGroupID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetScopeGroupID 从 context 获取数据范围限制的 group_id
// 返回 0 表示不限制（可看全部）
func GetScopeGroupID(r *http.Request) int64 {
	if v := r.Context().Value(ScopeGroupIDKey); v != nil {
		return v.(int64)
	}
	return 0
}
