package request

import (
	appuser "github.com/iiimomoniii/backend-challenge/internal/application/user"
	domainuser "github.com/iiimomoniii/backend-challenge/internal/domain/user"
)

// RegisterRequest คือ request สำหรับ register request จาก HTTP request
// JSON ที่เข้ามาจะถูกแปลงเป็น Struct นี้
type RegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// ToUseCase แปลง HTTP request
// เป็น CreateRequest ของ application/usecase layer
// เพื่อแยก transport model ออกจาก application model
// ทำให้ use case ไม่ต้องผูกกับ HTTP หรือ JSON
func (r RegisterRequest) ToUseCase() appuser.CreateRequest {
	return appuser.CreateRequest{
		Name:     r.Name,
		Email:    r.Email,
		Password: r.Password,
	}
}

// UpdateRequest คือ request สำหรับ update user จาก HTTP request
// JSON ที่เข้ามาจะถูกแปลงเป็น Struct นี้
type UpdateRequest struct {
	Name  *string `json:"name,omitempty"`
	Email *string `json:"email,omitempty"`
}

// ToUseCase แปลง HTTP request
// เป็น UpdateRequest ของ domain layer
// เพื่อแยก transport model ออกจาก domain model
// ทำให้ domain layer ไม่ต้องผูกกับ HTTP หรือ JSON
func (r UpdateRequest) ToUseCase() domainuser.UpdateRequest {
	return domainuser.UpdateRequest{
		Name:  r.Name,
		Email: r.Email,
	}
}