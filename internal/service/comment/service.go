package commentsvc

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"

	commentdto "github.com/Jarimnark/sugar-cane-service/pkg/dto/comment"
)

type Service interface {
	List(ctx context.Context, gameID, sort string, page, limit int, viewerID pgtype.UUID) (*commentdto.ListCommentsResponse, error)
	Create(ctx context.Context, gameID string, authorID pgtype.UUID, req commentdto.CreateCommentRequest) (*commentdto.CommentResponse, error)
	Delete(ctx context.Context, commentID string, authorID pgtype.UUID) error
	Like(ctx context.Context, commentID string, userID pgtype.UUID) error
	Unlike(ctx context.Context, commentID string, userID pgtype.UUID) error
}
