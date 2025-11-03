package domain

import (
	"time"

	"github.com/google/uuid"
)

type SubscriptionDTO struct {
	ServiceName string    `json:"service_name" example:"Yandex Plus"`
	Price       int       `json:"price"        example:"400"` // rub
	UserID      uuid.UUID `json:"user_id"    example:"60601fee-2bf1-4721-ae6f-7636e79a0cba"`
	StartDate   string    `json:"start_date"   example:"07-2025"`
	EndDate     *string   `json:"end_date,omitempty" example:"09-2025"`
}

type SubscriptionResponse struct {
	ID          uuid.UUID `json:"id"`
	ServiceName string    `json:"service_name"`
	Price       int       `json:"price"`
	UserID      uuid.UUID `json:"user_id"`
	StartDate   string    `json:"start_date"`
	EndDate     *string   `json:"end_date,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Subscription struct {
	ID          uuid.UUID  `db:"id"`
	ServiceName string     `db:"service_name"`
	Price       int        `db:"price"`
	UserID      uuid.UUID  `db:"user_id"`
	StartMonth  time.Time  `db:"start_month"`
	EndMonth    *time.Time `db:"end_month"`
	CreatedAt   time.Time  `db:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at"`
}
