package comment

import (
	"time"

	"brokle/pkg/ulid"
)

type Reaction struct {
	ID        ulid.ULID `json:"id" gorm:"type:char(26);primaryKey"`
	CommentID ulid.ULID `json:"comment_id" gorm:"type:char(26);not null"`
	UserID    ulid.ULID `json:"user_id" gorm:"type:char(26);not null"`
	Emoji     string    `json:"emoji" gorm:"type:varchar(8);not null"`
	CreatedAt time.Time `json:"created_at" gorm:"not null;autoCreateTime"`
}

func (Reaction) TableName() string {
	return "comment_reactions"
}

func NewReaction(commentID, userID ulid.ULID, emoji string) *Reaction {
	return &Reaction{
		ID:        ulid.New(),
		CommentID: commentID,
		UserID:    userID,
		Emoji:     emoji,
		CreatedAt: time.Now(),
	}
}

type ReactionSummary struct {
	Emoji   string   `json:"emoji"`
	Count   int      `json:"count"`
	Users   []string `json:"users"`   // User names who reacted with this emoji
	HasUser bool     `json:"has_user"` // Whether the current user has reacted with this emoji
}

type ToggleReactionRequest struct {
	Emoji string `json:"emoji" binding:"required,max=8"`
}

const (
	MaxEmojisPerComment = 6 // Maximum different emoji types allowed per comment
)
