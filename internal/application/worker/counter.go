package worker

import (
	"context"
	"log"
	"time"

	domainuser "github.com/iiimomoniii/backend-challenge/internal/domain/user"
)

// Counter ทำหน้าที่เป็น Background Service
// สำหรับนับจำนวน user ที่เข้ามา
// ใช้ Repository ผ่าน domainuser.Repository interface
// ทำให้ Counter ไม่ผูกกับรายละเอียดของ Database
type Counter struct {
	repo     domainuser.Repository
	interval time.Duration
}

// NewCounter สร้าง Counter
// โดยรับ Repository และช่วงเวลาที่ต้องการให้ทำงาน
func NewCounter(repo domainuser.Repository, interval time.Duration) *Counter {
	return &Counter{
		repo:     repo,
		interval: interval,
	}
}

// Run เริ่มทำงาน Background Service
// โดยจะนับจำนวน user ตาม interval
// service จะทำงานต่อเนื่องจนกว่า Context จะถูก cancel
func (c *Counter) Run(ctx context.Context) {
	// สร้าง Ticker สำหรับกำหนดช่วงเวลาการทำงาน
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			// หยุดการทำงานเมื่อ Context ถูก cancel
			return

		case <-ticker.C:
			// เรียก Repository เพื่อหาจำนวน user ทั้งหมด
			count, err := c.repo.Count(ctx)
			if err != nil {
				// Log Error และรอทำงานรอบถัดไป
				log.Printf(
					"user counter: failed to count users: %v",
					err,
				)
				continue
			}

			// Log จำนวน user ทั้งหมด
			log.Printf(
				"user counter: total users = %d",
				count,
			)
		}
	}
}
