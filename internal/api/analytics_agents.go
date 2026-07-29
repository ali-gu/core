package api

import (
	"net/http"

	"github.com/ali-gulzar/speechory-core/internal/biz"
	"github.com/ali-gulzar/speechory-core/internal/contracts"
	"github.com/ali-gulzar/speechory-core/internal/storage"
	"github.com/ali-gulzar/speechory-core/pkg/ptr"
	"github.com/ali-gulzar/speechory-core/pkg/rerror"
	"github.com/gin-gonic/gin"
)

type GetAgentsAnalyticsHandler struct {
	*API

	user storage.User
}

func (h *GetAgentsAnalyticsHandler) IsWrite() bool {
	return false
}

func (h *GetAgentsAnalyticsHandler) Setup(_ *gin.Context, _ storage.DB, api *API) error {
	h.API = api
	return nil
}

func (h *GetAgentsAnalyticsHandler) Permissions(ctx *gin.Context, _ storage.DB) error {
	user, err := requireUser(ctx)
	if err != nil {
		return rerror.Wrap(err)
	}

	h.user = user
	return nil
}

func (h *GetAgentsAnalyticsHandler) Handle(ctx *gin.Context, db storage.DB) error {
	performances, err := h.Biz.Analytics.GetAgentsAnalytics(ctx, db, h.user.PracticeID)
	if err != nil {
		return rerror.Wrap(err)
	}

	data := make([]contracts.AgentAnalytics, len(performances))
	for i, performance := range performances {
		data[i] = mapAgentPerformance(i+1, performance)
	}

	ctx.JSON(http.StatusOK, contracts.GetAgentsAnalyticsResponse{
		Summary: buildAgentsAnalyticsSummary(performances),
		Data:    data,
	})
	return nil
}

func mapAgentPerformance(rank int, performance biz.AgentPerformance) contracts.AgentAnalytics {
	a := performance.Agent
	location := ptr.From(a.Location)
	phoneNumber := ptr.From(a.PhoneNumber)

	return contracts.AgentAnalytics{
		Rank:    rank,
		AgentID: a.ID,
		Name:    a.Name,
		Status:  a.EntityState,
		Location: &contracts.AgentLocation{
			ID:      ptr.ToPtrOrNil(location.ID),
			Address: ptr.ToPtrOrNil(location.Address),
		},
		PhoneNumber: &contracts.AgentPhoneNumber{
			ID:          ptr.ToPtrOrNil(phoneNumber.ID),
			PhoneNumber: ptr.ToPtrOrNil(phoneNumber.PhoneNumber),
		},
		ConversationCount:           performance.ConversationCount,
		BookingsMade:                performance.BookingsMade,
		BookingConversionRate:       performance.ConversionRate(),
		LongestConversationSeconds:  performance.LongestConversation.Seconds(),
		ShortestConversationSeconds: performance.ShortestConversation.Seconds(),
		LastConversationAt:          performance.LastConversationAt,
	}
}

func buildAgentsAnalyticsSummary(performances []biz.AgentPerformance) contracts.AgentsAnalyticsSummary {
	summary := contracts.AgentsAnalyticsSummary{
		TotalAgents: len(performances),
		ByState:     contracts.EntityStateCounts{},
	}

	var topAgent *biz.AgentPerformance
	for i, performance := range performances {
		summary.ByState[string(performance.Agent.EntityState)]++
		summary.TotalConversations += performance.ConversationCount
		summary.TotalBookingsMade += performance.BookingsMade

		if performance.BookingsMade > 0 && (topAgent == nil || performance.BookingsMade > topAgent.BookingsMade) {
			topAgent = &performances[i]
		}
	}

	if summary.TotalConversations > 0 {
		summary.OverallConversionRate = float64(summary.TotalBookingsMade) / float64(summary.TotalConversations)
	}
	if topAgent != nil {
		summary.TopAgentID = &topAgent.Agent.ID
		summary.TopAgentName = &topAgent.Agent.Name
	}

	return summary
}
