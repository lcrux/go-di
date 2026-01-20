package models

import "time"

type Todo struct {
	ID        uint32    `json:"id"`
	Title     string    `json:"title"`
	Done      bool      `json:"done" default:"false"`
	DueDate   time.Time `json:"due_date,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

type CreateTodoRequest struct {
	Title   string    `json:"title" validate:"required"`
	Done    bool      `json:"done" default:"false"`
	DueDate time.Time `json:"due_date,omitempty"`
}

type CloseTodoRequest struct {
	ID uint32 `json:"id" validate:"required"`
}
