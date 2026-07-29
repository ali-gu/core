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

type BookAppointmentHandler struct {
	*API

	req contracts.BookAppointmentRequest
}

func (h *BookAppointmentHandler) Validator(ctx *gin.Context, api *API, body []byte) error {
	h.API = api

	if err := lib.VerifyWebhookSignature(body, ctx.Request.Header, api.Config.Telnyx.PublicKey); err != nil {
		return rerror.New(fmt.Errorf("%w: %v", ErrInvalidSignature, err)).WithKind(rerror.Permission)
	}

	return bindErr(ctx.ShouldBindJSON(&h.req))
}

func (h *BookAppointmentHandler) Handle(ctx *gin.Context, db storage.DB) error {
	booked, err := h.Biz.Appointment.Book(ctx, db, h.req)
	if err != nil {
		return rerror.Wrap(err)
	}

	ctx.JSON(http.StatusOK, contracts.BookAppointmentResponse{
		Data: contracts.BookedAppointment{
			AppointmentID: booked.ID,
			Time:          booked.Time,
			ProviderID:    booked.ProviderID,
			Confirmed:     booked.Confirmed,
		},
	})
	return nil
}
