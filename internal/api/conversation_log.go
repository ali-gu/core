package api

import (
	"fmt"
	"net/http"

	"github.com/ali-gulzar/speechory-core/internal/contracts"
	"github.com/ali-gulzar/speechory-core/internal/storage"
	"github.com/ali-gulzar/speechory-core/pkg/rerror"
	"github.com/gin-gonic/gin"
	"github.com/team-telnyx/telnyx-go/v4/lib"
)

type LogConversationHandler struct {
	*API

	req contracts.LogConversationRequest
}

func (h *LogConversationHandler) Validator(ctx *gin.Context, api *API, body []byte) error {
	h.API = api

	if err := lib.VerifyWebhookSignature(body, ctx.Request.Header, api.Config.Telnyx.PublicKey); err != nil {
		return rerror.New(fmt.Errorf("%w: %v", ErrInvalidSignature, err)).WithKind(rerror.Permission)
	}

	return bindErr(ctx.ShouldBindJSON(&h.req))
}

func (h *LogConversationHandler) Handle(ctx *gin.Context, _ storage.DB) error {
	db, err := h.DBMux.BeginWrite()
	if err != nil {
		return rerror.Wrap(err)
	}

	if err := h.Biz.Conversation.LogConversation(ctx, db, h.req); err != nil {
		return rerror.Wrap(err)
	}

	ctx.Status(http.StatusNoContent)
	return nil
}
