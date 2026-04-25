package comment

// commentBody — shared shape for create + update + create-reply
// (all three accept the same field).
type commentBody struct {
	Content string `json:"content" validate:"required,min=1,max=10000"`
}

// toggleReactionBody — POST /comments/{id}/reactions
type toggleReactionBody struct {
	Emoji string `json:"emoji" validate:"required,max=8"`
}
