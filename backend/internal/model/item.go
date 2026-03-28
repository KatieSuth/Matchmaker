package model

import "time"

type Item struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateItemRequest struct {
	Name        string `json:"name"        binding:"required,min=1,max=120"`
	Description string `json:"description" binding:"max=500"`
}

type UpdateItemRequest struct {
	Name        string `json:"name"        binding:"omitempty,min=1,max=120"`
	Description string `json:"description" binding:"omitempty,max=500"`
}
