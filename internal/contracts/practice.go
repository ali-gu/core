package contracts

import (
	"time"

	"github.com/ali-gulzar/speechory-core/internal/storage/states"
	"github.com/segmentio/ksuid"
)

type CreatePracticeRequest struct {
	Name    string  `json:"name" binding:"required"`
	Email   *string `json:"email"`
	ZipCode *string `json:"zip_code"`
	Website *string `json:"website"`
}

type CreatePracticeResponse struct {
	ID        ksuid.KSUID `json:"id"`
	Name      string      `json:"name"`
	Email     *string     `json:"email"`
	ZipCode   *string     `json:"zip_code"`
	Website   *string     `json:"website"`
	CreatedAt time.Time   `json:"created_at"`
}

type PracticeURI struct {
	PracticeID ksuid.KSUID `uri:"practice_id,parser=encoding.TextUnmarshaler"`
}

type GetPracticeResponse struct {
	ID        ksuid.KSUID          `json:"id"`
	Name      string               `json:"name"`
	Email     *string              `json:"email"`
	ZipCode   *string              `json:"zip_code"`
	Website   *string              `json:"website"`
	Status    states.PracticeState `json:"status"`
	CreatedAt time.Time            `json:"created_at"`
}

type UpdatePracticeRequest struct {
	Email   *string `json:"email"`
	ZipCode *string `json:"zip_code"`
	Website *string `json:"website"`
}

type UpdatePracticeResponse struct {
	ID        ksuid.KSUID          `json:"id"`
	Name      string               `json:"name"`
	Email     *string              `json:"email"`
	ZipCode   *string              `json:"zip_code"`
	Website   *string              `json:"website"`
	Status    states.PracticeState `json:"status"`
	CreatedAt time.Time            `json:"created_at"`
}
