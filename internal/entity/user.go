package entity

import "time"

type User struct {
	ID        string    `json:"user_id" db:"id"`
	Username  string    `json:"username" db:"username"`
	TeamID    *string   `json:"-" db:"team_id"`
	IsActive  bool      `json:"is_active" db:"is_active"`
	CreatedAt time.Time `json:"-" db:"created_at"`
}
