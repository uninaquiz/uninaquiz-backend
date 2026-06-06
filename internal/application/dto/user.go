package dto

import "time"

type UserResponse struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type CreateUserResponse struct {
	User  UserResponse `json:"user"`
	Token string       `json:"token"`
}

type GetAllUsersResponse struct {
	Data       []UserResponse `json:"data"`
	Total      int64          `json:"total"`
	Page       int            `json:"page"`
	Limit      int            `json:"limit"`
	TotalPages int            `json:"totalPages"`
}

type LoginResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}
