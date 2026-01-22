package comment

import "errors"

var (
	ErrNotFound           = errors.New("comment not found")
	ErrNotOwner           = errors.New("not the comment owner")
	ErrMaxEmojisExceeded  = errors.New("maximum emoji types per comment exceeded")
	ErrCannotReplyToReply = errors.New("cannot reply to a reply")
)
