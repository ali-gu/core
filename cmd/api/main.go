package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ali-gulzar/speechory-core/internal"
	internalAPI "github.com/ali-gulzar/speechory-core/internal/api"
	"github.com/ali-gulzar/speechory-core/internal/biz"
	"github.com/ali-gulzar/speechory-core/internal/constants"
	agentTelnyx "github.com/ali-gulzar/speechory-core/internal/services/agent/telnyx"
	supabaseAuth "github.com/ali-gulzar/speechory-core/internal/services/auth/supabase"
	"github.com/ali-gulzar/speechory-core/internal/services/ehr/nexhealth"
	phoneNumberTelnyx "github.com/ali-gulzar/speechory-core/internal/services/phonenumber/telnyx"
	supabaseService "github.com/ali-gulzar/speechory-core/internal/services/supabase"
	"github.com/ali-gulzar/speechory-core/internal/services/telnyx"
	toolTelnyx "github.com/ali-gulzar/speechory-core/internal/services/tool/telnyx"
	"github.com/gin-gonic/gin"

	"github.com/ali-gulzar/speechory-core/internal/storage"
	"github.com/ali-gulzar/speechory-core/pkg/rlog"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	appCfg, err := internal.NewConfig(constants.GetLandscapeFromEnv())
	if err != nil {
		log.Fatal(err)
	}
	if appCfg.Landscape != constants.LandscapeLocal {
		gin.SetMode(gin.ReleaseMode)
	}

	ctx = rlog.Initialize(ctx)
	appLogger := rlog.GetLogger(ctx)

	dbMux, err := storage.NewDBMux(ctx, *appCfg)
	if err != nil {
		log.Fatal(err)
	}
	defer dbMux.Close()

	// storage layer
	storageDeps := storage.NewStorage()

	// services
	telnyxSvc := telnyx.New(appCfg.Telnyx)
	nexhealthClient := nexhealth.New(appCfg.NexHealth, dbMux, storageDeps.NexHealthAuth)
	supabaseSvc := supabaseService.New(appCfg.Supabase)

	// biz layer
	globalBiz := biz.NewBiz(biz.Dependencies{
		Storage:                  storageDeps,
		TelnyxAgent:              agentTelnyx.NewTelnyx(telnyxSvc.Client, appCfg.Landscape),
		TelnyxPhoneNumberManager: phoneNumberTelnyx.NewTelnyx(telnyxSvc.Client),
		SupabaseAuth:             supabaseAuth.NewSupabase(supabaseSvc.Client),
		EHR:                      nexhealthClient,
		TelnyxTool:               toolTelnyx.NewTelnyx(telnyxSvc.Client),
		Domain:                   appCfg.Domain,
	})

	// api layer
	api := internalAPI.API{
		Biz:     globalBiz,
		Config:  *appCfg,
		DBMux:   dbMux,
		Logger:  appLogger,
		Storage: storageDeps,
	}
	router := api.Routes()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	var listenAddress string
	if appCfg.Landscape == constants.LandscapeLocal {
		listenAddress = "0.0.0.0:" + port
	} else {
		listenAddress = ":" + port
	}

	srv := &http.Server{
		Addr:    listenAddress,
		Handler: router,
	}

	go func() {
		serErr := srv.ListenAndServe()
		if serErr != nil {
			log.Fatal(serErr)
		}
	}()

	<-ctx.Done()
	stop()

	appLogger.Info("shutting down gracefully")
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	if err = srv.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}
	appLogger.Info("server shutdown complete")
}
