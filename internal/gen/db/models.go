// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.15.0

package db

import (
	"time"
)

type Post struct {
	UserID    string
	PostID    string
	Data      string
	CreatedAt time.Time
	UpdatedAt time.Time
}