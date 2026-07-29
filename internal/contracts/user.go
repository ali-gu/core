package contracts

import (
	"time"

	"github.com/ali-gulzar/speechory-core/internal/storage/states"
	"github.com/segmentio/ksuid"
)

type SignUpRequest struct {
	Email        string  `json:"email" binding:"required,email"`
	Password     string  `json:"password" binding:"required"`
	PracticeName *string `json:"practice_name"`
}

type SignInRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type SignInResponse struct {
	GetUserResponse
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	ExpiresAt   int64  `json:"expires_at"`
}

type GetUserResponse struct {
	ID          ksuid.KSUID      `json:"id"`
	Email       string           `json:"email"`
	RoleID      ksuid.KSUID      `json:"role_id"`
	PracticeID  ksuid.KSUID      `json:"practice_id"`
	EntityState states.UserState `json:"entity_state"`
	CreatedAt   time.Time        `json:"created_at"`
	Practice    UserPractice     `json:"practice"`
}

type UserPractice struct {
	ID     ksuid.KSUID          `json:"id"`
	Name   string               `json:"name"`
	Status states.PracticeState `json:"status"`
}

type SignUpResponse struct {
	ID          ksuid.KSUID      `json:"id"`
	Email       string           `json:"email"`
	RoleID      ksuid.KSUID      `json:"role_id"`
	PracticeID  ksuid.KSUID      `json:"practice_id"`
	EntityState states.UserState `json:"entity_state"`
	CreatedAt   time.Time        `json:"created_at"`
}

type InviteUserRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type InviteUserResponse struct {
	ID          ksuid.KSUID      `json:"id"`
	Email       string           `json:"email"`
	RoleID      ksuid.KSUID      `json:"role_id"`
	PracticeID  ksuid.KSUID      `json:"practice_id"`
	EntityState states.UserState `json:"entity_state"`
	CreatedAt   time.Time        `json:"created_at"`
}
