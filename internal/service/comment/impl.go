package commentsvc

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Jarimnark/sugar-cane-service/internal/dbutil"
	"github.com/Jarimnark/sugar-cane-service/internal/domain"
	commentrepo "github.com/Jarimnark/sugar-cane-service/internal/repository/comment"
	commentdto "github.com/Jarimnark/sugar-cane-service/pkg/dto/comment"
)

const anonymousAuthorName = "Anonymous"

type svc struct {
	comments commentrepo.Repository
	pool     *pgxpool.Pool
}

func New(comments commentrepo.Repository, pool *pgxpool.Pool) Service {
	return &svc{comments: comments, pool: pool}
}

func parseUUID(s string) (pgtype.UUID, error) {
	var u pgtype.UUID
	if err := u.Scan(s); err != nil {
		return pgtype.UUID{}, err
	}
	return u, nil
}

func (s *svc) List(ctx context.Context, gameID, sort string, page, limit int, viewerID pgtype.UUID) (*commentdto.ListCommentsResponse, error) {
	gid, err := parseUUID(gameID)
	if err != nil {
		return nil, domain.BadRequest("invalid id", err)
	}

	switch sort {
	case "", commentdto.SortNewest:
		sort = commentdto.SortNewest
	case commentdto.SortTop:
		// ok
	default:
		return nil, domain.BadRequest("invalid sort", nil)
	}

	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * limit

	exists, err := s.comments.GameExists(ctx, gid)
	if err != nil {
		return nil, fmt.Errorf("List exists: %w", err)
	}
	if !exists {
		return nil, domain.NotFound("game not found")
	}

	count, err := s.comments.Count(ctx, gid)
	if err != nil {
		return nil, fmt.Errorf("List count: %w", err)
	}

	items := make([]commentdto.CommentResponse, 0, limit)

	if sort == commentdto.SortTop {
		rows, err := s.comments.ListTop(ctx, gid, viewerID, int32(limit), int32(offset))
		if err != nil {
			return nil, fmt.Errorf("List top: %w", err)
		}
		for _, row := range rows {
			items = append(items, mapTopRow(row, viewerID))
		}
	} else {
		rows, err := s.comments.ListNewest(ctx, gid, viewerID, int32(limit), int32(offset))
		if err != nil {
			return nil, fmt.Errorf("List newest: %w", err)
		}
		for _, row := range rows {
			items = append(items, mapNewestRow(row, viewerID))
		}
	}

	return &commentdto.ListCommentsResponse{
		Count: int(count),
		Items: items,
	}, nil
}

func (s *svc) Create(ctx context.Context, gameID string, authorID pgtype.UUID, req commentdto.CreateCommentRequest) (*commentdto.CommentResponse, error) {
	gid, err := parseUUID(gameID)
	if err != nil {
		return nil, domain.BadRequest("invalid id", err)
	}

	if !authorID.Valid {
		req.Anonymous = true
	}

	exists, err := s.comments.GameExists(ctx, gid)
	if err != nil {
		return nil, fmt.Errorf("Create exists: %w", err)
	}
	if !exists {
		return nil, domain.NotFound("game not found")
	}

	inserted, err := s.comments.Insert(ctx, gid, authorID, req.Content, req.Anonymous)
	if err != nil {
		return nil, fmt.Errorf("Create insert: %w", err)
	}

	row, err := s.comments.GetWithAuthor(ctx, inserted.ID)
	if err != nil {
		return nil, fmt.Errorf("Create fetch: %w", err)
	}

	resp := mapAuthorRow(row, authorID, false)
	return &resp, nil
}

func (s *svc) Delete(ctx context.Context, commentID string, authorID pgtype.UUID) error {
	cid, err := parseUUID(commentID)
	if err != nil {
		return domain.BadRequest("invalid id", err)
	}

	existing, err := s.comments.GetByID(ctx, cid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.NotFound("comment not found")
		}
		return fmt.Errorf("Delete lookup: %w", err)
	}
	if existing.UserID != authorID {
		return domain.Forbidden("not comment author")
	}

	if err := s.comments.Delete(ctx, cid, authorID); err != nil {
		return fmt.Errorf("Delete: %w", err)
	}
	return nil
}

func (s *svc) Like(ctx context.Context, commentID string, userID pgtype.UUID) error {
	cid, err := parseUUID(commentID)
	if err != nil {
		return domain.BadRequest("invalid id", err)
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("Like begin: %w", err)
	}
	defer tx.Rollback(ctx)

	txCtx := dbutil.WithTx(ctx, tx)
	if err := s.comments.Like(txCtx, userID, cid); err != nil {
		return fmt.Errorf("Like insert: %w", err)
	}
	if err := s.comments.IncrementLikeCount(txCtx, cid); err != nil {
		return fmt.Errorf("Like increment: %w", err)
	}
	return tx.Commit(ctx)
}

func (s *svc) Unlike(ctx context.Context, commentID string, userID pgtype.UUID) error {
	cid, err := parseUUID(commentID)
	if err != nil {
		return domain.BadRequest("invalid id", err)
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("Unlike begin: %w", err)
	}
	defer tx.Rollback(ctx)

	txCtx := dbutil.WithTx(ctx, tx)
	if err := s.comments.Unlike(txCtx, userID, cid); err != nil {
		return fmt.Errorf("Unlike delete: %w", err)
	}
	if err := s.comments.DecrementLikeCount(txCtx, cid); err != nil {
		return fmt.Errorf("Unlike decrement: %w", err)
	}
	return tx.Commit(ctx)
}
