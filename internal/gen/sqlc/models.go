// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.22.0

package sqlc

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type TodoappTodo struct {
	ID        int64
	UserID    string
	TodoID    string
	Todo      string
	CreatedAt pgtype.Timestamptz
	UpdatedAt pgtype.Timestamptz
}
