package user

import (
	"context"

	domainuser "github.com/iiimomoniii/backend-challenge/internal/domain/user"
)

// SearchUseCase ทำหน้าที่เป็น use case service
// สำหรับการค้นหา user จาก id
type SearchUseCase struct {
	repo domainuser.Repository
}

// NewSearchUseCase สร้าง SearchUseCase
// โดยรับ Repository ทีที่ต้องการใช้สำหรับค้นหา user เข้ามา
func NewSearchUseCase(repo domainuser.Repository) *SearchUseCase {
	return &SearchUseCase{repo: repo}
}

//ค้นหาข้อมูล user ตาม ID
func (uc *SearchUseCase) Execute(ctx context.Context, id string) (*domainuser.User, error) {
	//validate
	if id == "" {
		return nil, domainuser.ErrUserNotFound
	}
	return uc.repo.SearchByID(ctx, id)
}

// SearchAllUseCase ทำหน้าที่เป็น use case service
// สำหรับการค้นหา user ทั้งหมด
type SearchAllUseCase struct {
	repo domainuser.Repository
}

// NewSearchAllUseCase สร้าง SearchAllUseCase
// โดยรับ Repository ที่ใช้สำหรับดึงข้อมูล user ทั้งหมด
func NewSearchAllUseCase(repo domainuser.Repository) *SearchAllUseCase {
	return &SearchAllUseCase{
		repo: repo,
	}
}

//ค้นหาข้อมูลทั้งหมด
func (uc *SearchAllUseCase) Execute(ctx context.Context,) ([]*domainuser.User, error) {

	// ดึงข้อมูล user ทั้งหมด
	users, err := uc.repo.List(ctx)
	if err != nil {
		return nil, err
	}

	// กรณีไม่เจอข้อมูล จะ return เป็น []
	if users == nil {
		users = []*domainuser.User{}
	}

	return users, nil
}