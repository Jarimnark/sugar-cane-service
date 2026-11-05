package commentrepo

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/Jarimnark/sugar-cane-service/internal/store/sqlcgen"
)

type Repository interface {
	GameExists(ctx context.Context, gameID pgtype.UUID) (bool, error)
	Count(ctx context.Context, gameID pgtype.UUID) (int64, error)
	ListNewest(ctx context.Context, gameID, viewerID pgtype.UUID, limit, offset int32) ([]sqlcgen.ListCommentsForGameNewestRow, error)
	ListTop(ctx context.Context, gameID, viewerID pgtype.UUID, limit, offset int32) ([]sqlcgen.ListCommentsForGameTopRow, error)
	GetByID(ctx context.Context, id pgtype.UUID) (sqlcgen.GameComment, error)
	GetWithAuthor(ctx context.Context, id pgtype.UUID) (sqlcgen.GetCommentWithAuthorRow, error)
	Insert(ctx context.Context, gameID, userID pgtype.UUID, content string, anonymous bool) (sqlcgen.GameComment, error)
	Delete(ctx context.Context, id, userID pgtype.UUID) error

	Like(ctx context.Context, userID, commentID pgtype.UUID) error
	Unlike(ctx context.Context, userID, commentID pgtype.UUID) error
	IncrementLikeCount(ctx context.Context, commentID pgtype.UUID) error
	DecrementLikeCount(ctx context.Context, commentID pgtype.UUID) error
}
