package models

import (
	"time"
)

type Order struct {
	ID          int
	UserID      int
	ProductName string
	Amount      float64
	CreatedAt   time.Time
}
