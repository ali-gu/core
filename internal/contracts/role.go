package contracts

import (
	"time"

	"github.com/ali-gulzar/speechory-core/internal/constants"
	"github.com/segmentio/ksuid"
)

type CreateRoleRequest struct {
	Type constants.RoleType `json:"type" binding:"required"`
}

type CreateRoleResponse struct {
	ID        ksuid.KSUID        `json:"id"`
	Type      constants.RoleType `json:"type"`
	CreatedAt time.Time          `json:"created_at"`
}
