package user

import (
	"context"
	"strings"

	domainuser "github.com/iiimomoniii/backend-challenge/internal/domain/user"
)

// UpdateUseCase ทำหน้าที่เป็น use case service
// สำหรับการ update user 
type UpdateUseCase struct {
	repo domainuser.Repository
}

// NewUpdateUseCase สร้าง UpdateUseCase
// โดยรับ Repository ที่ใช้สำหรับแก้ไขข้อมูล user เข้ามา
func NewUpdateUseCase(repo domainuser.Repository) *UpdateUseCase {
	return &UpdateUseCase{repo: repo}
}

//แก้ไขข้อมูล User ตาม ID
func (uc *UpdateUseCase) Execute(ctx context.Context, id string, req domainuser.UpdateRequest) (*domainuser.User, error) {
	//validate
	if id == "" {
		return nil, domainuser.ErrUserNotFound
	}

	if req.Name != nil {
		//ตัด space
		trimmed := strings.TrimSpace(*req.Name)

		// Name ต้องไม่เป็นค่าว่างหลังตัด space
		if trimmed == "" {
			return nil, domainuser.ErrNameRequired
		}
		req.Name = &trimmed
	}

	if req.Email != nil {
		//ตัด space
		trimmed := strings.TrimSpace(*req.Email)
		// Email ต้องไม่เป็นค่าว่างหลังตัด space
		if trimmed == "" {
			return nil, domainuser.ErrUsernameRequired
		}
		req.Email = &trimmed

		// ตรวจสอบว่า Email ใหม่ถูกใช้งานโดย user คนอื่นยัง
		// ถ้ามีแล้วแจ้ง error กลับไป
		if existing, err := uc.repo.SearchByEmail(ctx, trimmed); err == nil && existing != nil && existing.ID != id {
			return nil, domainuser.ErrUsernameAlreadyExists
		}
	}
	//ถ้าผ่าน validate หมดทำการ update
	return uc.repo.Update(ctx, id, req)
}