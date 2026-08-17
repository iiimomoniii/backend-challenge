package user

import (
	"context"

	domainuser "github.com/iiimomoniii/backend-challenge/internal/domain/user"
)

// DeleteUseCase ทำหน้าที่เป็น use case service
// สำหรับการdelete user โดยใช้ id
type DeleteUseCase struct {
	repo domainuser.Repository
}

// NewDeleteUseCase สร้าง DeleteUseCase
// โดยรับ Repository ที่ใช้สำหรับลบ user เข้ามา
func NewDeleteUseCase(repo domainuser.Repository) *DeleteUseCase {
	return &DeleteUseCase{repo: repo}
}

//ลบ User ตาม ID
func (uc *DeleteUseCase) Execute(ctx context.Context, id string) error {
	//validate
	if id == "" {
		return domainuser.ErrUserNotFound
	}
	return uc.repo.Delete(ctx, id)
}