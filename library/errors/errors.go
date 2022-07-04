package errors

import (
	"errors"
)

var (
	ErrInstanceHasExists = errors.New("instance has exists")
	ErrRaftConfigIsEmpty = errors.New("raft config raft_addr is empty")
)
