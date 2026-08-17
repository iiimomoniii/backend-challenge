package handler

import (
	"encoding/json"
	"net/http"

	appauth "github.com/iiimomoniii/backend-challenge/internal/application/auth"
	domainuser "github.com/iiimomoniii/backend-challenge/internal/domain/user"
	"github.com/iiimomoniii/backend-challenge/internal/interface/http/request"
	"github.com/iiimomoniii/backend-challenge/internal/interface/http/response"
)

// AuthHandler ทำหน้าที่จัดการ HTTP request/response
// สำหรับ authentication api
type AuthHandler struct {
	login *appauth.LoginUseCase
}

// NewAuthHandler สร้าง AuthHandler
// โดยรับ use case ที่เกี่ยวข้องกับ authentication เข้ามาใช้งาน
func NewAuthHandler(login *appauth.LoginUseCase) *AuthHandler {
	return &AuthHandler{login: login}
}

// loginResponseBody คือ response สำหรับ login
// ประกอบด้วย token และข้อมูล user
type loginResponseBody struct {
	Token string                 `json:"token"`
	User  response.UserResponse `json:"user"`
}

// Login รับ HTTP request สำหรับ login
// และเรียก LoginUseCase เพื่อทำการ authentication
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req request.LoginRequest
	// decode JSON จาก request body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, domainuser.CodeInvalidInput)
		return
	}

	// แปลง HTTP request เป็น use case input
	// แล้วส่งให้ LoginUseCase ทำการ authentication
	result, err := h.login.Execute(r.Context(), req.ToUseCase())
	if err != nil {
		writeDomainError(w, r, err)
		return
	}

	// สร้าง login response
	// โดยแปลง domain user เป็น HTTP response
	// และส่ง token กลับไปให้ client
	writeJSON(w, http.StatusOK, loginResponseBody{
		Token: result.Token,
		User:  response.FromUser(result.User),
	})
}