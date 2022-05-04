package router

import "net/http"

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	// NOTE: Access log is not being used in this case.
	w.WriteHeader(http.StatusNotFound)
}

func jsonContentTypeHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, DELETE")
		w.Header().Add("Content-Type", "application/json; charset=utf-8")
		next.ServeHTTP(w, r)
	})
}
