package request

import (
	appauth "github.com/iiimomoniii/backend-challenge/internal/application/auth"
)

// LoginRequest คือ request สำหรับ login request จาก HTTP request
// JSON ที่เข้ามาจะถูกแปลงเป็น Struct นี้
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// ToUseCase แปลง HTTP request
// เป็น LoginRequest ของ application/usecase layer
// เพื่อแยก transport model ออกจาก application model
// ทำให้ use case ไม่ต้องผูกกับ HTTP หรือ JSON
func (r LoginRequest) ToUseCase() appauth.LoginRequest {
	return appauth.LoginRequest{
		Email:    r.Email,
		Password: r.Password,
	}
}