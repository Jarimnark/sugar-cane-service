package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Jarimnark/sugar-cane-service/internal/config"
	"github.com/Jarimnark/sugar-cane-service/internal/handler"
	authhandler "github.com/Jarimnark/sugar-cane-service/internal/handler/auth"
	commenthandler "github.com/Jarimnark/sugar-cane-service/internal/handler/comment"
	devhandler "github.com/Jarimnark/sugar-cane-service/internal/handler/dev"
	gamehandler "github.com/Jarimnark/sugar-cane-service/internal/handler/game"
	listhandler "github.com/Jarimnark/sugar-cane-service/internal/handler/list"
	settinghandler "github.com/Jarimnark/sugar-cane-service/internal/handler/setting"
	commentrepo "github.com/Jarimnark/sugar-cane-service/internal/repository/comment"
	gamerepo "github.com/Jarimnark/sugar-cane-service/internal/repository/game"
	likerepo "github.com/Jarimnark/sugar-cane-service/internal/repository/like"
	oauthrepo "github.com/Jarimnark/sugar-cane-service/internal/repository/oauth"
	savedgamerepo "github.com/Jarimnark/sugar-cane-service/internal/repository/savedgame"
	sessionrepo "github.com/Jarimnark/sugar-cane-service/internal/repository/session"
	userrepo "github.com/Jarimnark/sugar-cane-service/internal/repository/user"
	authsvc "github.com/Jarimnark/sugar-cane-service/internal/service/auth"
	commentsvc "github.com/Jarimnark/sugar-cane-service/internal/service/comment"
	usersvc "github.com/Jarimnark/sugar-cane-service/internal/service/user"
	devsvc "github.com/Jarimnark/sugar-cane-service/internal/service/dev"
	gamesvc "github.com/Jarimnark/sugar-cane-service/internal/service/game"
	listsvc "github.com/Jarimnark/sugar-cane-service/internal/service/list"
	settingsvc "github.com/Jarimnark/sugar-cane-service/internal/service/setting"
	"github.com/Jarimnark/sugar-cane-service/internal/storage"
	"github.com/Jarimnark/sugar-cane-service/internal/store/sqlcgen"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	cfg, err := config.Load()
	if err != nil {
		slog.Error("config", "error", err)
		os.Exit(1)
	}

	pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		slog.Error("db connect", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := pool.Ping(context.Background()); err != nil {
		slog.Error("db ping", "error", err)
		os.Exit(1)
	}
	slog.Info("database connected")

	storageClient, err := storage.New(context.Background(), cfg)
	if err != nil {
		// Non-fatal: upload handlers null-check h.storage and return 503.
		// Boot the rest of the app so the public site stays reachable even when
		// the S3 endpoint is misconfigured or temporarily unreachable.
		slog.Warn("storage init failed; uploads will be disabled",
			"error", err, "endpoint", cfg.S3Endpoint, "bucket", cfg.S3Bucket)
		storageClient = nil
	} else {
		slog.Info("storage connected", "bucket", cfg.S3Bucket)
	}

	q := sqlcgen.New(pool)

	sessions := sessionrepo.New(q)
	games := gamerepo.New(q)
	users := userrepo.New(q)
	oauths := oauthrepo.New(q)
	likes := likerepo.New(q)
	saved := savedgamerepo.New(q)
	comments := commentrepo.New(q)

	userService := usersvc.New(users, sessions, oauths, pool)
	gameService := gamesvc.New(games, sessions, q)
	authService := authsvc.New(userService, cfg)
	settingService := settingsvc.New(users, oauths, sessions, pool)
	listService := listsvc.New(likes, saved, pool)
	devService := devsvc.New(games, users)
	commentService := commentsvc.New(comments, pool)

	r := handler.NewRouter(
		authhandler.New(authService, cfg),
		settinghandler.New(settingService),
		gamehandler.New(gameService),
		listhandler.New(listService),
		devhandler.New(devService, storageClient),
		commenthandler.New(commentService),
		sessions,
	)

	addr := fmt.Sprintf(":%s", cfg.Port)
	slog.Info("server starting", "addr", addr)

	if err := http.ListenAndServe(addr, r); err != nil {
		slog.Error("server", "error", err)
		os.Exit(1)
	}
}
