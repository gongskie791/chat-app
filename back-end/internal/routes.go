package internal

import (
	"chat-app/back-end/internal/handler"
	"chat-app/back-end/internal/hub"
	"chat-app/back-end/internal/middleware"
	"chat-app/back-end/internal/repository"
	"chat-app/back-end/internal/service"
	"chat-app/back-end/internal/util"
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

func NewRouter(db *sqlx.DB, rdb *redis.Client, jwtManager *util.JWTManager) http.Handler {
	router := http.NewServeMux()

	h := hub.NewHub()
	go h.Run()

	router.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

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
