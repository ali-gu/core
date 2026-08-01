package contracts

import (
	"time"

	"github.com/ali-gulzar/speechory-core/internal/storage/states"
	"github.com/segmentio/ksuid"
)

type CreateTestAgentResponse struct {
	ID          ksuid.KSUID       `json:"id"`
	Status      states.AgentState `json:"status"`
	Name        string            `json:"name"`
	AgentRef    *string           `json:"agent_ref"`
	Location    *AgentLocation    `json:"location"`
	PhoneNumber *AgentPhoneNumber `json:"phone_number"`
	CreatedAt   time.Time         `json:"created_at"`
}
