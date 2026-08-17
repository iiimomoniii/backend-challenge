package response

import (
	"net/http"
	"time"

	domainuser "github.com/iiimomoniii/backend-challenge/internal/domain/user"
	"github.com/iiimomoniii/backend-challenge/pkg/code"
)

// UserResponse คือ response สำหรับตอบกลับ user ไปยัง Client
// โดยแปลงจาก domain user ให้อยู่ในรูปแบบที่ใช้สำหรับ HTTP response
type UserResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

// FromUser แปลง domain user 1 user
// เป็น UserResponse สำหรับส่งกลับไปยัง client
//
// เพื่อแยก domain model ออกจาก response model
// ทำให้ domain model ไม่ต้องผูกกับ HTTP หรือ JSON
func FromUser(u *domainuser.User) UserResponse {
	return UserResponse{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		CreatedAt: u.CreatedAt,
	}
}

// FromUsers แปลงข้อมูล domain หลาย user
// เป็น UserResponse สำหรับส่งกลับไปยัง client
// กรณีที่ api ต้องส่งข้อมูลหลาย user
func FromUsers(users []*domainuser.User) []UserResponse {
	out := make([]UserResponse, 0, len(users))

	for _, u := range users {
		out = append(out, FromUser(u))
	}

	return out
}

// ErrorResponse คือ response สำหรับส่ง error กลับไปยัง client
// โดยประกอบด้วย ErrorCode และ Message
type ErrorResponse struct {
	ErrorCode string `json:"error_code"`
	Message   string `json:"message"`
}

// FromErrorCode สร้าง ErrorResponse จาก ErrorCode
// และ language ที่ต้องการใช้สำหรับแสดง Message
func FromErrorCode(errCode string, lang code.Lang) ErrorResponse {
	return ErrorResponse{
		ErrorCode: errCode,
		Message:   code.Message(errCode, lang),
	}
}

// StatusForCode แปลง ErrorCode
// เป็น HTTP status code สำหรับ response
// ใช้สำหรับ HTTP status code ตามประเภทของ Error
func StatusForCode(errCode string) int {
	switch errCode {
	case "USR001", "USR002", "USR003", "USR004":
		return http.StatusBadRequest //ข้อมูลที่ Client ส่งมาไม่ถูกต้อง
	case "USR005":
		return http.StatusConflict //ข้อมูลเกิดความขัดแย้ง เช่น Email ถูกใช้งานแล้ว
	case "USR006":
		return http.StatusNotFound //ไม่พบ User ที่ต้องการ
	case "USR007":
		return http.StatusUnauthorized // authentication ไม่ผ่าน
	default:
		return http.StatusBadRequest //ไม่รู้จัก Error Code ให้ใช้ Bad Request เป็นค่า default
	}
}
