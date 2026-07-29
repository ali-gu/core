package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (api *API) HealthHandler(ctx *gin.Context) {
	if err := api.DBMux.PingReader(ctx); err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}
	if err := api.DBMux.PingWriter(ctx); err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}
	ctx.String(http.StatusOK, "success")
}
