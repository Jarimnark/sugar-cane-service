package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/Jarimnark/sugar-cane-service/internal/config"
	"github.com/Jarimnark/sugar-cane-service/internal/mcp"
	"github.com/Jarimnark/sugar-cane-service/internal/mcp/hopbu"
	devsvc "github.com/Jarimnark/sugar-cane-service/internal/service/dev"
	gamesvc "github.com/Jarimnark/sugar-cane-service/internal/service/game"
	"github.com/Jarimnark/sugar-cane-service/internal/storage"
	gamedto "github.com/Jarimnark/sugar-cane-service/pkg/dto/game"
)

// mcpGames binds the dev/game services to the configured MCP owner account.
// The user ID comes from server config (MCP_USER_ID), never from the client.
type mcpGames struct {
	dev    devsvc.Service
	games  gamesvc.Service
	userID pgtype.UUID
}

func (m mcpGames) Options(ctx context.Context) (*gamedto.MasterOptionResponse, error) {
	return m.games.MasterOption(ctx)
}

func (m mcpGames) ListMine(ctx context.Context) ([]gamedto.GameResponse, error) {
	return m.dev.ListMyGames(ctx, m.userID)
}

func (m mcpGames) GetMine(ctx context.Context, id string) (*gamedto.GameResponse, error) {
	return m.dev.GetMyGame(ctx, m.userID, id)
}

func (m mcpGames) Save(ctx context.Context, req gamedto.CreateOrUpdateGameRequest, publish bool) (*gamedto.GameResponse, error) {
	return m.dev.CreateOrUpdateGame(ctx, m.userID, req, publish)
}

// newMCPHandler returns the /mcp handler, or nil when MCP_TOKEN is not set.
func newMCPHandler(cfg *config.Config, dev devsvc.Service, games gamesvc.Service, store *storage.Client) (http.Handler, error) {
	if cfg.MCPToken == "" {
		return nil, nil
	}

	var userID pgtype.UUID
	if err := userID.Scan(cfg.MCPUserID); err != nil {
		return nil, fmt.Errorf("MCP_USER_ID is not a valid UUID: %w", err)
	}

	deps := hopbu.Deps{
		Games:    mcpGames{dev: dev, games: games, userID: userID},
		Fetcher:  hopbu.NewImageFetcher(),
		Validate: validator.New().Struct,
	}
	// Assign only when non-nil so a missing client stays a nil interface.
	if store != nil {
		deps.Media = store
	}

	srv, err := mcp.NewServer("hopbu", "1.0.0", hopbu.Instructions, cfg.MCPToken, hopbu.Tools(deps))
	if err != nil {
		return nil, err
	}
	return srv, nil
}
