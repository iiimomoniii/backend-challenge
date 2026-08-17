package user

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"

	domainauth "github.com/iiimomoniii/backend-challenge/internal/domain/auth"
	domainuser "github.com/iiimomoniii/backend-challenge/internal/domain/user"
)

// CreateRequest คือข้อมูลที่รับเข้ามาสำหรับการสร้าง user
// เป็น input model ของ CreateUseCase
type CreateRequest struct {
	Name     string
	Email    string
	Password string
}

// CreateUseCase ทำหน้าที่เป็น use case service
// สำหรับการสร้าง user ใหม่
//
// ใช้ Repository ผ่าน domainuser.Repository interface
// และใช้ PasswordHasher interface สำหรับ hash password
// ทำให้ use case ไม่ผูกกับ implementation ของ password hasher
type CreateUseCase struct {
	repo   domainuser.Repository
	hasher domainauth.PasswordHasher
}

// NewCreateUseCase สร้าง CreateUseCase
// โดยรับ Repository และ PasswordHasher ที่ต้องการใช้งานเข้ามา
func NewCreateUseCase(repo domainuser.Repository, hasher domainauth.PasswordHasher) *CreateUseCase {
	return &CreateUseCase{repo: repo, hasher: hasher}
}

func (uc *CreateUseCase) Execute(ctx context.Context, req CreateRequest) (*domainuser.User, error) {
	//ตัดช่องว่าง
	name := strings.TrimSpace(req.Name)
	email := strings.TrimSpace(req.Email)

	//validate CreateRequest
	switch {
	case email == "":
		return nil, domainuser.ErrUsernameRequired
	case name == "":
		return nil, domainuser.ErrNameRequired
	case req.Password == "":
		return nil, domainuser.ErrPasswordRequired
	case len(req.Password) < 6:
		return nil, domainuser.ErrPasswordTooShort
	}

	// ตรวจสอบว่า email ถูกใช้งานไปแล้วหรือยัง 
	// ป้องกันการมี email ซ้ำในระบบ
	existing, err := uc.repo.SearchByEmail(ctx, email)
	if err != nil && !errors.Is(err, domainuser.ErrUserNotFound) {
		return nil, err
	}
	if existing != nil {
		return nil, domainuser.ErrUsernameAlreadyExists
	}

	// hash password ก่อนนำไปsaveใน database
	hashed, err := uc.hasher.Hash(req.Password)
	if err != nil {
		return nil, err
	}

	//ปั้น User Object ขึ้นมา
	u := &domainuser.User{
		ID:        uuid.NewString(),
		Name:      name,
		Email:     email,
		Password:  hashed,
		CreatedAt: time.Now().UTC(),
	}

	//สร้าง domain user ใหม่
	if err := uc.repo.Create(ctx, u); err != nil {
		return nil, err
	}

	return u, nil
}