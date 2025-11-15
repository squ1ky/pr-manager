package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/squ1ky/pr-manager/config"
	v1 "github.com/squ1ky/pr-manager/internal/handler/http/v1"
	"github.com/squ1ky/pr-manager/internal/infrastructure/db"
	"github.com/squ1ky/pr-manager/internal/repository/pgrepo"
	"github.com/squ1ky/pr-manager/internal/service"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		logger.Error("failed to load config", slog.String("error", err.Error()))
		return
	}

	dbConn, err := db.NewPostgresConnection(&cfg.Database, logger)
	if err != nil {
		logger.Error("failed to connect database", slog.String("error", err.Error()))
		return
	}
	defer func() {
		if err := dbConn.Close(); err != nil {
			logger.Error("failed to close database", slog.String("error", err.Error()))
		}
	}()

	teamRepo := pgrepo.NewTeamRepository(dbConn.DB)
	userRepo := pgrepo.NewUserRepository(dbConn.DB)
	prRepo := pgrepo.NewPullRequestRepository(dbConn.DB)
	prReviewerRepo := pgrepo.NewPullRequestReviewerRepository(dbConn.DB)

	teamSvc := service.NewTeamService(teamRepo, userRepo)
	userSvc := service.NewUserService(userRepo)
	prSvc := service.NewPullRequestService(prRepo, prReviewerRepo, userRepo)

	gin.SetMode(gin.DebugMode)
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	v1.NewRouter(router, teamSvc, userSvc, prSvc)

	addr := fmt.Sprintf("%:d", cfg.App.Port)

	srv := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	go func() {
		logger.Info("starting HTTP server", slog.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server exited with error", slog.String("error", err.Error()))
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("server shutdown error", slog.String("error", err.Error()))
	} else {
		logger.Info("server gracefully stopped")
	}
}
