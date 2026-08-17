package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	domainauth "github.com/iiimomoniii/backend-challenge/internal/domain/auth"
	domainuser "github.com/iiimomoniii/backend-challenge/internal/domain/user"
	"github.com/iiimomoniii/backend-challenge/internal/interface/http/response"
	"github.com/iiimomoniii/backend-challenge/pkg/code"
)

//contextKey
type contextKey string

//Request Context
const (
	userIDContextKey contextKey = "auth_user_id"
	emailContextKey  contextKey = "auth_email"
)

// Auth ทำหน้าที่จัดการ authentication middleware
// โดยใช้ TokenProvider สำหรับตรวจสอบ authentication token
type Auth struct {
	tokens domainauth.TokenProvider
}

// NewAuth สร้าง auth middleware
// โดยรับ TokenProvider ที่ใช้ตรวจสอบ Token เข้ามา
func NewAuth(tokens domainauth.TokenProvider) *Auth {
	return &Auth{tokens: tokens}
}

// ตรวจสอบว่าเป็น Bearer Token
// verify token ผ่าน TokenProvider
// และเก็บข้อมูล user ลงใน Request Context
func (a *Auth) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// อ่าน authorization header จาก HTTP request
		header := r.Header.Get("Authorization")
		// ตัด "Bearer " ออกจาก authorization header
		// เพื่อให้ได้เฉพาะ token
		token, ok := strings.CutPrefix(header, "Bearer ")
	    // ถ้าไม่มี Bearer Token หรือ token เป็นค่าว่าง
		// ให้คืน unauthorized กลับไปทันที
		if !ok || token == "" {
			writeUnauthorized(w, r)
			return
		}

		// ตรวจสอบ token ผ่าน TokenProvider
		payload, err := a.tokens.Verify(r.Context(), token)
		if err != nil {
			writeUnauthorized(w, r)
			return
		}
		// นำข้อมูล User จาก Token Payload
		// ไปเก็บไว้ใน Request Context
		ctx := context.WithValue(r.Context(), userIDContextKey, payload.UserID)
		ctx = context.WithValue(ctx, emailContextKey, payload.Email)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// UserIDFromContext ดึง user ID
// ที่ถูกเก็บไว้ใน Request Context
func UserIDFromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(userIDContextKey).(string)
	return id, ok
}

// EmailFromContext ดึง email
// ที่ถูกเก็บไว้ใน Request Context
func EmailFromContext(ctx context.Context) (string, bool) {
	email, ok := ctx.Value(emailContextKey).(string)
	return email, ok
}

// writeUnauthorized สร้าง HTTP Response
// สำหรับกรณีที่ authentication ไม่ผ่าน
func writeUnauthorized(w http.ResponseWriter, r *http.Request) {
	lang := code.ParseLang(r.Header.Get("Accept-Language"))
	errCode := domainuser.CodeInvalidCredentials

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(response.StatusForCode(errCode))
	_ = json.NewEncoder(w).Encode(response.FromErrorCode(errCode, lang))
}