package model

import (
	"time"
)

// User represents a user in the system
// swagger:model User
type User struct {
	// The ID of the user
	// required: true
	ID int64 `db:"id" json:"id"`
	// The email of the user
	// required: true
	Email string `db:"email" json:"email"`
	// The name of the user
	// required: true
	Name string `db:"name" json:"name"`
	// The password of the user (not returned in JSON)
	Password string `db:"password" json:"-"`
	// Creation time
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	// Last update time
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// DTOs (Data Transfer Objects)

// CreateUserRequest represents the payload for creating a user
// swagger:model CreateUserRequest
type CreateUserRequest struct {
	// The email address
	// required: true
	Email string `json:"email" binding:"required,email" example:"user@gmail.com"`
	// The user's name
	// required: true
	// min length: 3
	Name string `json:"name" binding:"required,min=3,max=100" example:"user"`
	// The password
	// required: true
	// min length: 6
	Password string `json:"password" binding:"required,min=6" example:"password"`
}

// UpdateUserRequest represents the payload for updating a user
// swagger:model UpdateUserRequest
type UpdateUserRequest struct {
	// The new email address
	Email string `json:"email" binding:"omitempty,email" example:"user@gmail.com"`
	// The new name
	// min length: 3
	Name string `json:"name" binding:"omitempty,min=3,max=100" example:"user"`
}

// UserResponse represents the user response data
// swagger:model UserResponse
type UserResponse struct {
	// The user ID
	ID int64 `json:"id" example:"1"`
	// The user email
	Email string `json:"email" example:"user@gmail.com"`
	// The user name
	Name string `json:"name" example:"user"`
	// Creation time
	CreatedAt time.Time `json:"created_at" example:"2025-12-06T17:16:43+07:00"`
	// Last update time
	UpdatedAt time.Time `json:"updated_at" example:"2025-12-06T17:16:43+07:00"`
}

func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:        u.ID,
		Email:     u.Email,
		Name:      u.Name,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}
