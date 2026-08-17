package auth

import (
	"context"
	"errors"
	"strings"

	domainauth "github.com/iiimomoniii/backend-challenge/internal/domain/auth"
	domainuser "github.com/iiimomoniii/backend-challenge/internal/domain/user"
)

// LoginRequest คือข้อมูลที่รับเข้ามาสำหรับการlogin ของ user
// เป็น input model ของ LoginUseCase
type LoginRequest struct {
	Email    string
	Password string
}

// LoginResponse เป็น reponse กรณี Login สำเร็จ
// ประกอบด้วย access token และข้อมูล user
type LoginResponse struct {
	Token string
	User  *domainuser.User
}

// LoginUseCase ทำหน้าที่เป็น use case service
// สำหรับการlogin ของ user

// ใช้ UserRepository สำหรับค้นหาข้อมูล user
// ใช้ PasswordHasher สำหรับตรวจสอบ password
// ใช้ TokenProvider สำหรับสร้าง authentication token
type LoginUseCase struct {
	users  domainuser.Repository
	hasher domainauth.PasswordHasher
	tokens domainauth.TokenProvider
}

// NewLoginUseCase สร้าง LoginUseCase
// โดยรับ Repository, PasswordHasher และ TokenProvider ที่ต้องการใช้งานเข้ามา
func NewLoginUseCase(users domainuser.Repository, hasher domainauth.PasswordHasher, tokens domainauth.TokenProvider) *LoginUseCase {
	return &LoginUseCase{users: users, hasher: hasher, tokens: tokens}
}

//ทำการ login และreturn authentication token  พร้อมกับข้อมูล user เมื่อ login ผ่าน
func (uc *LoginUseCase) Execute(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
	//ตัด space
	email := strings.TrimSpace(req.Email)
	// ตรวจสอบว่ามี Email และ Password ค่าใดค่าหนึ่งว่างจะ return error
	if email == "" || req.Password == "" {
		return nil, domainuser.ErrInvalidCredentials
	}

	// ค้นหา user จาก email ผ่าน Repository Interface
	u, err := uc.users.SearchByEmail(ctx, email)
	if err != nil {
		// กรณีไม่เจอ user ให้คืน Invalid Credentials
		// แทนการบอกตรงๆว่ามี email นี้มีอยู่ในระบบแล้ว
		if errors.Is(err, domainuser.ErrUserNotFound) {
			return nil, domainuser.ErrInvalidCredentials
		}
		return nil, err
	}

	// ตรวจสอบ password ที่ีuserส่งมา
	// กับ password hash ใน database
	if err := uc.hasher.Verify(u.Password, req.Password); err != nil {
		return nil, domainuser.ErrInvalidCredentials
	}

	// สร้าง authentication token
	// โดยส่งข้อมูลที่จำเป็นเช่น id กับ email  
    // ไปใส่ไว้ใน Token Payload
	token, err := uc.tokens.Generate(ctx, domainauth.AuthPayload{
		UserID: u.ID,
		Email:  u.Email,
	})
	if err != nil {
		return nil, err
	}

	// login สำเร็จ return Token และข้อมูล user กลับไป
	return &LoginResponse{Token: token, User: u}, nil
}