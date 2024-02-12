package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"test-auth/internal/config"
	"test-auth/internal/handlers/post"
	"test-auth/internal/logger"
	"test-auth/internal/service/auth"
	"test-auth/internal/service/hasher"
	tokenmanager "test-auth/internal/service/token-manager"
	"test-auth/internal/storage"
	repomongo "test-auth/internal/storage/repo-mongo"

	"github.com/labstack/echo"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

type Server struct {
	e   *echo.Echo
	log *slog.Logger
	db  *repomongo.UsersDB
}

func New(cfg config.Config) *Server {
	e := echo.New()

	log := setupLogger(cfg.Env)

	tm, err := tokenmanager.NewManager(cfg.SigningKey)
	if err != nil {
		panic("can't create token manager")
	}

	db, err := repomongo.NewUsersDB(cfg.DBName, cfg.Collection, cfg.Host)
	if err != nil {
		panic("can't connect to db")
	}

	service := auth.NewServiceAuth(hasher.NewBcryptHasher(), db,
		tm, cfg.AccessTokenTTL, cfg.RefreshTokenTTL)

	e.POST("/auth", func(ctx echo.Context) error {
		res, err := post.Auth(ctx, *service)
		if errors.Is(err, storage.ErrNotFound) {
			log.Error("user not found", logger.Err(err))
			return ctx.String(http.StatusNotFound, err.Error())
		}
		if err != nil {
			log.Error("can't get tokens", logger.Err(err))
			return ctx.String(http.StatusInternalServerError, "internal error")
		}
		return ctx.JSON(http.StatusOK, res)
	})
	e.POST("/refresh", func(ctx echo.Context) error {
		res, err := post.UserRefresh(ctx, *service)
		if err != nil {
			if errors.Is(err, auth.ErrValidationError) {
				log.Error("validation error")
				return ctx.String(http.StatusBadRequest, "validation error")
			}
			log.Error("can't refresh tokens", logger.Err(err))
			return ctx.String(http.StatusInternalServerError, err.Error())
		}
		return ctx.JSON(http.StatusOK, res)
	})

	return &Server{
		e:   e,
		log: log,
		db:  db,
	}
}

func (s *Server) Run(port int) error {
	defer s.db.Close(context.Background())
	s.log.Info("start server")
	err := s.e.Start(fmt.Sprintf(":%d", port))

	if err != nil {
		return err
	}
	return err
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(
			slog.NewTextHandler(
				os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(
				os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(
				os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}
	return log
}
