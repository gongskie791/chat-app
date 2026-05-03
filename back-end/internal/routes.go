package internal

import (
	"chat-app/back-end/internal/handler"
	"chat-app/back-end/internal/hub"
	"chat-app/back-end/internal/middleware"
	"chat-app/back-end/internal/repository"
	"chat-app/back-end/internal/service"
	"chat-app/back-end/internal/util"
	"context"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
)

func NewRouter(db *sqlx.DB, rdb *redis.Client, jwtManager *util.JWTManager) http.Handler {
	router := http.NewServeMux()

	h := hub.NewHub()
	go h.Run()

	router.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		type healthStatus struct {
			Status   string `json:"status"`
			Postgres string `json:"postgres"`
			Redis    string `json:"redis"`
		}

		result := healthStatus{Status: "ok", Postgres: "ok", Redis: "ok"}

		if err := db.PingContext(ctx); err != nil {
			result.Status = "unhealthy"
			result.Postgres = err.Error()
		}

		if err := rdb.Ping(ctx).Err(); err != nil {
			result.Status = "unhealthy"
			result.Redis = err.Error()
		}

		statusCode := http.StatusOK
		if result.Status == "unhealthy" {
			statusCode = http.StatusServiceUnavailable
		}

		util.JSON(w, statusCode, result)
	})
	router.Handle("GET /metrics", promhttp.Handler())

	loadAuthRoutes(router, db, rdb, jwtManager)
	loadRoomRoutes(router, db, rdb, jwtManager, h)

	return middleware.Logger(middleware.Cors(router))
}

func loadAuthRoutes(router *http.ServeMux, db *sqlx.DB, rdb *redis.Client, jwtManager *util.JWTManager) {
	userRepo := repository.NewUserRepository(db)

	userService := service.NewUserService(userRepo, jwtManager, rdb)
	authService := service.NewAuthService(rdb, jwtManager)

	userHandler := handler.NewUserHandler(userService)
	authHandler := handler.NewAuthHandler(authService)

	auth := middleware.VerifyJWT(jwtManager)

	// public
	router.HandleFunc("POST /api/auth/register", userHandler.Register)
	router.HandleFunc("POST /api/auth/login", userHandler.Login)
	router.HandleFunc("POST /api/auth/refresh", authHandler.RefreshToken)

	// protected
	router.Handle("POST /api/auth/logout", auth(http.HandlerFunc(userHandler.Logout)))
}

func loadRoomRoutes(router *http.ServeMux, db *sqlx.DB, rdb *redis.Client, jwtManager *util.JWTManager, h *hub.Hub) {
	userRepo := repository.NewUserRepository(db)
	roomRepo := repository.NewRoomRepository(db)
	msgRepo := repository.NewMessageRepository(rdb)

	roomService := service.NewRoomService(roomRepo)

	roomHandler := handler.NewRoomHandler(roomService)
	wsHandler := handler.NewWSHandler(h, userRepo, msgRepo)

	auth := middleware.VerifyJWT(jwtManager)

	// protected
	router.Handle("POST /api/rooms", auth(http.HandlerFunc(roomHandler.CreateRoom)))
	router.Handle("GET /api/rooms", auth(http.HandlerFunc(roomHandler.GetRooms)))
	router.Handle("GET /api/ws/{roomID}", auth(http.HandlerFunc(wsHandler.ServeWS)))
}
