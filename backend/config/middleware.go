package config

import (
	"net/http"
)

// Security headers middleware
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Strict-Transport-Security", "max-age=31536000")
		next.ServeHTTP(w, r)
	})
}

// Application headers middleware
func AppHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-API-Version", "v1.0")
		w.Header().Set("X-Powered-By", "SNSMS")
		next.ServeHTTP(w, r)
	})
}

// Apply middlewares for public routes (no JWT required)
func ApplyPublicMiddlewares(handler http.Handler) http.Handler {
	return SecurityHeaders(
		AppHeaders(
			CORS(handler),
		),
	)
}

// Apply middlewares for protected routes (JWT required)
func ApplyProtectedMiddlewares(handler http.Handler) http.Handler {
	return SecurityHeaders(
		AppHeaders(
			CORS(
				JWTMiddleware(handler),
			),
		),
	)
}

// Apply middlewares for admin-only routes
func ApplyAdminMiddlewares(handler http.Handler) http.Handler {
	return SecurityHeaders(
		AppHeaders(
			CORS(
				JWTMiddleware(
					AdminOnly(handler),
				),
			),
		),
	)
}
