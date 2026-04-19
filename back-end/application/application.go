package application

import (
	"chat-app/back-end/internal"
	"chat-app/back-end/internal/config"
	"chat-app/back-end/internal/util"
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

type App struct {
	router http.Handler
	rdb    *redis.Client
	db     *sqlx.DB
}

func New() *App {
	return &App{}
}

func (a *App) Start(ctx context.Context) error {
	db, err := config.DB(os.Getenv("DATABASE_URL"))
	if err != nil {
		return fmt.Errorf("failed to connect to postgres: %w", err)
	}
	a.db = db

	rdb, err := config.Redis(os.Getenv("REDIS_ADDR"), os.Getenv("REDIS_PASSWORD"))
	if err != nil {
		return err
	}
	a.rdb = rdb

	jwtManager := util.NewJWTManager(
		os.Getenv("JWT_SECRET"),
		15*time.Minute,
		7*24*time.Hour,
	)

	a.router = internal.NewRouter(db, rdb, jwtManager)

	server := &http.Server{
		Addr:    ":8080",
		Handler: a.router,
	}

	defer func() {
		if err := a.db.Close(); err != nil {
			fmt.Println("failed to close postgres", err)
		}
	}()

	defer func() {
		if err := a.rdb.Close(); err != nil {
			fmt.Println("failed to close redis", err)
		}
	}()

	fmt.Println("Starting server on :8080")

	ch := make(chan error, 1)

	go func() {
		err = server.ListenAndServe()
		if err != nil {
			ch <- fmt.Errorf("failed to start server: %w", err)
		}
		close(ch)
	}()

	select {
	case err := <-ch:
		return err
	case <-ctx.Done():
		timeout, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		return server.Shutdown(timeout)
	}
}
