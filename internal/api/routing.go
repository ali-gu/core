package api

import (
	"github.com/ali-gulzar/speechory-core/internal"
	"github.com/ali-gulzar/speechory-core/internal/biz"
	"github.com/ali-gulzar/speechory-core/internal/constants"
	"github.com/ali-gulzar/speechory-core/internal/middleware"
	"github.com/ali-gulzar/speechory-core/internal/storage"
	"github.com/ali-gulzar/speechory-core/pkg/rlog"

	"github.com/gin-gonic/gin"
)

type API struct {
	Biz     *biz.Biz
	Config  internal.Config
	DBMux   storage.IDBMux
	Logger  rlog.ILogger
	Storage storage.Storage
}

func (api *API) Routes() *gin.Engine {
	router := gin.New()
	router.Use(middleware.Logger(api.Logger))
	router.Use(middleware.CORS())
	router.Use(middleware.HTTPRequestLogger(api.DBMux, api.Storage.HTTPRequest))
	router.Use(middleware.ErrorHandler())
	router.GET("/health", api.HealthHandler)

	v1router := router.Group("/v1")

	v1router.POST("/users", api.Build(func() Handler {
		return &SignUpHandler{}
	}))
	v1router.POST("/users/signin", api.Build(func() Handler {
		return &SignInHandler{}
	}))
	v1router.POST("/users/refresh", api.Build(func() Handler {
		return &RefreshHandler{}
	}))

	webhookRouter := v1router.Group("/webhooks")
	webhookRouter.GET("/appointments/find", api.BuildWebhook(func() WebhookHandler {
		return &FindAppointmentHandler{}
	}))
	webhookRouter.POST("/appointments/book", api.BuildWebhook(func() WebhookHandler {
		return &BookAppointmentHandler{}
	}))
	webhookRouter.POST("/appointments/cancel", api.BuildWebhook(func() WebhookHandler {
		return &CancelAppointmentHandler{}
	}))
	webhookRouter.POST("/conversations/log", api.BuildWebhook(func() WebhookHandler {
		return &LogConversationHandler{}
	}))

	authed := v1router.Group("", middleware.Auth(api.DBMux, api.Biz.User))

	authed.GET("/users/me", api.Build(func() Handler {
		return &GetUserHandler{}
	}))
	authed.POST("/users/invite", api.Build(func() Handler {
		return &InviteUserHandler{}
	}))
	authed.POST("/users/signout", api.Build(func() Handler {
		return &SignOutHandler{}
	}))

	authed.GET("/practices/:practice_id", api.Build(func() Handler {
		return &GetPracticeHandler{}
	}))
	authed.PATCH("/practices/:practice_id", api.Build(func() Handler {
		return &UpdatePracticeHandler{}
	}))

	authed.POST("/agents", api.Build(func() Handler {
		return &CreateAgentHandler{}
	}))
	authed.GET("/agents", api.Build(func() Handler {
		return &GetAgentsHandler{}
	}))
	authed.PATCH("/agents/:agent_id", api.Build(func() Handler {
		return &UpdateAgentHandler{}
	}))
	authed.PATCH("/agents/:agent_id/activate", api.Build(func() Handler {
		return &ActivateAgentHandler{}
	}))
	authed.DELETE("/agents/:agent_id", api.Build(func() Handler {
		return &DeleteAgentHandler{}
	}))
	authed.PATCH("/agents/:agent_id/disable", api.Build(func() Handler {
		return &DisableAgentHandler{}
	}))
	authed.GET("/agents/:agent_id/conversations", api.Build(func() Handler {
		return &GetAgentConversationsHandler{}
	}))
	authed.GET("/agents/:agent_id/conversations/:conversation_id", api.Build(func() Handler {
		return &GetAgentConversationHandler{}
	}))
	authed.POST("/agents/test-agent", api.Build(func() Handler {
		return &CreateTestAgentHandler{}
	}))

	authed.POST("/locations", api.Build(func() Handler {
		return &CreateLocationHandler{}
	}))
	authed.GET("/locations", api.Build(func() Handler {
		return &GetLocationsHandler{}
	}))
	authed.PATCH("/locations/:location_id", api.Build(func() Handler {
		return &UpdateLocationHandler{}
	}))
	authed.DELETE("/locations/:location_id", api.Build(func() Handler {
		return &DeleteLocationHandler{}
	}))
	authed.GET("/locations/:location_id", api.Build(func() Handler {
		return &GetLocationHandler{}
	}))

	authed.GET("/ehrs", api.Build(func() Handler {
		return &GetEHRHandler{}
	}))

	authed.POST("/phone-numbers/reserve", api.Build(func() Handler {
		return &ReservePhoneNumberHandler{}
	}))
	authed.GET("/phone-numbers", api.Build(func() Handler {
		return &GetPhoneNumbersHandler{}
	}))
	authed.GET("/phone-numbers/available", api.Build(func() Handler {
		return &ListAvailablePhoneNumbersHandler{}
	}))
	authed.DELETE("/phone-numbers/:phone_number_id", api.Build(func() Handler {
		return &DeletePhoneNumberHandler{}
	}))
	authed.PATCH("/phone-numbers/:phone_number_id/disable", api.Build(func() Handler {
		return &DisablePhoneNumberHandler{}
	}))

	authed.GET("/analytics/agents", api.Build(func() Handler {
		return &GetAgentsAnalyticsHandler{}
	}))
	authed.GET("/analytics/locations", api.Build(func() Handler {
		return &GetLocationsAnalyticsHandler{}
	}))
	authed.GET("/analytics/phone-numbers", api.Build(func() Handler {
		return &GetPhoneNumbersAnalyticsHandler{}
	}))

	adminRouter := v1router.Group("/admin",
		middleware.Auth(api.DBMux, api.Biz.User),
		middleware.RequireRole(constants.RoleTypeSuperAdmin),
	)
	adminRouter.POST("/practices", api.Build(func() Handler {
		return &CreatePracticeHandler{}
	}))
	adminRouter.POST("/phone-numbers/activate", api.Build(func() Handler {
		return &ActivatePhoneNumberHandler{}
	}))
	adminRouter.GET("/phone-numbers/purchased", api.Build(func() Handler {
		return &ListPurchasedPhoneNumbersHandler{}
	}))
	adminRouter.POST("/roles", api.Build(func() Handler {
		return &CreateRoleHandler{}
	}))
	adminRouter.POST("/tools", api.Build(func() Handler {
		return &CreateToolHandler{}
	}))

	return router
}
