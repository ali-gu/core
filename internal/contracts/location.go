package contracts

import (
	"time"

	"github.com/ali-gulzar/speechory-core/internal/constants"
	"github.com/ali-gulzar/speechory-core/internal/storage/states"
	"github.com/segmentio/ksuid"
)

type CreateLocationRequest struct {
	Address string                 `json:"address" binding:"required"`
	EHR     constants.NexHealthEHR `json:"ehr" binding:"required"`
}

type CreateLocationResponse struct {
	Location Location `json:"location"`
	EHR      EHR      `json:"ehr"`
}

type LocationURI struct {
	LocationID ksuid.KSUID `uri:"location_id,parser=encoding.TextUnmarshaler"`
}

type Location struct {
	ID        ksuid.KSUID          `json:"id"`
	Status    states.LocationState `json:"status"`
	Address   string               `json:"address"`
	CreatedAt time.Time            `json:"created_at"`
	EHR       *EHR                 `json:"ehr,omitempty"`
}

type GetLocationsResponse struct {
	Data []Location `json:"data"`
}

type UpdateLocationRequest struct {
}

type UpdateLocationResponse struct {
	ID        ksuid.KSUID          `json:"id"`
	Status    states.LocationState `json:"status"`
	Address   string               `json:"address"`
	CreatedAt time.Time            `json:"created_at"`
}
