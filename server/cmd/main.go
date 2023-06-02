package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Wave-95/boards/server/db"
	"github.com/Wave-95/boards/server/internal/api/auth"
	"github.com/Wave-95/boards/server/internal/api/board"
	"github.com/Wave-95/boards/server/internal/api/user"
	"github.com/Wave-95/boards/server/internal/api/ws"
	"github.com/Wave-95/boards/server/internal/config"
	"github.com/Wave-95/boards/server/internal/jwt"
	"github.com/Wave-95/boards/server/internal/middleware"
	"github.com/Wave-95/boards/server/pkg/logger"
	"github.com/Wave-95/boards/server/pkg/validator"
	"github.com/go-chi/chi/v5"
)

func main() {
	logger := logger.New()
	validator := validator.New()
	// load env vars into config
	cfg, err := config.Load()
	if err != nil {
		logger.Fatalf("Error loading config: %v", err)
	}
	// connect to db
	conn, err := db.Connect(cfg.DatabaseConfig)
	if err != nil {
		logger.Fatalf("Error connecting to db: %v", err)
	}
	defer conn.Close()

	// setup server
	r := chi.NewRouter()
	server := http.Server{Addr: ":8080", Handler: buildHandler(r, conn, logger, validator, cfg)}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Could not start server: %s", err)
		}
	}()

	<-stop
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced shutdown: %s", err)
	}
}

func buildHandler(r chi.Router, db *db.DB, logger logger.Logger, v validator.Validate, cfg *config.Config) chi.Router {
	// set up middleware
	r.Use(middleware.RequestLogger(logger))
	r.Use(middleware.Cors())

	// set up repositories
	userRepo := user.NewRepository(db)
	boardRepo := board.NewRepository(db)

	// set up services
	jwtService := jwt.New(cfg.JwtSecret, cfg.JwtExpiration)
	authService := auth.NewService(userRepo, jwtService)
	userService := user.NewService(userRepo, v)
	boardService := board.NewService(boardRepo, v)

	// set up APIs
	userAPI := user.NewAPI(userService, jwtService, v)
	authAPI := auth.NewAPI(authService, v)
	boardAPI := board.NewAPI(boardService, v)
	wsAPI := ws.NewAPI(jwtService)

	// set up auth handler
	authHandler := middleware.Auth(jwtService)

	// register handlers
	userAPI.RegisterHandlers(r, authHandler)
	authAPI.RegisterHandlers(r)
	boardAPI.RegisterHandlers(r, authHandler)
	wsAPI.RegisterHandlers(r)

	return r
}
