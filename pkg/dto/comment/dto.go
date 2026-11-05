package commentdto

import "time"

const (
	SortNewest = "newest"
	SortTop    = "top"
)

type CreateCommentRequest struct {
	Content   string `json:"content"   validate:"required,min=1,max=500"`
	Anonymous bool   `json:"anonymous"`
}

type CommentResponse struct {
	ID             string    `json:"id"`
	GameID         string    `json:"game_id"`
	UserID         string    `json:"user_id"`
	AuthorName     string    `json:"author_name"`
	AuthorPhotoURL string    `json:"author_photo_url"`
	IsDeveloper    bool      `json:"is_developer"`
	IsOwn          bool      `json:"is_own"`
	Anonymous      bool      `json:"anonymous"`
	Content        string    `json:"content"`
	LikeCount      int       `json:"like_count"`
	Liked          bool      `json:"liked"`
	CreatedAt      time.Time `json:"created_at"`
}

type ListCommentsResponse struct {
	Count int               `json:"count"`
	Items []CommentResponse `json:"items"`
}
