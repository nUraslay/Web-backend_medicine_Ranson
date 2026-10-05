package handler

import "sync"

var (
	currentUserOnce sync.Once
	currentUser     uint
)

func CurrentUserID() uint {
	currentUserOnce.Do(func() {
		currentUser = currentUserID
	})

	return currentUser
}
