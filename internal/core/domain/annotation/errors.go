package annotation

import "errors"

var (
	// Queue errors
	ErrQueueNotFound = errors.New("annotation queue not found")
	ErrQueueExists   = errors.New("annotation queue with this name already exists")

	// Item errors
	ErrItemNotFound     = errors.New("queue item not found")
	ErrItemExists       = errors.New("item already exists in queue")
	ErrNoItemsAvailable = errors.New("no items available for annotation")

	// Assignment errors
	ErrAssignmentNotFound = errors.New("queue assignment not found")
	ErrAssignmentExists   = errors.New("user is already assigned to this queue")
	ErrInsufficientRole   = errors.New("insufficient role for this operation")

)
