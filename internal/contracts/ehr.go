package contracts

import (
	"github.com/ali-gulzar/speechory-core/internal/constants"
	"github.com/segmentio/ksuid"
)

type CreateEHRRequest struct {
	LocationID ksuid.KSUID            `json:"location_id" binding:"required"`
	Type       constants.EHR          `json:"type" binding:"required"`
	EHR        constants.NexHealthEHR `json:"ehr" binding:"required"`

	Subdomain *string `json:"subdomain"`

	Name    *string `json:"name"`
	Email   *string `json:"email"`
	ZipCode *string `json:"zip_code"`
	Website *string `json:"website"`
}

type EHR struct {
	OnboardingID  string `json:"onboarding_id"`
	OnboardingURL string `json:"onboarding_url"`
}

type GetEHRResponse struct {
	Data []constants.NexHealthEHR `json:"data"`
}
