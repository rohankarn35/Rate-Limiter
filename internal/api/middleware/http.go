package middleware

// func RateLimiterMiddleware(l limiter.Limiter, next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		if !l.Allow() {
// 			http.Error(w, "Too Many Request", http.StatusTooManyRequests)
// 			return
// 		}
// 		next.ServeHTTP(w, r)
// 	})
// }
