package contracts

import (
	"time"

	"github.com/ali-gulzar/speechory-core/internal/storage/states"
	"github.com/segmentio/ksuid"
)

type EntityStateCounts map[string]int

type AgentAnalytics struct {
	Rank                        int               `json:"rank"`
	AgentID                     ksuid.KSUID       `json:"agent_id"`
	Name                        string            `json:"name"`
	Status                      states.AgentState `json:"status"`
	Location                    *AgentLocation    `json:"location"`
	PhoneNumber                 *AgentPhoneNumber `json:"phone_number"`
	ConversationCount           int               `json:"conversation_count"`
	BookingsMade                int               `json:"bookings_made"`
	BookingConversionRate       float64           `json:"booking_conversion_rate"`
	LongestConversationSeconds  float64           `json:"longest_conversation_seconds"`
	ShortestConversationSeconds float64           `json:"shortest_conversation_seconds"`
	LastConversationAt          *time.Time        `json:"last_conversation_at"`
}

type AgentsAnalyticsSummary struct {
	TotalAgents           int               `json:"total_agents"`
	ByState               EntityStateCounts `json:"by_state"`
	TotalConversations    int               `json:"total_conversations"`
	TotalBookingsMade     int               `json:"total_bookings_made"`
	OverallConversionRate float64           `json:"overall_conversion_rate"`
	TopAgentID            *ksuid.KSUID      `json:"top_agent_id"`
	TopAgentName          *string           `json:"top_agent_name"`
}

type GetAgentsAnalyticsResponse struct {
	Summary AgentsAnalyticsSummary `json:"summary"`
	Data    []AgentAnalytics       `json:"data"`
}

type LocationsAnalyticsSummary struct {
	TotalLocations int               `json:"total_locations"`
	ByState        EntityStateCounts `json:"by_state"`
}

type GetLocationsAnalyticsResponse struct {
	Summary LocationsAnalyticsSummary `json:"summary"`
	Data    []Location                `json:"data"`
}

type PhoneNumbersAnalyticsSummary struct {
	TotalPhoneNumbers int               `json:"total_phone_numbers"`
	ByState           EntityStateCounts `json:"by_state"`
}

type GetPhoneNumbersAnalyticsResponse struct {
	Summary PhoneNumbersAnalyticsSummary `json:"summary"`
	Data    []PhoneNumber                `json:"data"`
}
