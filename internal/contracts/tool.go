package contracts

import (
	"time"

	"github.com/ali-gulzar/speechory-core/internal/constants"
	"github.com/ali-gulzar/speechory-core/internal/storage/states"
	"github.com/segmentio/ksuid"
)

type Tool struct {
	ID        ksuid.KSUID        `json:"id"`
	Status    states.ToolState   `json:"status"`
	Type      constants.ToolType `json:"type"`
	Kind      constants.ToolKind `json:"kind"`
	Config    map[string]any     `json:"config"`
	CreatedAt time.Time          `json:"created_at"`
	UpdatedAt *time.Time         `json:"updated_at,omitempty"`
}

type GetToolsResponse struct {
	Data []Tool `json:"data"`
}
