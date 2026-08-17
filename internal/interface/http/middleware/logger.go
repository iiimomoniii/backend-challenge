package middleware

import (
	"log"
	"net/http"
	"time"
)

// statusRecorder ทำหน้าที่เก็บ HTTP status code
// ที่ Handler ส่งกลับมา เพื่อให้ Logger สามารถนำไปบันทึกเป็น Log ได้
type statusRecorder struct {
	http.ResponseWriter
	status int
}

// WriteHeader Log ข้อมูล HTTP status code
// ก่อนส่ง Response Header กลับไปยัง client
func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

// Log ข้อมูล HTTP Method, URL Path, Status Code
// และระยะเวลาของ Request
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(rec, r)
		// แสดง Method, Path, Status Code
		// และระยะเวลาของ Request
		log.Printf("%s %s %d %s", r.Method, r.URL.Path, rec.status, time.Since(start))
	})
}