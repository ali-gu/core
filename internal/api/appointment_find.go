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

type FindAppointmentHandler struct {
	*API

	req contracts.FindAppointmentRequest
}

func (h *FindAppointmentHandler) Validator(ctx *gin.Context, api *API, body []byte) error {
	h.API = api

	if err := lib.VerifyWebhookSignature(body, ctx.Request.Header, api.Config.Telnyx.PublicKey); err != nil {
		return rerror.New(fmt.Errorf("%w: %v", ErrInvalidSignature, err)).WithKind(rerror.Permission)
	}

	return bindErr(ctx.ShouldBindQuery(&h.req))
}

func (h *FindAppointmentHandler) Handle(ctx *gin.Context, db storage.DB) error {
	appointments, err := h.Biz.Appointment.Find(ctx, db, h.req)
	if err != nil {
		return rerror.Wrap(err)
	}

	data := make([]contracts.Appointment, len(appointments))
	for i, a := range appointments {
		data[i] = contracts.Appointment{
			Time:        a.Time,
			DoctorsName: a.ProviderName,
			DoctorID:    a.ProviderID,
		}
	}

	ctx.JSON(http.StatusOK, contracts.FindAppointmentsResponse{Data: data})
	return nil
}
