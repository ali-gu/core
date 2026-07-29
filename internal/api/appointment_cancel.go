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

type CancelAppointmentHandler struct {
	*API

	req contracts.CancelAppointmentRequest
}

func (h *CancelAppointmentHandler) Validator(ctx *gin.Context, api *API, body []byte) error {
	h.API = api

	if err := lib.VerifyWebhookSignature(body, ctx.Request.Header, api.Config.Telnyx.PublicKey); err != nil {
		return rerror.New(fmt.Errorf("%w: %v", ErrInvalidSignature, err)).WithKind(rerror.Permission)
	}

	return bindErr(ctx.ShouldBindJSON(&h.req))
}

func (h *CancelAppointmentHandler) Handle(ctx *gin.Context, db storage.DB) error {
	cancelled, err := h.Biz.Appointment.Cancel(ctx, db, h.req)
	if err != nil {
		return rerror.Wrap(err)
	}

	ctx.JSON(http.StatusOK, contracts.CancelAppointmentResponse{
		Data: contracts.CancelledAppointment{
			AppointmentID: cancelled.ID,
			Time:          cancelled.Time,
			DoctorID:      cancelled.ProviderID,
		},
	})
	return nil
}
