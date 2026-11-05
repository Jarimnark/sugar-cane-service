package commentsvc

import (
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/Jarimnark/sugar-cane-service/internal/store/sqlcgen"
	commentdto "github.com/Jarimnark/sugar-cane-service/pkg/dto/comment"
)

func mapNewestRow(row sqlcgen.ListCommentsForGameNewestRow, viewerID pgtype.UUID) commentdto.CommentResponse {
	return buildResponse(responseInput{
		id:             row.ID,
		gameID:         row.GameID,
		userID:         row.UserID,
		content:        row.Content,
		anonymous:      row.Anonymous,
		likeCount:      row.LikeCount,
		createdAt:      row.CreatedAt,
		authorName:     row.AuthorName,
		authorPhotoURL: row.AuthorPhotoUrl,
		gameOwnerID:    row.GameOwnerID,
		liked:          row.Liked,
	}, viewerID)
}

func mapTopRow(row sqlcgen.ListCommentsForGameTopRow, viewerID pgtype.UUID) commentdto.CommentResponse {
	return buildResponse(responseInput{
		id:             row.ID,
		gameID:         row.GameID,
		userID:         row.UserID,
		content:        row.Content,
		anonymous:      row.Anonymous,
		likeCount:      row.LikeCount,
		createdAt:      row.CreatedAt,
		authorName:     row.AuthorName,
		authorPhotoURL: row.AuthorPhotoUrl,
		gameOwnerID:    row.GameOwnerID,
		liked:          row.Liked,
	}, viewerID)
}

func mapAuthorRow(row sqlcgen.GetCommentWithAuthorRow, viewerID pgtype.UUID, liked bool) commentdto.CommentResponse {
	return buildResponse(responseInput{
		id:             row.ID,
		gameID:         row.GameID,
		userID:         row.UserID,
		content:        row.Content,
		anonymous:      row.Anonymous,
		likeCount:      row.LikeCount,
		createdAt:      row.CreatedAt,
		authorName:     row.AuthorName,
		authorPhotoURL: row.AuthorPhotoUrl,
		gameOwnerID:    row.GameOwnerID,
		liked:          liked,
	}, viewerID)
}

type responseInput struct {
	id             pgtype.UUID
	gameID         pgtype.UUID
	userID         pgtype.UUID
	content        string
	anonymous      bool
	likeCount      int32
	createdAt      pgtype.Timestamptz
	authorName     string
	authorPhotoURL pgtype.Text
	gameOwnerID    pgtype.UUID
	liked          bool
}

func buildResponse(in responseInput, viewerID pgtype.UUID) commentdto.CommentResponse {
	resp := commentdto.CommentResponse{
		ID:        in.id.String(),
		GameID:    in.gameID.String(),
		Anonymous: in.anonymous,
		Content:   in.content,
		LikeCount: int(in.likeCount),
		Liked:     in.liked && viewerID.Valid,
		CreatedAt: in.createdAt.Time,
		IsOwn:     viewerID.Valid && in.userID == viewerID,
	}
	if in.anonymous {
		resp.UserID = ""
		resp.AuthorName = anonymousAuthorName
		resp.AuthorPhotoURL = ""
		resp.IsDeveloper = false
	} else {
		resp.UserID = in.userID.String()
		resp.AuthorName = in.authorName
		if in.authorPhotoURL.Valid {
			resp.AuthorPhotoURL = in.authorPhotoURL.String
		}
		resp.IsDeveloper = in.userID == in.gameOwnerID
	}
	return resp
}
