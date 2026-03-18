// services/mountcore/workflow/middleware.go

package workflow

import (
	"net/http"
)

// CORSMiddleware 实现CORS中间件
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 设置允许的来源
		w.Header().Set("Access-Control-Allow-Origin", "*")
		
		// 设置允许的HTTP方法
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, PATCH")
		
		// 设置允许的请求头
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		
		// 设置是否允许发送Cookie
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		
		// 处理OPTIONS请求
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		// 调用下一个处理函数
		next.ServeHTTP(w, r)
	})
}
